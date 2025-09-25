package provider

import (
	"errors"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/stortr/internal/config"
)

type Provider struct {
	Logger ezutil.Logger
	*DBs
	*Repositories
	*Services
}

func All(configs config.Config, logger ezutil.Logger) *Provider {
	dbs := ProvideDBs(logger, configs.Valkey)
	repos := ProvideRepositories(configs, dbs, logger)

	return &Provider{
		Logger:       logger,
		DBs:          dbs,
		Repositories: repos,
		Services:     ProvideServices(configs.Google, repos, logger),
	}
}

func (p *Provider) Shutdown() error {
	var errs error

	if p.DBs != nil {
		if e := p.DBs.Shutdown(); e != nil {
			errs = errors.Join(errs, e)
		}
	}

	if p.Repositories != nil {
		if e := p.Repositories.Shutdown(); e != nil {
			errs = errors.Join(errs, e)
		}
	}

	return errs
}
