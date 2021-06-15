package controllers

import (
	"context"

	enocorev1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type NodePool struct {
	PoolC PoolConf `yaml:"poolConf"`
}

type PoolConf struct {
	Name string  `yaml:"name"`
	NetC NetConf `yaml:"netConf"`
}

type NetConf struct {
	Interfaces []Interface `yaml:"interfaces"`
}
type Interface struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	State     string `yaml:"state"`
	ConnPoint string `yaml:"connectionPoint"`
}

type Link struct {
	Hostname      string `yaml:"hostname"`
	InterfaceName string `yaml:"interfaceName"`
	SwitchName    string `yaml:"switchName"`
	SwitchPort    string `yaml:"switchPort"`
}

func (r *L2BridgeDomainReconciler) GetPoolsAndLinks(ctx context.Context, log logr.Logger) ([]NodePool, []Link, error) {
	nps := []NodePool{}
	links := []Link{}

	// Read the Pool Configuration from the ConfigMap
	poolConf := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: r.Config.PoolConfigMapName, Namespace: r.Config.FabricPluginNamespace}, poolConf); err != nil {
		log.Error(err, "Failed to find ConfigMap", "ConfigMap.Name", r.Config.PoolConfigMapName)
		return nil, nil, err
	}

	// Read the Fabric configuration from the ConfigMap
	fabricConf := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: r.Config.FabricConfigMapName, Namespace: r.Config.FabricPluginNamespace}, fabricConf); err != nil {
		log.Error(err, "Failed to find ConfigMap", "ConfigMap.Name", r.Config.FabricConfigMapName)
		return nil, nil, err
	}
	// Unmarshal poolConf Object
	err := yaml.Unmarshal([]byte(poolConf.Data["nodePools"]), &nps)
	if err != nil {
		log.Error(err, "Failed to Unmarshal", "ConfigMap.Name", r.Config.PoolConfigMapName)
		return nil, nil, err
	}
	// Unmarshal fabricConf Object
	err = yaml.Unmarshal([]byte(fabricConf.Data["links"]), &links)
	if err != nil {
		log.Error(err, "Failed to Unmarshal", "ConfigMap.Name", r.Config.FabricConfigMapName)
		return nil, nil, err
	}

	return nps, links, nil

}

// UpdateStatus - Update the Status of L2BridgeDomain instance
func (r *L2BridgeDomainReconciler) UpdateStatus(ctx context.Context, log logr.Logger, phase, message string, brDom *enocorev1alpha1.L2BridgeDomain) error {
	brDom.Status.Phase = phase
	brDom.Status.Message = message
	if err := r.Status().Update(ctx, brDom); err != nil {
		log.Error(err, "Failed to update L2BridgeDomain status")
		return err
	}
	return nil

}
