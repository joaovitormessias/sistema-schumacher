package automation

import (
	"context"
	"time"

	"schumacher-tur/api/internal/shared/config"
)

func StartChatAutoSendRetryLoop(ctx context.Context, svc *Service, cfg config.Config, logger chatBufferFlushLogger) {
	if svc == nil || !cfg.ChatAutoSendRetryEnabled {
		return
	}

	interval := time.Duration(cfg.ChatAutoSendRetryIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}

	run := func() {
		result, err := svc.runChatAutoSendRetry(ctx, RunChatAutoSendRetryInput{
			Limit:           cfg.ChatAutoSendRetryLimit,
			CooldownSeconds: cfg.ChatAutoSendRetryCooldownSeconds,
			Channel:         "WHATSAPP",
			Status:          "ACTIVE",
			HandoffStatus:   "BOT",
		}, "SYSTEM")
		if logger == nil {
			return
		}
		if err != nil {
			logger.Printf("automation chat-auto-send-retry failed: %v", err)
			return
		}
		logger.Printf(
			"automation chat-auto-send-retry status=%s reason=%s checked=%d due=%d retried=%d blocked=%d skipped=%d failed=%d",
			result.Status,
			result.Reason,
			result.CheckedCount,
			result.DueCount,
			result.RetriedCount,
			result.BlockedCount,
			result.SkippedCount,
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
