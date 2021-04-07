package controllers

import (
	"context"
	"errors"

	enocorev1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	enocorecommon "github.com/Nordix/eno/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

func (r *L2BridgeDomainReconciler) CreateDesiredState(ctx context.Context, log logr.Logger, nps []NodePool, links []Link, brDom *enocorev1alpha1.L2BridgeDomain) ([]string, error) {
	cpNames := brDom.Spec.ConnectionPoints
	cpNodePools := []string{}
	cpNodePoolHostnames := make(map[string][]string)
	cpNodePoolLinks := make(map[string][]Link)
	cpNodePoolInterfaces := make(map[string][]Interface)
	desiredPorts := []string{}

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

	// Create a Map that has as keys the NodePool names that are listed in the L2BridgeDomain CPs
	// and as values a list of the links that belong to each of those NodePools

	for cpNodePool, hostnames := range cpNodePoolHostnames {
		for _, link := range links {
			if enocorecommon.SearchInSlice(link.Hostname, hostnames) {
				cpNodePoolLinks[cpNodePool] = append(cpNodePoolLinks[cpNodePool], link)
			}
		}
	}

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

	// Create a Map that has as keys the SwitchNames and as values the list of SwitchPorts
	// that are relevant to L2BridgeDomain CPs

	for cpNodePool, Interfaces := range cpNodePoolInterfaces {
		for _, inter := range Interfaces {
			for _, link := range cpNodePoolLinks[cpNodePool] {
				if inter.Name == link.InterfaceName {
					desiredPorts = append(desiredPorts, link.SwitchPort)
				}
			}
		}
	}

	return desiredPorts, nil
}

func (r *L2BridgeDomainReconciler) GetActualState(ctx context.Context, log logr.Logger, nps []NodePool, links []Link, brDom *enocorev1alpha1.L2BridgeDomain) ([]string, error) {
	brDomVlan := brDom.Spec.Vlan
	nodePoolHostnames := make(map[string][]string)
	nodePoolLinks := make(map[string][]Link)
	nodePoolInterfaces := make(map[string][]Interface)
	portVlans := make(map[string][]uint16)
	interestedPorts := []string{}
	actualPorts := []string{}

	// Gather all the fabric switch ports that actually correspond to a CP
	for _, cmNodePool := range nps {
		nodeList := &corev1.NodeList{}
		listOpts := []client.ListOption{client.MatchingLabels(map[string]string{"node-pool": cmNodePool.PoolC.Name})}

		if err := r.List(ctx, nodeList, listOpts...); err != nil {
			log.Error(err, "Failed to list nodes", "Node.PoolLabel", cmNodePool.PoolC.Name)
			return nil, err
		}

		for _, node := range nodeList.Items {
			nodePoolHostnames[cmNodePool.PoolC.Name] = append(nodePoolHostnames[cmNodePool.PoolC.Name], node.Name)
		}
	}

	for cmNodePool, hostnames := range nodePoolHostnames {
		for _, link := range links {
			if enocorecommon.SearchInSlice(link.Hostname, hostnames) {
				nodePoolLinks[cmNodePool] = append(nodePoolLinks[cmNodePool], link)
			}
		}
	}

	for _, cmNodePool := range nps {
		for _, inter := range cmNodePool.PoolC.NetC.Interfaces {
			if inter.ConnPoint != "" {
				nodePoolInterfaces[cmNodePool.PoolC.Name] = append(nodePoolInterfaces[cmNodePool.PoolC.Name], inter)
			}
		}
	}

	for cmNodePool, Interfaces := range nodePoolInterfaces {
		for _, inter := range Interfaces {
			for _, link := range nodePoolLinks[cmNodePool] {
				if inter.Name == link.InterfaceName {
					interestedPorts = append(interestedPorts, link.SwitchPort)
				}
			}
		}
	}

	// Gather all the ports that are present to all ovs bridges and are part of interestedPorts list
	bridges, err := r.OvsClient.VSwitch.ListBridges()
	if err != nil {
		log.Error(err, "Failed to list OvS Bridges")
		return nil, err
	}
	for _, bridge := range bridges {
		brPorts, err := r.OvsClient.VSwitch.ListPorts(bridge)
		if err != nil {
			log.Error(err, "Failed to list OvS Ports for Bridge", bridge)
			return nil, err
		}
		for _, brPort := range brPorts {
			if enocorecommon.SearchInSlice(brPort, interestedPorts) {
				portVlans[brPort], err = r.OvsClient.VSwitch.Get.PortTrunkVlans(brPort)
				if err != nil {
					log.Error(err, "Failed to list Vlans for Port", brPort)
					return nil, err
				}
			}
		}
	}

	if len(portVlans) == 0 {
		err := errors.New("No OvS fabric ports found related to ConnectionPoints")
		log.Error(err, "")
		return nil, err
	}

	for port, vlans := range portVlans {
		for _, vlan := range vlans {
			if brDomVlan == vlan {
				actualPorts = append(actualPorts, port)
				break
			}
		}
	}
	return actualPorts, nil
}

func (r *L2BridgeDomainReconciler) Apply(ctx context.Context, log logr.Logger, brDom *enocorev1alpha1.L2BridgeDomain) error {
	delPortsVlan := []string{}
	addPortsVlan := []string{}

	brDomVlan := brDom.Spec.Vlan

	nps, links, err := r.GetPoolsAndLinks(ctx, log)
	if err != nil {
		log.Error(err, "Failed to fetch NodePool and Link info")
		return err
	}

	desiredPorts, err := r.CreateDesiredState(ctx, log, nps, links, brDom)
	if err != nil {
		log.Error(err, "Failed to create Desired state")
		return err
	}
	actualPorts, err := r.GetActualState(ctx, log, nps, links, brDom)
	if err != nil {
		log.Error(err, "Failed to create Actual state")
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
			log.Error(err, "Failed to remove Vlan from port", "Port.ID", port, "Vlan.ID", brDomVlan)
			return err
		}
	}

	for _, port := range addPortsVlan {
		err := r.OvsClient.VSwitch.Add.AddVlanToTrunk(port, brDomVlan)
		if err != nil {
			log.Error(err, "Failed to add Vlan to port", "Port.ID", port, "Vlan.ID", brDomVlan)
			return err
		}
	}

	return nil
}

func (r *L2BridgeDomainReconciler) CheckDesiredActualDiff(ctx context.Context, log logr.Logger, brDom *enocorev1alpha1.L2BridgeDomain) (bool, error) {
	exists := make(map[string]bool)

	nps, links, err := r.GetPoolsAndLinks(ctx, log)
	if err != nil {
		log.Error(err, "Failed to fetch NodePool and Link info")
		return false, err
	}

	desiredPorts, err := r.CreateDesiredState(ctx, log, nps, links, brDom)
	if err != nil {
		log.Error(err, "Failed to create Desired state")
		return false, err
	}
	actualPorts, err := r.GetActualState(ctx, log, nps, links, brDom)
	if err != nil {
		log.Error(err, "Failed to create Actual state")
		return false, err
	}

	if len(desiredPorts) != len(actualPorts) {
		return true, nil
	}
	for _, port := range desiredPorts {
		exists[port] = true
	}
	for _, port := range actualPorts {
		if !exists[port] {
			return true, nil
		}
	}
	return false, nil
}
