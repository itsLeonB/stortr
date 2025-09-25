package job

import (
	"context"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/stortr/internal/config"
	"github.com/itsLeonB/stortr/internal/provider"
	"github.com/itsLeonB/stortr/internal/service"
)

type cleanupOrphanedBillsJob struct {
	expenseBillSvc service.ExpenseBillService
}

func CleanupOrphanedBillsJob(configs config.Config) *ezutil.Job {
	logger := provider.ProvideLogger("Cleanup Orphaned Bills", configs.Env)
	providers := provider.All(configs, logger)
	jobImpl := cleanupOrphanedBillsJob{providers.ExpenseBill}

	return ezutil.NewJob(logger, jobImpl.Run).
		WithSetupFunc(providers.Ping).
		WithCleanupFunc(providers.Shutdown)
}

func (j *cleanupOrphanedBillsJob) Run() error {
	return j.expenseBillSvc.CleanupOrphaned(context.Background())
}
