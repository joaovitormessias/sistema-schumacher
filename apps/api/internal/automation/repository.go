package automation

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	RecordEvolutionStatus(ctx context.Context, input RecordEvolutionStatusInput) (RecordEvolutionStatusResult, error)
	CreateJobRun(ctx context.Context, input CreateJobRunInput) (JobRun, error)
	UpdateJobRun(ctx context.Context, input UpdateJobRunInput) (JobRun, error)
	FindLatestJobRunByKey(ctx context.Context, jobName string, alertSignature string) (*JobRun, error)
	ListJobRuns(ctx context.Context, input ListJobRunsInput) ([]JobRun, error)
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) RecordEvolutionStatus(ctx context.Context, input RecordEvolutionStatusInput) (RecordEvolutionStatusResult, error) {
	result := RecordEvolutionStatusResult{}

	payload, err := encodePayload(input.Payload)
	if err != nil {
		return RecordEvolutionStatusResult{}, err
	}
	messageTag, err := encodePayload(map[string]interface{}{
		"latest_provider_status": input.ProviderStatus,
		"status_event":           input.Event,
		"status_observed_at":     input.ObservedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
	if err != nil {
		return RecordEvolutionStatusResult{}, err
	}

	res, err := r.pool.Exec(ctx, `
		update chat_messages
		set processing_status = $2,
				normalized_payload = coalesce(normalized_payload, '{}'::jsonb) || $3::jsonb
		where provider_message_id = $1
	`, input.ProviderMessageID, input.ProviderStatus, messageTag)
	if err != nil {
		return RecordEvolutionStatusResult{}, err
	}
	result.MatchedChatMessages = res.RowsAffected()

	res, err = r.pool.Exec(ctx, `
		update outbound_messages
		set status = $2,
				updated_at = now(),
				sent_at = case
					when $2 in ('SERVER_ACK', 'SENT', 'DELIVERY_ACK', 'DELIVERED', 'READ') then coalesce(sent_at, $3)
					else sent_at
				end,
				delivered_at = case
					when $2 in ('DELIVERY_ACK', 'DELIVERED', 'READ') then coalesce(delivered_at, $3)
					else delivered_at
				end,
				payload = coalesce(payload, '{}'::jsonb) || $4::jsonb
		where provider_message_id = $1
	`, input.ProviderMessageID, input.ProviderStatus, input.ObservedAt.UTC(), payload)
	if err != nil {
		return RecordEvolutionStatusResult{}, err
	}
	result.MatchedOutboundMessages = res.RowsAffected()

	return result, nil
}

func (r *Repository) CreateJobRun(ctx context.Context, input CreateJobRunInput) (JobRun, error) {
	inputPayload, err := encodePayload(input.InputPayload)
	if err != nil {
		return JobRun{}, err
	}
	resultPayload, err := encodePayload(input.ResultPayload)
	if err != nil {
		return JobRun{}, err
	}
	row := r.pool.QueryRow(ctx, `
		insert into automation_job_runs (
			job_name,
			trigger_source,
			status,
			input_payload,
			result_payload,
			error_text,
			started_at,
			finished_at
		)
		values ($1, $2, $3, $4, $5, nullif($6, ''), $7, $8)
		returning
			id::text,
			job_name,
			trigger_source,
			coalesce(requested_by_user_id::text, ''),
			status,
			input_payload,
			result_payload,
			coalesce(error_text, ''),
			started_at,
			finished_at,
			created_at
	`, input.JobName, input.TriggerSource, input.Status, inputPayload, resultPayload, input.ErrorText, input.StartedAt.UTC(), input.FinishedAt)
	return scanJobRun(row)
}

func (r *Repository) UpdateJobRun(ctx context.Context, input UpdateJobRunInput) (JobRun, error) {
	resultPayload, err := encodePayload(input.ResultPayload)
	if err != nil {
		return JobRun{}, err
	}
	row := r.pool.QueryRow(ctx, `
		update automation_job_runs
		set status = $2,
				result_payload = $3,
				error_text = nullif($4, ''),
				finished_at = $5
		where id = $1::uuid
		returning
			id::text,
			job_name,
			trigger_source,
			coalesce(requested_by_user_id::text, ''),
			status,
			input_payload,
			result_payload,
			coalesce(error_text, ''),
			started_at,
			finished_at,
			created_at
	`, input.ID, input.Status, resultPayload, input.ErrorText, input.FinishedAt.UTC())
	return scanJobRun(row)
}

func (r *Repository) FindLatestJobRunByKey(ctx context.Context, jobName string, alertSignature string) (*JobRun, error) {
	row := r.pool.QueryRow(ctx, `
		select
			id::text,
			job_name,
			trigger_source,
			coalesce(requested_by_user_id::text, ''),
			status,
			input_payload,
			result_payload,
			coalesce(error_text, ''),
			started_at,
			finished_at,
			created_at
		from automation_job_runs
		where job_name = $1
			and status = 'SENT'
			and coalesce(result_payload->>'alert_signature', '') = $2
		order by started_at desc, created_at desc
		limit 1
	`, jobName, alertSignature)
	item, err := scanJobRun(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListJobRuns(ctx context.Context, input ListJobRunsInput) ([]JobRun, error) {
	rows, err := r.pool.Query(ctx, `
		select
			id::text,
			job_name,
			trigger_source,
			coalesce(requested_by_user_id::text, ''),
			status,
			input_payload,
			result_payload,
			coalesce(error_text, ''),
			started_at,
			finished_at,
			created_at
		from automation_job_runs
		where job_name = $1
			and ($2 = '' or status = $2)
			and ($3 = '' or trigger_source = $3)
		order by started_at desc, created_at desc
		limit $4
	`, input.JobName, input.Status, input.TriggerSource, input.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]JobRun, 0, input.Limit)
	for rows.Next() {
		item, err := scanJobRun(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func encodePayload(input map[string]interface{}) (string, error) {
	if len(input) == 0 {
		return "{}", nil
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

type jobRunScanner interface {
	Scan(dest ...interface{}) error
}

func scanJobRun(row jobRunScanner) (JobRun, error) {
	var item JobRun
	var inputPayload []byte
	var resultPayload []byte
	var requestedBy string
	var errorText string
	if err := row.Scan(
		&item.ID,
		&item.JobName,
		&item.TriggerSource,
		&requestedBy,
		&item.Status,
		&inputPayload,
		&resultPayload,
		&errorText,
		&item.StartedAt,
		&item.FinishedAt,
		&item.CreatedAt,
	); err != nil {
		return JobRun{}, err
	}
	item.RequestedByUserID = requestedBy
	item.ErrorText = errorText
	if len(inputPayload) > 0 {
		if err := json.Unmarshal(inputPayload, &item.InputPayload); err != nil {
			return JobRun{}, err
		}
	}
	if len(resultPayload) > 0 {
		if err := json.Unmarshal(resultPayload, &item.ResultPayload); err != nil {
			return JobRun{}, err
		}
	}
	if item.InputPayload == nil {
		item.InputPayload = map[string]interface{}{}
	}
	if item.ResultPayload == nil {
		item.ResultPayload = map[string]interface{}{}
	}
	item.StartedAt = item.StartedAt.UTC()
	if item.FinishedAt != nil {
		finished := item.FinishedAt.UTC()
		item.FinishedAt = &finished
	}
	item.CreatedAt = item.CreatedAt.UTC()
	return item, nil
}
