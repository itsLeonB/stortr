package provider

import (
	"errors"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/meq"
	"github.com/itsLeonB/stortr/internal/config"
)

type Provider struct {
	Logger ezutil.Logger
	meq.DB
	*Repositories
	*Services
}

func All(configs config.Config, logger ezutil.Logger) *Provider {
	db := meq.NewAsynqDB(logger, configs.ToRedisOpts())
	repos := ProvideRepositories(configs, db, logger)

	return &Provider{
		Logger:       logger,
		DB:           db,
		Repositories: repos,
		Services:     ProvideServices(configs.Google, repos, logger),
	}
}

func (p *Provider) Shutdown() error {
	var errs error

	if p.DB != nil {
		if e := p.DB.Shutdown(); e != nil {
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
