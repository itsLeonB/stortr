package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App
	Google
}

type App struct {
	Name    string        `default:"Stortr"`
	Env     string        `default:"dev"`
	Port    string        `default:"50051"`
	Timeout time.Duration `default:"10s"`
}

type Google struct {
	ServiceAccount string `split_words:"true" required:"true"`
	BillBucketName string `split_words:"true" required:"true"`
}

func Load() Config {
	var app App
	envconfig.MustProcess("APP", &app)

	var google Google
	envconfig.MustProcess("GOOGLE", &google)

	return Config{
		App:    app,
		Google: google,
	}
}
