package controllers

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"

	enocorev1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	enocorecommon "github.com/Nordix/eno/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	enofabricv1alpha1 "github.com/Nordix/eno/fabric-plugins/ovs-plugin/api/v1alpha1"
	"github.com/go-logr/logr"
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

func (r *L2BridgeDomainReconciler) CreateDesiredState(ctx context.Context, log logr.Logger, brDom *enofabricv1alpha1.L2BridgeDomain) ([]string, error) {
	cpNames := brDom.Spec.ConnectionPoints
	cpNodePools := []string{}
	cpNodePoolHostnames := make(map[string][]string)
	cpNodePoolLinks := make(map[string][]Link)
	cpNodePoolInterfaces := make(map[string][]Interface)
	//desiredSwitchPorts := make(map[string][]string)
	desiredPorts := []string{}
	nps := []NodePool{}
	links := []Link{}

	// Unmarshal the Pool Configuration from the ConfigMap
	// to a struct

	poolConf := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: r.Config.PoolConfigMapName, Namespace: r.Config.FabricPluginNamespace}, poolConf); err != nil {
		log.Error(err, "Failed to find ConfigMap", "ConfigMap.Name", r.Config.PoolConfigMapName)
		return nil, err
	}

	// Unmarshal the Fabric configuration from the ConfigMap
	// to a struct

	fabricConf := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: r.Config.FabricConfigMapName, Namespace: r.Config.FabricPluginNamespace}, fabricConf); err != nil {
		log.Error(err, "Failed to find ConfigMap", "ConfigMap.Name", r.Config.FabricConfigMapName)
		return nil, err
	}

	err := yaml.Unmarshal([]byte(poolConf.Data["nodePools"]), &nps)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}
	//fmt.Print(poolConf.Data["nodePools"])
	//fmt.Printf("%+v\n", nps)

	err = yaml.Unmarshal([]byte(fabricConf.Data["links"]), &links)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}
	//fmt.Print(poolConf.Data["nodePools"])
	//fmt.Printf("%+v\n", links)

	// Create a list which holds the nodePool names that are included in the
	// relevant L2BridgeDomain CPs. Each item of that list is unique

	for _, cpName := range cpNames {
		tempObj := &enocorev1alpha1.ConnectionPoint{}
		if err := r.Get(ctx, types.NamespacedName{Name: cpName}, tempObj); err != nil {
			log.Error(err, "Failed to find ConnectionPoint", "ConnectionPoint.Name", cpName)
			return nil, err
		}
		if !enocorecommon.SearchInSlice(tempObj.Spec.NodePool, cpNodePools) {
			cpNodePools = append(cpNodePools, tempObj.Spec.NodePool)
		}
	}

	//fmt.Printf("%+v\n", cpNodePools)

	// Create a Map which has as keys the NodePool names that are listed in the L2BridgeDomain CPs
	// and as values a list of all the Nodes that belong to each pool.

	for _, cpNodePool := range cpNodePools {
		nodeList := &corev1.NodeList{}
		listOpts := []client.ListOption{client.MatchingLabels(map[string]string{"node-pool": cpNodePool})}

		if err := r.List(ctx, nodeList, listOpts...); err != nil {
			log.Error(err, "Failed to list nodes", "Node.PoolLabel", cpNodePool)
			return nil, err
		}

		for _, node := range nodeList.Items {
			cpNodePoolHostnames[cpNodePool] = append(cpNodePoolHostnames[cpNodePool], node.Name)
		}
	}

	//fmt.Printf("%+v\n", cpNodePoolHostnames)

	// Create a Map that has as keys the NodePool names that are listed in the L2BridgeDomain CPs
	// and as values a list of the links that belong to each of those NodePools

	for cpNodePool, hostnames := range cpNodePoolHostnames {
		for _, link := range links {
			if enocorecommon.SearchInSlice(link.Hostname, hostnames) {
				cpNodePoolLinks[cpNodePool] = append(cpNodePoolLinks[cpNodePool], link)
			}
		}
	}

	//fmt.Printf("%+v\n", cpNodePoolLinks)

	// Create a Map that has as keys the NodePool names that are listed in the L2BridgeDomain CPs
	// and as values a list of all the interfaces that are relevant to the L2BridgeDomain CPs

	for _, cmNodePool := range nps {
		if enocorecommon.SearchInSlice(cmNodePool.PoolC.Name, cpNodePools) {
			for _, inter := range cmNodePool.PoolC.NetC.Interfaces {
				if inter.ConnPoint != "" && enocorecommon.SearchInSlice(inter.ConnPoint, cpNames) {
					cpNodePoolInterfaces[cmNodePool.PoolC.Name] = append(cpNodePoolInterfaces[cmNodePool.PoolC.Name], inter)
				}
			}
		}
	}

	//fmt.Printf("%+v\n", cpNodePoolInterfaces)

	// Create a Map that has as keys the SwitchNames and as values the list of SwitchPorts
	// that are relevant to L2BridgeDomain CPs

	for cpNodePool, Interfaces := range cpNodePoolInterfaces {
		for _, inter := range Interfaces {
			for _, link := range cpNodePoolLinks[cpNodePool] {
				if inter.Name == link.InterfaceName {
					//desiredSwitchPorts[link.SwitchName] = append(desiredSwitchPorts[link.SwitchName], link.SwitchPort)
					desiredPorts = append(desiredPorts, link.SwitchPort)
				}
			}
		}
	}

	//fmt.Printf("%+v\n", desiredSwitchPorts)
	fmt.Printf("%+v\n", desiredPorts)

	return desiredPorts, nil
}

func (r *L2BridgeDomainReconciler) GetActualState(brDom *enofabricv1alpha1.L2BridgeDomain) ([]string, error) {
	brDomVlan := brDom.Spec.Vlan
	portVlans := make(map[string][]uint16)
	actualPorts := []string{}

	bridges, err := r.OvsClient.VSwitch.ListBridges()
	if err != nil {
		return nil, err
	}
	for _, bridge := range bridges {
		brPorts, err := r.OvsClient.VSwitch.ListPorts(bridge)
		if err != nil {
			return nil, err
		}
		for _, brPort := range brPorts {
			portVlans[brPort], err = r.OvsClient.VSwitch.Get.PortTrunkVlans(brPort)
			if err != nil {
				return nil, err
			}
		}
	}

	for port, vlans := range portVlans {
		for _, vlan := range vlans {
			if brDomVlan == vlan {
				actualPorts = append(actualPorts, port)
				break
			}
		}
	}
	fmt.Printf("%+v\n", actualPorts)
	//fmt.Printf("%+v\n", portVlans)
	//fmt.Printf("%+v\n", bridges)
	return actualPorts, nil
}

func (r *L2BridgeDomainReconciler) Apply(ctx context.Context, log logr.Logger, brDom *enofabricv1alpha1.L2BridgeDomain) error {
	delPortsVlan := []string{}
	addPortsVlan := []string{}

	brDomVlan := brDom.Spec.Vlan
	desiredPorts, err := r.CreateDesiredState(ctx, log, brDom)
	if err != nil {
		return err
	}
	actualPorts, err := r.GetActualState(brDom)
	if err != nil {
		return err
	}

	for _, port := range actualPorts {
		if !enocorecommon.SearchInSlice(port, desiredPorts) {
			delPortsVlan = append(delPortsVlan, port)
		}
	}

	for _, port := range desiredPorts {
		if !enocorecommon.SearchInSlice(port, actualPorts) {
			addPortsVlan = append(addPortsVlan, port)
		}
	}

	for _, port := range delPortsVlan {
		err := r.OvsClient.VSwitch.Remove.RemoveVlanFromTrunk(port, brDomVlan)
		if err != nil {
			return err
		}
	}

	for _, port := range addPortsVlan {
		err := r.OvsClient.VSwitch.Add.AddVlanToTrunk(port, brDomVlan)
		if err != nil {
			return err
		}
	}

	return nil
}
