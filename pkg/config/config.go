package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	defaultUseFabricPlugin = false
)

// Configuration instance
type Configuration struct {
	UseFabricPlugin bool `yaml:"UseFabricPlugin"`
}

// NewConfiguration - creates instance of Configuration
func NewConfiguration() *Configuration {
	return &Configuration{}
}

// GetConfiguration - Fills Configuration instance with the appropriate configuration values
func (c *Configuration) GetConfiguration(confFile string) error {
	yamlFile, err := ioutil.ReadFile(confFile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return err
	}

	if err := c.addDefault(); err != nil {
		return err
	}

	if err := c.validateConfiguration(); err != nil {
		return err
	}
	return nil

}

func (c *Configuration) addDefault() error {
	if c == nil {
		err := errors.New("Pointer to Configuration instance is nil")
		return err
	}

	if !c.UseFabricPlugin {
		c.UseFabricPlugin = defaultUseFabricPlugin
	}
	return nil
}

func (c *Configuration) validateConfiguration() error {
	if c == nil {
		err := errors.New("Pointer to Configuration instance is nil")
		return err
	}

	return nil
}
