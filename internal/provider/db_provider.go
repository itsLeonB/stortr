package provider

import (
	"crypto/tls"
	"errors"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/stortr/internal/config"
)

type DBs struct {
	AsynqClient    *asynq.Client
	AsynqInspector *asynq.Inspector
}

func ProvideDBs(logger ezutil.Logger, valkeyCfg config.Valkey) *DBs {
	return &DBs{
		connectAsynqClient(logger, valkeyCfg),
		connectAsynqInspector(logger, valkeyCfg),
	}
}

func (d *DBs) Ping() error {
	var errs error

	if d.AsynqClient != nil {
		if err := d.AsynqClient.Ping(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if d.AsynqInspector != nil {
		if _, err := d.AsynqInspector.Queues(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (d *DBs) Shutdown() error {
	var errs error
	if d.AsynqClient != nil {
		if err := d.AsynqClient.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if d.AsynqInspector != nil {
		if err := d.AsynqInspector.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func connectAsynqClient(logger ezutil.Logger, cfg config.Valkey) *asynq.Client {
	if cfg.Addr == "" {
		logger.Warn("valkey config not provided, will not init asynq client")
		return nil
	}

	return asynq.NewClient(getAsynqConnOpts(cfg))
}

func connectAsynqInspector(logger ezutil.Logger, cfg config.Valkey) *asynq.Inspector {
	if cfg.Addr == "" {
		logger.Warn("valkey config not provided, will not init asynq inspector")
		return nil
	}

	return asynq.NewInspector(getAsynqConnOpts(cfg))
}

func getAsynqConnOpts(cfg config.Valkey) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}
