package cron_infra

import (
	"github.com/robfig/cron/v3"
	cron_adapter "github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/cron"
)

func InitScheduler(emailJobs *cron_adapter.EmailJobs) *cron.Cron {
	c := cron.New()

	c.AddFunc("*/5 * * * *", emailJobs.RunVerifyAndSaveInbox)
	c.AddFunc("*/5 * * * *", emailJobs.RunSyncEmailStatus)

	return c
}
