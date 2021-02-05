package parser

import (
	"fmt"
	"net"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/go-logr/logr"
)

// RouteParser instance
type RouteParser struct {
	SubnetResource *enov1alpha1.Subnet
	RouteResource  []*enov1alpha1.Route
	log            logr.Logger
}

// NewRouteParser parses Route resources
func NewRouteParser(subnetObj *enov1alpha1.Subnet, routes []*enov1alpha1.Route, logger logr.Logger) *RouteParser {
	return &RouteParser{SubnetResource: subnetObj,
		RouteResource: routes,
		log:           logger}
}

// ValidateRoute validates Route
func (sp *RouteParser) ValidateRoute() error {
	mask := sp.SubnetResource.Spec.Mask
	cidr := sp.SubnetResource.Spec.Address + "/" + fmt.Sprint(mask)
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		err := fmt.Errorf("invalid CIDR: %s", cidr)
		sp.log.Error(err, "")
		return err
	}

	var route *enov1alpha1.Route
	for _, route = range sp.RouteResource {
		if !ipnet.Contains(net.ParseIP(route.Spec.NextHop)) {
			err := fmt.Errorf("nextHop %s of route %s doesnot belong to CIDR %s", route.Spec.NextHop, route.Spec.Prefix, cidr)
			sp.log.Error(err, "")
			return err
		}
		routeCidr := route.Spec.Prefix + "/" + fmt.Sprint(route.Spec.Mask)
		_, routeIpnet, err := net.ParseCIDR(routeCidr)
		if err != nil {
			err := fmt.Errorf("invalid CIDR: %s", routeCidr)
			sp.log.Error(err, "")
			return err
		}
		if route.Spec.Prefix != routeIpnet.IP.String() {
			err := fmt.Errorf("invalid Prefix %s,Route Prefix field should be a valid subnet", route.Spec.Prefix)
			sp.log.Error(err, "")
			return err
		}
	}
	return nil
}
