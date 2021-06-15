package l2serviceattachmentparser

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Nordix/eno/pkg/cni/cniconfig"
	"github.com/Nordix/eno/pkg/parser"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/Nordix/eno/pkg/cni"
	"github.com/Nordix/eno/pkg/common"
	"github.com/Nordix/eno/pkg/config"
	"github.com/go-logr/logr"
)

// L2SrvAttParser instance
type L2SrvAttParser struct {
	srvAttResource  *enov1alpha1.L2ServiceAttachment
	cpResource      *enov1alpha1.ConnectionPoint
	l2srvResources  []*enov1alpha1.L2Service
	subnetResources []*enov1alpha1.Subnet
	routeResources  map[string][]*enov1alpha1.Route
	config          *config.Configuration
	cniMapping      map[string]cni.Cnier
	ipamMapping     map[string]cni.Ipam
	log             logr.Logger
}

// NewL2SrvAttParser - creates instance of L2SrvAttParser
func NewL2SrvAttParser(srvAttObj *enov1alpha1.L2ServiceAttachment, l2srvObjs []*enov1alpha1.L2Service,
	cpObj *enov1alpha1.ConnectionPoint, subnetObjs []*enov1alpha1.Subnet,
	routeObjs map[string][]*enov1alpha1.Route, c *config.Configuration, mc map[string]cni.Cnier, ipamMap map[string]cni.Ipam, logger logr.Logger) *L2SrvAttParser {
	return &L2SrvAttParser{srvAttResource: srvAttObj,
		cpResource:      cpObj,
		l2srvResources:  l2srvObjs,
		subnetResources: subnetObjs,
		routeResources:  routeObjs,
		config:          c,
		cniMapping:      mc,
		ipamMapping:     ipamMap,
		log:             logger}
}

// ParseL2ServiceAttachment - parses a L2ServiceAttachment Resource
func (sattp *L2SrvAttParser) ParseL2ServiceAttachment(data map[string]interface{}) (string, string, error) {
	cniManifestFile, err := sattp.populateCni(data)
	if err != nil {
		sattp.log.Error(err, "Error occurred while populating cni")
		return "", "", err
	}
	var ipamManifestFile string
	if len(sattp.subnetResources) > 0 {
		ipamManifestFile, err = sattp.populateIpam(data)
		if err != nil {
			sattp.log.Error(err, "Error occurred while populating ipam")
			return "", "", err
		}
	}
	return cniManifestFile, ipamManifestFile, nil
}

func (sattp *L2SrvAttParser) populateCni(data map[string]interface{}) (string, error) {

	if err := sattp.validateCpSupportedCnis(); err != nil {
		sattp.log.Error(err, "Error occurred while validating ConnectionPoint's supported CNIs")
		return "", err
	}
	if err := sattp.pickCni(); err != nil {
		sattp.log.Error(err, "Error occurred while picking cni")
		return "", err
	}

	cniObj, ok := sattp.cniMapping[sattp.srvAttResource.Spec.Implementation]
	if !ok {
		err := fmt.Errorf("The %s is not supported by ENO yet", sattp.srvAttResource.Spec.Implementation)
		sattp.log.Error(err, "")
		return "", err
	}
	cniConfigObj, err := sattp.instantiateCniConfig()
	if err != nil {
		sattp.log.Error(err, "Error occured while instantiating cni config")
		return "", err
	}
	cniManifestFile, err := cniObj.HandleCni(cniConfigObj, data)
	if err != nil {
		sattp.log.Error(err, "Error occured while handling cni")
		return "", err
	}
	return cniManifestFile, nil
}

func (sattp *L2SrvAttParser) populateIpam(data map[string]interface{}) (string, error) {
	var ipType, ipamType string
	for i, subnet := range sattp.subnetResources {
		if i > 0 {
			if ipType == subnet.Spec.Type {
				err := errors.New("subnets in one L2Service should not be of same ip type")
				sattp.log.Error(err, "")
				return "", err
			}
			if ipamType != subnet.Spec.Ipam {
				err := errors.New("subnets in one L2Service should have same ipam type")
				sattp.log.Error(err, "")
				return "", err
			}
		}
		ipType = subnet.Spec.Type
		ipamType = subnet.Spec.Ipam
		sp := parser.NewSubnetParser(subnet, sattp.log)
		if err := sp.ValidateSubnet(); err != nil {
			return "", err
		}
		if err := sp.ValidateRoute(sattp.routeResources[subnet.GetName()]); err != nil {
			return "", err
		}
	}
	ipam := sattp.ipamMapping[ipamType]
	ipamConfigObj := sattp.instantiateIpamConfig()
	ipamManifestFile, err := ipam.HandleIpam(ipamConfigObj, data)
	if err != nil {
		return "", err
	}
	return ipamManifestFile, nil
}

// validateCpSupportedCnis - validate the SupportedCnis list in CP CR
func (sattp *L2SrvAttParser) validateCpSupportedCnis() error {
	var supportedCniUniqueness []string

	//Check uniqueness of the SupportedCni Items
	for _, supportedCni := range sattp.cpResource.Spec.SupportedCnis {
		if common.SearchInSlice(supportedCni.Name, supportedCniUniqueness) {
			err := errors.New("SupportedCni items must be unique")
			return err
		}
		supportedCniUniqueness = append(supportedCniUniqueness, supportedCni.Name)

		// Check if the Supported Interface types are unique and valid
		supportedInterfaceTypesUniqueness := []string{}

		for _, supportedInterfaceType := range supportedCni.SupportedInterfaceTypes {

			if common.SearchInSlice(supportedInterfaceType, supportedInterfaceTypesUniqueness) {
				err := errors.New("SupportedInterfaceTypes items must be unique")
				return err
			}

			if !common.SearchInSlice(supportedInterfaceType, common.GetValidInterfaceTypes()) {
				err := fmt.Errorf(" %s is not a valid interface type", supportedInterfaceType)
				return err
			}

			supportedInterfaceTypesUniqueness = append(supportedInterfaceTypesUniqueness, supportedInterfaceType)
		}
	}

	return nil
}

// pickCni - pick the CNI to be used for net-attach-def
func (sattp *L2SrvAttParser) pickCni() error {
	if sattp.srvAttResource.Spec.Implementation == "" {
		// No CNI has been specified
		if sattp.srvAttResource.Spec.PodInterfaceType == "" {
			sattp.srvAttResource.Spec.Implementation = sattp.cpResource.Spec.SupportedCnis[0].Name
			sattp.srvAttResource.Spec.PodInterfaceType = sattp.cpResource.Spec.SupportedCnis[0].SupportedInterfaceTypes[0]
			return nil
		} else {
			for _, supportedCni := range sattp.cpResource.Spec.SupportedCnis {
				if common.SearchInSlice(sattp.srvAttResource.Spec.PodInterfaceType, supportedCni.SupportedInterfaceTypes) {
					sattp.srvAttResource.Spec.Implementation = supportedCni.Name
					return nil
				}
			}

			err := fmt.Errorf(" %s interface type is not supported by the supported CNIs", sattp.srvAttResource.Spec.PodInterfaceType)
			return err
		}
	}
	// CNI has been specified
	for _, supportedCni := range sattp.cpResource.Spec.SupportedCnis {
		if sattp.srvAttResource.Spec.Implementation == supportedCni.Name {
			if sattp.srvAttResource.Spec.PodInterfaceType == "" {
				sattp.srvAttResource.Spec.PodInterfaceType = supportedCni.SupportedInterfaceTypes[0]
				return nil
			} else {
				if common.SearchInSlice(sattp.srvAttResource.Spec.PodInterfaceType, supportedCni.SupportedInterfaceTypes) {
					return nil
				}

				err := fmt.Errorf(" %s interface type is not supported by %s CNI", sattp.srvAttResource.Spec.PodInterfaceType,
					sattp.srvAttResource.Spec.Implementation)
				return err
			}
		}
	}
	err := fmt.Errorf(" %s CNI is not supported by the supported CNIs", sattp.srvAttResource.Spec.Implementation)
	return err
}

// instantiateCniConfig - Instantiate the CniConfig object to be used
func (sattp *L2SrvAttParser) instantiateCniConfig() (*cniconfig.CniConfig, error) {
	segIds := sattp.getSegIds()
	cniOptions, err := sattp.getCniOptions()
	if err != nil {
		sattp.log.Error(err, "")
		return nil, err
	}
	return cniconfig.NewCniConfig(segIds,
		sattp.srvAttResource.Spec.VlanType,
		sattp.srvAttResource.Spec.PodInterfaceType,
		cniOptions,
		sattp.log), nil
}

// instantiateIpamConfig - Instantiate the IpamConfig object to be used
func (sattp *L2SrvAttParser) instantiateIpamConfig() *cniconfig.IpamConfig {
	return cniconfig.NewIpamConfig(sattp.subnetResources, sattp.routeResources, sattp.log)
}

// getSegIds - returns a list with the segmentation Ids
func (sattp *L2SrvAttParser) getSegIds() []uint16 {
	var tmpslice []uint16
	for _, l2srvObj := range sattp.l2srvResources {
		tmpslice = append(tmpslice, l2srvObj.Spec.SegmentationID)
	}
	return tmpslice
}

// getCniOptions - Returns a map or an error of the picked CNI Opts
func (sattp *L2SrvAttParser) getCniOptions() (map[string]interface{}, error) {
	rawCniOptions := make(map[string]interface{})

	for _, supportedCni := range sattp.cpResource.Spec.SupportedCnis {
		if sattp.srvAttResource.Spec.Implementation == supportedCni.Name {
			if supportedCni.Opts != "" {
				optionsBytes := []byte(supportedCni.Opts)
				err := json.Unmarshal(optionsBytes, &rawCniOptions)
				if err != nil {
					sattp.log.Error(err, "")
					return nil, err
				}
				return rawCniOptions, nil
			}

			return rawCniOptions, nil
		}
	}
	err := errors.New("Error occured while getting Cni Opts")
	return nil, err

}
