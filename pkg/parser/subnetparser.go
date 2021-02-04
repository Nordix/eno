package parser

import (
	"bytes"
	"fmt"
	"net"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/go-logr/logr"
)

// SubnetParser instance
type SubnetParser struct {
	subnetResource *enov1alpha1.Subnet
	log            logr.Logger
}

func NewSubnetParser(subnetObj *enov1alpha1.Subnet, logger logr.Logger) *SubnetParser {
	return &SubnetParser{subnetResource: subnetObj,
		log: logger}
}

func (sp *SubnetParser) ValidateSubnet() error {
	ipStr := sp.subnetResource.Spec.Address
	ipAddr := net.ParseIP(ipStr)
	if ipAddr == nil {
		err := fmt.Errorf("Invalid IP: %s", ipStr)
		sp.log.Error(err, "")
		return err
	}
	ipType := sp.subnetResource.Spec.Type
	mask := sp.subnetResource.Spec.Mask
	if ipType == "v4" && mask >= 32 {
		err := fmt.Errorf("Invalid Mask for ipv4: %v", mask)
		sp.log.Error(err, "")
		return err
	}
	if ipType == "v6" && mask >= 128 {
		err := fmt.Errorf("Invalid Mask for ipv6: %v", mask)
		sp.log.Error(err, "")
		return err
	}
	cidr := ipStr + "/" + fmt.Sprint(mask)
	err := sp.validateAllocationPool(cidr, sp.subnetResource.Spec.AllocationPool)
	if err != nil {
		return fmt.Errorf("failed populating whereabouts ipam config: %s", err)
	}
	return nil
}

func (sp *SubnetParser) validateAllocationPool(cidr string, ipPools []enov1alpha1.IpPool) error {
	for _, ipPool := range ipPools {
		startIP := net.ParseIP(ipPool.Start)
		if startIP == nil {
			err := fmt.Errorf("Invalid start IP: %s", startIP)
			sp.log.Error(err, "")
			return err
		}
		endIP := net.ParseIP(ipPool.End)
		if endIP == nil {
			err := fmt.Errorf("Invalid end IP: %s", endIP)
			sp.log.Error(err, "")
			return err
		}
		//TODO check heterogeneous ip types
		if bytes.Compare(startIP, endIP) > 0 {
			err := fmt.Errorf("Start IP  %s should be lesser than end IP  %s", startIP, endIP)
			sp.log.Error(err, "")
			return err
		}
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			err := fmt.Errorf("Invalid CIDR: %s", cidr)
			sp.log.Error(err, "")
			return err
		}
		if !ipnet.Contains(startIP) && !ipnet.Contains(endIP) {
			err := fmt.Errorf("StartIP %s and/or EndIP %s not in cidr: %s", startIP, endIP, cidr)
			sp.log.Error(err, "")
			return err
		}

	}
	return nil

}

func (sp *SubnetParser) ValidateRoute(routes []*enov1alpha1.Route) error {
	rp := NewRouteParser(sp.subnetResource, routes, sp.log)
	if err := rp.ValidateRoute(); err != nil {
		return err
	}
	return nil
}
