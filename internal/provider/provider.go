package provider

import (
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/stortr/internal/config"
)

type Provider struct {
	Logger ezutil.Logger
	*Repositories
	*Services
}

func All(configs config.Config) *Provider {
	logger := ProvideLogger(configs.App)
	repos := ProvideRepositories(configs)

	return &Provider{
		Logger:       logger,
		Repositories: repos,
		Services:     ProvideServices(configs.Google, repos),
	}
}

func (p *Provider) Shutdown() error {
	return p.Repositories.Shutdown()
}
