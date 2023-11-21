package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/bueti/shrinkster/internal/model"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const (
	AppName = "shrinkster"
	CfgFile = "config.yaml"
)

type Config struct {
	ID    uuid.UUID `yaml:"id"`
	Email string    `yaml:"email"`
}

// Load reads the config file and returns the config.
func Load() (Config, error) {
	dir, err := xdg.ConfigFile(AppName)
	if err != nil {
		return Config{}, err
	}
	fullPath := filepath.Join(dir, CfgFile)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("can't open config file: %s", fullPath)
		}
		return Config{}, err
	}

	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

// Save writes the config to the config file.
func Save(user model.UserLoginResponse) error {
	fullPath, err := xdg.ConfigFile(AppName)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(fullPath, CfgFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	config := Config{
		ID:    user.ID,
		Email: user.Email,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		f.Close() // ignore error; SetUsername error takes precedence
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
