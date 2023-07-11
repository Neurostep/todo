package config

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type (
	Config struct {
		Database Database `yaml:"database" validate:"required,dive"`
		Server   Server   `yaml:"server" validate:"required,dive"`
		Metrics  Metrics  `yaml:"metrics"`
	}

	Database struct {
		Address string `yaml:"address" validate:"required"`
	}

	Server struct {
		Debug       bool `yaml:"debug"`
		Port        int  `yaml:"port"`
		AuthEnabled bool `yaml:"authEnabled"`
	}

	Metrics struct {
		TracingEnable bool `yaml:"tracingEnable"`
	}
)

func ReadConfigFile(configPath string) (Config, error) {
	var c Config
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return c, errors.Errorf("failed to read config file %q: %s", configPath, err)
	}
	if err := yaml.Unmarshal([]byte(os.ExpandEnv(string(raw))), &c); err != nil {
		return c, errors.Wrap(err, "yaml parse failed")
	}

	return c, validator.New().Struct(c)
}
