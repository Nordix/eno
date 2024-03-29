package controllers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"text/template"

	"io"
	"strings"

	"github.com/Nordix/eno/api/v1alpha1"
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/Nordix/eno/pkg/common"
	"github.com/Nordix/eno/pkg/l2serviceattachmentparser"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const nadFileName = "netattachdef.yaml"

// UpdateStatus - Update the Status of L2ServiceAttachment instance
func (r *L2ServiceAttachmentReconciler) UpdateStatus(ctx context.Context, req ctrl.Request, log logr.Logger, phase, message string) error {

	svcAttSt := &enov1alpha1.L2ServiceAttachment{}
	if err := r.Get(ctx, req.NamespacedName, svcAttSt); err != nil {
		log.Error(err, "Failed to get L2ServiceAttachment")
		return err
	}
	svcAttSt.Status.Phase = phase
	svcAttSt.Status.Message = message
	if err := r.Status().Update(ctx, svcAttSt); err != nil {
		log.Error(err, "Failed to update L2ServiceAttachment status")
		return err
	}
	return nil

}

// DefineNetAttachDef - Create and returns an Unstructured net-attach-def
func (r *L2ServiceAttachmentReconciler) DefineNetAttachDef(ctx context.Context, log logr.Logger, s *enov1alpha1.L2ServiceAttachment) (*uns.Unstructured, error) {

	data := make(map[string]interface{})
	l2srvObjs := []*enov1alpha1.L2Service{}

	cpName := s.Spec.ConnectionPoint
	l2srvNames := s.Spec.L2Services

	// Get the ConnectionPoint resource
	cp := &enov1alpha1.ConnectionPoint{}
	if err := r.Get(ctx, types.NamespacedName{Name: cpName}, cp); err != nil {
		log.Error(err, "Failed to find ConnectionPoint", "ConnectionPoint.Name", cpName)
		return nil, err
	}

	// Get one or more L2Service resources
	for _, l2srvName := range l2srvNames {
		tempObj := &enov1alpha1.L2Service{}
		if err := r.Get(ctx, types.NamespacedName{Name: l2srvName}, tempObj); err != nil {
			log.Error(err, "Failed to find L2Service", "L2Service.Name", l2srvName)
			return nil, err
		}
		l2srvObjs = append(l2srvObjs, tempObj)
	}

	var subnets []*v1alpha1.Subnet
	var routesMap map[string][]*v1alpha1.Route
	if s.Spec.VlanType == "access" {
		if len(l2srvObjs) > 1 {
			err := errors.New("number of L2Services for access vlan type cannot be more than 1")
			log.Error(err, "")
			return nil, err
		}
		var err error
		subnets, err = r.getSubnetObjs(ctx, l2srvObjs[0], log)
		if err != nil {
			log.Error(err, "Error occurred while fetching Subnet resources")
			return nil, err
		}

		//Get one or more route resources. Its an optional attribute.
		routesMap, err = r.getRouteObjs(ctx, subnets, log)
		if err != nil {
			log.Error(err, "Error occurred while fetching Route resources")
			return nil, err
		}
	}

	// Initiate L2ServiceAttachment Parser
	l2srvAttParser := l2serviceattachmentparser.NewL2SrvAttParser(s, l2srvObjs, cp, subnets, routesMap, r.Config, r.CniMap, r.IpamMap, log)
	// Parse the resources and fill the data
	cniManifestFile, ipamManifestFile, err := l2srvAttParser.ParseL2ServiceAttachment(data)
	if err != nil {
		log.Error(err, "Error occurred during parsing the L2ServiceAttachment")
		return nil, err
	}

	data["NetAttachDefName"] = s.Name
	data["NetAttachDefNamespace"] = s.Namespace
	if prefix, ok := data["ResourcePrefix"]; ok {
		data["NetResourceName"] = prefix.(string) + data["NetObjName"].(string)
	} else {
		data["NetResourceName"] = data["NetObjName"].(string)
	}

	cniFilePath := filepath.Join("manifests", "cni", cniManifestFile)
	cniConfig, err := getConfig(cniFilePath, data)

	if err != nil {
		return nil, err
	}
	log.Info("CNI template resolved:", "cni config", cniConfig)
	data["CNI"] = cniConfig

	if ipamManifestFile != "" {
		ipamFilePath := filepath.Join("manifests", "ipam", ipamManifestFile)
		ipamConfig, err := getConfig(ipamFilePath, data)
		if err != nil {
			return nil, err
		}
		log.Info("ipam template resolved:" + ipamConfig)
		data["IPAM"] = ipamConfig
	}

	nadFilePath := filepath.Join("manifests", "netattachdef", nadFileName)
	nad, err := getConfig(nadFilePath, data)
	if err != nil {
		return nil, err
	}
	log.Info("NAD resource template resolved", "NAD resource config", nad)
	obj, err := convertToUnstructured(nad)
	if err != nil {
		log.Error(err, "Failed to parse NAD resource yaml string")
		return nil, err
	}

	ctrl.SetControllerReference(s, obj, r.Scheme)
	return obj, nil

}

func convertToUnstructured(yamlString string) (*unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(yamlString), 4096)
	unstructured := unstructured.Unstructured{}
	if err := decoder.Decode(&unstructured); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("failed to unmarshal yaml: %v", err)
		}
	}
	return &unstructured, nil
}

func getConfig(templateFilePath string, data map[string]interface{}) (string, error) {
	configTemplate, err := template.ParseFiles(templateFilePath)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	configTemplate.Execute(buf, data)
	return buf.String(), nil
}

func (r *L2ServiceAttachmentReconciler) getSubnetObjs(ctx context.Context, l2srv *enov1alpha1.L2Service, log logr.Logger) ([]*enov1alpha1.Subnet, error) {
	var subnetObjs []*enov1alpha1.Subnet
	for _, subnetName := range l2srv.Spec.Subnets {
		tempObj := &enov1alpha1.Subnet{}
		if err := r.Get(ctx, types.NamespacedName{Name: subnetName}, tempObj); err != nil {
			log.Error(err, "Failed to find Subnet ", "subnetName: ", subnetName)
			return nil, err
		}
		subnetObjs = append(subnetObjs, tempObj)
	}
	return subnetObjs, nil
}

func (r *L2ServiceAttachmentReconciler) getRouteObjs(ctx context.Context, subnets []*enov1alpha1.Subnet, log logr.Logger) (map[string][]*enov1alpha1.Route, error) {
	routeObjs := make(map[string][]*enov1alpha1.Route)
	for _, subnet := range subnets {
		var tempObjs []*enov1alpha1.Route
		for _, routeName := range subnet.Spec.Routes {
			tempObj := &enov1alpha1.Route{}
			if err := r.Get(ctx, types.NamespacedName{Name: routeName}, tempObj); err != nil {
				log.Error(err, "Failed to find Route ", "routeName: ", routeName)
				return nil, err
			}
			tempObjs = append(tempObjs, tempObj)
		}
		routeObjs[subnet.GetName()] = tempObjs
	}
	return routeObjs, nil
}

// Check status of the L2Services so we can
// update the status of L2ServiceAttachment
func (r *L2ServiceAttachmentReconciler) CheckL2ServicesStatus(ctx context.Context, log logr.Logger, s *enov1alpha1.L2ServiceAttachment) (bool, error) {
	for _, ltwoSvcName := range s.Spec.L2Services {
		tempObj := &enov1alpha1.L2Service{}
		if err := r.Get(ctx, types.NamespacedName{Name: ltwoSvcName}, tempObj); err != nil {
			log.Error(err, "Failed to find L2Service", "L2Service.Name", ltwoSvcName)
			return false, err
		}
		if tempObj.Status.Phase == "ready" {
			if !common.SearchInSlice(s.Spec.ConnectionPoint, tempObj.Status.ConnectionPoints) {
				return true, nil
			}
		} else if tempObj.Status.Phase == "error" {
			err := errors.New("L2Service in error state")
			log.Error(err, "Failed L2Service", "L2Service.Name", ltwoSvcName)
			return false, err
		} else {
			return true, nil
		}
	}

	return false, nil
}

// L2Service Utils //

// DefineBridgeDomain - Create and returns a L2BridgeDomain resource
func (r *L2ServiceReconciler) DefineBridgeDomain(ctx context.Context, log logr.Logger, s *enov1alpha1.L2Service) (*enov1alpha1.L2BridgeDomain, error) {
	l2SvcAttList := &enov1alpha1.L2ServiceAttachmentList{}
	listOpts := []client.ListOption{}
	cpList := []string{}

	if err := r.List(ctx, l2SvcAttList, listOpts...); err != nil {
		log.Error(err, "Failed to list L2ServiceAttachments")
		return nil, err
	}

	for _, l2SvcAtt := range l2SvcAttList.Items {
		if common.SearchInSlice(s.Name, l2SvcAtt.Spec.L2Services) {
			if !common.SearchInSlice(l2SvcAtt.Spec.ConnectionPoint, cpList) {
				cpList = append(cpList, l2SvcAtt.Spec.ConnectionPoint)
			}
		}
	}

	brDom := &enov1alpha1.L2BridgeDomain{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
		},
		Spec: enov1alpha1.L2BridgeDomainSpec{
			Vlan:             s.Spec.SegmentationID,
			ConnectionPoints: cpList,
		},
	}

	// Set L2Service instance as the owner and controller
	// of L2BridgeDomain instance
	ctrl.SetControllerReference(s, brDom, r.Scheme)
	return brDom, nil
}

func (r *L2ServiceReconciler) DiffDesiredActual(desiredCPs, actualCPs []string) bool {
	exists := make(map[string]bool)

	if len(desiredCPs) != len(actualCPs) {
		return true
	}
	for _, cp := range desiredCPs {
		exists[cp] = true
	}
	for _, cp := range actualCPs {
		if !exists[cp] {
			return true
		}
	}
	return false
}

// UpdateStatus - Update the Status of L2Service instance
func (r *L2ServiceReconciler) UpdateStatus(ctx context.Context, log logr.Logger, statusCPs []string,
	phase, message string, lTwoSvc *enov1alpha1.L2Service) error {
	lTwoSvc.Status.ConnectionPoints = statusCPs
	lTwoSvc.Status.Phase = phase
	lTwoSvc.Status.Message = message
	if err := r.Status().Update(ctx, lTwoSvc); err != nil {
		log.Error(err, "Failed to update L2Service status")
		return err
	}
	return nil

}
