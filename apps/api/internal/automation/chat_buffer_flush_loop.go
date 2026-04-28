package automation

import (
	"context"
	"time"

	"schumacher-tur/api/internal/shared/config"
)

type chatBufferFlushLogger interface {
	Printf(format string, v ...interface{})
}

func StartChatBufferFlushLoop(ctx context.Context, svc *Service, cfg config.Config, logger chatBufferFlushLogger) {
	if svc == nil || !cfg.ChatBufferAutoFlushEnabled {
		return
	}

	interval := time.Duration(cfg.ChatBufferAutoFlushIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}

	run := func() {
		result, err := svc.runChatBufferFlush(ctx, RunChatBufferFlushInput{
			Limit:         cfg.ChatBufferAutoFlushLimit,
			Channel:       "WHATSAPP",
			Status:        "ACTIVE",
			HandoffStatus: "BOT",
		}, "SYSTEM")
		if logger == nil {
			return
		}
		if err != nil {
			logger.Printf("automation chat-buffer-flush failed: %v", err)
			return
		}
		logChatBufferFlushSessionResults(logger, result)
		logger.Printf(
			"automation chat-buffer-flush status=%s reason=%s checked=%d due=%d flushed=%d failed=%d",
			result.Status,
			result.Reason,
			result.CheckedCount,
			result.DueCount,
			result.FlushedCount,
			result.FailedCount,
		)
	}

	go func() {
		run()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}

func logChatBufferFlushSessionResults(logger chatBufferFlushLogger, result RunChatBufferFlushResult) {
	if logger == nil {
		return
	}
	for _, item := range result.FlushedSessions {
		if item.Status == "failed" {
			logger.Printf(
				"automation chat-buffer-flush session status=failed session_id=%s contact_key=%s trigger=%s job_run_id=%s reason=%s error=%s",
				item.SessionID,
				item.ContactKey,
				"SYSTEM_BUFFER_FLUSH",
				result.JobRunID,
				item.Reason,
				item.ErrorText,
			)
			continue
		}
		logger.Printf(
			"automation chat-buffer-flush session status=%s session_id=%s contact_key=%s trigger=%s job_run_id=%s reason=%s has_draft=%t tool_call_count=%d",
			item.Status,
			item.SessionID,
			item.ContactKey,
			"SYSTEM_BUFFER_FLUSH",
			result.JobRunID,
			item.Reason,
			item.DraftMessageID != "",
			item.ToolCallCount,
		)
	}
}
