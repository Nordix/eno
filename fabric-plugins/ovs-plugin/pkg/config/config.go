package config

import (
	"fmt"
	"os"
)

const (
	fabricPluginNamespace = "FABRIC_PLUGIN_NAMESPACE"
	poolConfigMapName     = "POOL_CONFIGMAP_NAME"
	fabricConfigMapName   = "FABRIC_CONFIGMAP_NAME"
	fabricOvsIp           = "FABRIC_OVS_IP"
	fabricOvsPort         = "FABRIC_OVS_PORT"
)

// Configuration instance
type Configuration struct {
	FabricPluginNamespace string
	PoolConfigMapName     string
	FabricConfigMapName   string
	FabricOvsIp           string
	FabricOvsPort         string
}

// NewConfiguration - creates instance of Configuration
func NewConfiguration() *Configuration {
	return &Configuration{}
}

// GetConfiguration - Fills Configuration instance with the appropriate configuration values
func (c *Configuration) GetConfiguration() error {
	var err error

	c.FabricPluginNamespace, err = getEnv(fabricPluginNamespace)
	if err != nil {
		return err
	}
	c.PoolConfigMapName, err = getEnv(poolConfigMapName)
	if err != nil {
		return err
	}
	c.FabricConfigMapName, err = getEnv(fabricConfigMapName)
	if err != nil {
		return err
	}
	c.FabricOvsIp, err = getEnv(fabricOvsIp)
	if err != nil {
		return err
	}
	c.FabricOvsPort, err = getEnv(fabricOvsPort)
	if err != nil {
		return err
	}

	return nil

}

func getEnv(envVar string) (string, error) {
	value, ok := os.LookupEnv(envVar)

	if !ok {
		err := fmt.Errorf("The Env Variable %s is not defined", envVar)
		return "", err
	}

	return value, nil
}
