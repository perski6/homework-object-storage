package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Bucket  string `envconfig:"BUCKET" required:"true"`
	Port    string `envconfig:"HTTP_PORT" required:"true"`
	Timeout int    `envconfig:"TIMEOUT" required:"true"`
}

var App Config

func init() {
	envconfig.MustProcess("", &App)
}
