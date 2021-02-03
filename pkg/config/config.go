package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/Nordix/eno/pkg/cni"
	"github.com/Nordix/eno/pkg/common"

	"gopkg.in/yaml.v2"
)

const (
	defaultKernelCni = "ovs"
)

// Configuration instance
type Configuration struct {
	KernelCni string `yaml:"KernelCni"`
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

	if c.KernelCni == "" {
		c.KernelCni = defaultKernelCni
	}
	return nil
}

func (c *Configuration) validateConfiguration() error {
	if c == nil {
		err := errors.New("Pointer to Configuration instance is nil")
		return err
	}

	if !common.SearchInSlice(c.KernelCni, cni.GetKernelSupportedCnis()) {
		err := fmt.Errorf(" %s cni is not supported currently", c.KernelCni)
		return err
	}

	return nil
}
