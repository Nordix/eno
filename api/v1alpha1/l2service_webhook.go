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

package v1alpha1

import (
	"fmt"
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// log is for logging in this package.
var (
	l2servicelog = logf.Log.WithName("l2service-resource")
	C	     client.Client
)

func (r *L2Service) SetupWebhookWithManager(mgr ctrl.Manager) error {
	C = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

func (r *L2Service) checkSegmentationIdUniqueness() error {
	var l2SvcList L2ServiceList
        _ = C.List(context.TODO(), &l2SvcList)
        for _, l2Svc := range l2SvcList.Items {
                if r.Spec.SegmentationID == l2Svc.Spec.SegmentationID {
                        err := fmt.Errorf("L2Service with Segmentation ID %v already exists", r.Spec.SegmentationID)
                        return err
                }
        }

	return nil
}

func (r *L2Service) checkL2ServiceUsageByL2SvcAtt() error {
	var l2SvcAttList L2ServiceAttachmentList
        _ = C.List(context.TODO(), &l2SvcAttList)
        for _, l2SvcAtt := range l2SvcAttList.Items {
                for _, l2SvcName := range l2SvcAtt.Spec.L2Services {
                        if r.Name == l2SvcName {
                                err := fmt.Errorf("You can not update/delete L2Service %s while it is" +
                                                        " in use by one or more L2ServiceAttachments", r.Name)
                                return err
                        }
                }
        }

        return nil
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-eno-k8s-io-v1alpha1-l2service,mutating=false,failurePolicy=fail,groups=eno.k8s.io,resources=l2services,versions=v1alpha1,name=vl2service.kb.io

var _ webhook.Validator = &L2Service{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *L2Service) ValidateCreate() error {
	l2servicelog.Info("validate create", "name", r.Name)

	err := r.checkSegmentationIdUniqueness()
	if err != nil {
		l2servicelog.Error(err, "L2Service validate create failed", "name", r.Name)
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *L2Service) ValidateUpdate(old runtime.Object) error {
	l2servicelog.Info("validate update", "name", r.Name)

	if err := r.checkL2ServiceUsageByL2SvcAtt();err != nil {
		l2servicelog.Error(err, "L2Service validate update failed", "name", r.Name)
		return err
	}

	if err := r.checkSegmentationIdUniqueness();err != nil {
                l2servicelog.Error(err, "L2Service validate update failed", "name", r.Name)
                return err
        }
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *L2Service) ValidateDelete() error {
	l2servicelog.Info("validate delete", "name", r.Name)

	if err := r.checkL2ServiceUsageByL2SvcAtt();err != nil {
		l2servicelog.Error(err, "L2Service validate delete failed", "name", r.Name)
		return err
	}

	return nil
}
