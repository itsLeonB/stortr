package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

const AppName = "Stortr"

type Config struct {
	App
	Google
	Valkey
}

type App struct {
	Env     string        `default:"dev"`
	Port    string        `default:"50051"`
	Timeout time.Duration `default:"10s"`
}

type Google struct {
	ServiceAccount string `split_words:"true" required:"true"`
}

func Load() Config {
	var app App
	envconfig.MustProcess("APP", &app)

	var google Google
	envconfig.MustProcess("GOOGLE", &google)

	var valkey Valkey
	envconfig.MustProcess("VALKEY", &valkey)

	return Config{
		App:    app,
		Google: google,
		Valkey: valkey,
	}
}
