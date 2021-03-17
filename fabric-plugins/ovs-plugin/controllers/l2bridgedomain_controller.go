/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	//"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	enofabricv1alpha1 "github.com/Nordix/eno/fabric-plugins/ovs-plugin/api/v1alpha1"
	"github.com/Nordix/eno/fabric-plugins/ovs-plugin/pkg/config"
	"github.com/Nordix/eno/fabric-plugins/ovs-plugin/pkg/ovs"
)

// L2BridgeDomainReconciler reconciles a L2BridgeDomain object
type L2BridgeDomainReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Config    *config.Configuration
	OvsClient *ovs.Client
}

// +kubebuilder:rbac:groups=eno.k8s.io,resources=l2bridgedomains,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eno.k8s.io,resources=l2bridgedomains/status,verbs=get;update;patch

func (r *L2BridgeDomainReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("l2bridgedomain", req.NamespacedName)

	// Fetch the L2BridgeDomain instance
	brDom := &enofabricv1alpha1.L2BridgeDomain{}
	err := r.Get(ctx, req.NamespacedName, brDom)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			// ToDO: Here for the L2BridgeDomain deletion we need finalizer logic
			log.Info("L2BridgeDomain resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get L2BridgeDomain")
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	//desiredPorts, err := r.CreateDesiredState(ctx, log, brDom)
	//actualPorts, err := r.GetActualState(brDom)
	//fmt.Printf("%+v\n", desiredPorts)
	//fmt.Printf("%+v\n", actualPorts)
	_ = r.Apply(ctx, log, brDom)

	return ctrl.Result{}, nil
}

func (r *L2BridgeDomainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&enofabricv1alpha1.L2BridgeDomain{}).
		Complete(r)
}
