/*
Copyright 2025.

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
	"github.com/bitcoin-sv/cars-operator/internal/utils"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/bitcoin-sv/cars-operator/api/v1alpha1"
)

// CarsReconciler reconciles a Cars object
type CarsReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Log            logr.Logger
	NamespacedName types.NamespacedName
	Context        context.Context
}

//+kubebuilder:rbac:groups=infra.bsvblockchain.com,resources=cars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infra.bsvblockchain.com,resources=cars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infra.bsvblockchain.com,resources=cars/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=endpoints;configmaps;services;secrets;persistentvolumeclaims,verbs=get;create;update;list;watch
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;update;create;list;watch
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;update;create;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *CarsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result := ctrl.Result{}
	r.Log = log.FromContext(ctx).WithValues("cars", req.NamespacedName)
	r.Context = ctx
	r.NamespacedName = req.NamespacedName

	cars := infrav1alpha1.Cars{}
	if err := r.Get(ctx, req.NamespacedName, &cars); err != nil {
		r.Log.Error(err, "unable to fetch CARS CR")
		return result, nil
	}

	_, err := utils.ReconcileBatch(r.Log,
		r.ReconcileDeployment,
		r.ReconcileService,
		r.ReconcileIngress,
		r.ReconcileMysqlDeployment,
		r.ReconcileMysqlService,
		r.ReconcileMysqlPVC,
	)

	if err != nil {
		apimeta.SetStatusCondition(&cars.Status.Conditions,
			metav1.Condition{
				Type:    infrav1alpha1.ConditionReconciled,
				Status:  metav1.ConditionFalse,
				Reason:  infrav1alpha1.ReconciledReasonError,
				Message: err.Error(),
			},
		)
		_ = r.Client.Status().Update(ctx, &cars)
		// Since error is written on the status, let's log it and requeue
		// Returning error here is redundant
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	} else {
		apimeta.SetStatusCondition(&cars.Status.Conditions,
			metav1.Condition{
				Type:    infrav1alpha1.ConditionReconciled,
				Status:  metav1.ConditionTrue,
				Reason:  infrav1alpha1.ReconciledReasonComplete,
				Message: infrav1alpha1.ReconcileCompleteMessage,
			},
		)
	}
	err = r.Client.Status().Update(ctx, &cars)

	return ctrl.Result{Requeue: false, RequeueAfter: 0}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *CarsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.Cars{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// getAppLabels defines the label applied to created resources. This label is used by the predicate to determine which resources are ours
func getAppLabels() map[string]string {
	return map[string]string{
		infrav1alpha1.CarsLabel: "true",
	}
}
