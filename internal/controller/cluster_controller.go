/*
Copyright 2024.

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

package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1 "sempex/pf-talos-operator/api/v1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var logger = ctrl.Log.WithName("cluster_controller")

func checkCertificate() (certificate int, err error) {
	resp, err := http.Get("http://www.randomnumberapi.com/api/v1.0/random?min=100&max=110&count=1")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var number []int
	err = json.Unmarshal(body, &number)
	if err != nil {
		return 0, err
	}
	certificate = number[0]
	return certificate, nil
}

//+kubebuilder:rbac:groups=cluster.sempex,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.sempex,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cluster.sempex,resources=clusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	var cluster clusterv1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		logger.Error(err, "unable to fetch Cluster")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	newCertificate, err := checkCertificate()
	if err != nil {
		logger.Error(err, "failed to check for a new certificate")
		return ctrl.Result{}, err
	}

	if cluster.Spec.Certificate != newCertificate {
		logger.Info("Updating certificate", "old", cluster.Spec.Certificate, "new", newCertificate)
		cluster.Spec.Certificate = newCertificate

		// Update the Cluster resource
		if err := r.Update(ctx, &cluster); err != nil {
			logger.Error(err, "failed to update Cluster certificate")
			return ctrl.Result{}, err
		}
	} else {
		logger.Info("No update required for the certificate")
	}

	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil // This should be outside of the else block
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.Cluster{}).
		Complete(r)
}
