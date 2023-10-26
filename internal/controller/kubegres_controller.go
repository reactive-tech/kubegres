/*
Copyright 2023 Reactive Tech Limited.
"Reactive Tech Limited" is a company located in England, United Kingdom.
https://www.reactive-tech.io

Lead Developer: Alex Arica

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
	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctx2 "reactive-tech.io/kubegres/internal/controller/ctx"
	"reactive-tech.io/kubegres/internal/controller/ctx/resources"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	kubegresv1 "reactive-tech.io/kubegres/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubegresReconciler reconciles a Kubegres object
type KubegresReconciler struct {
	client.Client
	Logger   logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kubegres.reactive-tech.io,resources=kubegres,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegres.reactive-tech.io,resources=kubegres/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegres.reactive-tech.io,resources=kubegres/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="batch",resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="storage.k8s.io",resources=storageclasses,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kubegres object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *KubegresReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = r.Logger.WithValues("kubegres", req.NamespacedName)

	r.Logger.Info("=======================================================")
	r.Logger.Info("=======================================================")
	kubegres, err := r.getDeployedKubegresResource(ctx, req)
	if err != nil {
		r.Logger.Info("Kubegres resource does not exist")
		return ctrl.Result{}, nil
	}

	resourcesContext, err := resources.CreateResourcesContext(kubegres, ctx, r.Logger, r.Client, r.Recorder)
	if err != nil {
		return ctrl.Result{}, err
	}

	nbreSecondsLeftBeforeTimeOut := resourcesContext.BlockingOperation.LoadActiveOperation()
	resourcesContext.BlockingOperationLogger.Log()
	resourcesContext.ResourcesStatesLogger.Log()

	if nbreSecondsLeftBeforeTimeOut > 0 {

		resultWithRequeue := ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(nbreSecondsLeftBeforeTimeOut) * time.Second,
		}
		return r.returnn(resultWithRequeue, nil, resourcesContext)
	}

	specCheckResult, err := resourcesContext.SpecChecker.CheckSpec()
	if err != nil {
		return r.returnn(ctrl.Result{}, err, resourcesContext)

	} else if specCheckResult.HasSpecFatalError {
		return r.returnn(ctrl.Result{}, nil, resourcesContext)
	}

	return r.returnn(ctrl.Result{}, r.enforceSpec(resourcesContext), resourcesContext)
}

func (r *KubegresReconciler) returnn(result ctrl.Result,
	err error,
	resourcesContext *resources.ResourcesContext) (ctrl.Result, error) {

	errStatusUpt := resourcesContext.KubegresContext.Status.UpdateStatusIfChanged()
	if errStatusUpt != nil && err == nil {
		return result, errStatusUpt
	}

	return result, err
}

func (r *KubegresReconciler) getDeployedKubegresResource(ctx context.Context, req ctrl.Request) (*kubegresv1.Kubegres, error) {

	// We sleep 1 second to let sufficient time to Kubernetes to update its system
	// so that when we will call the Get method below, we will receive the latest Kubegres resource
	time.Sleep(1 * time.Second)

	kubegres := &kubegresv1.Kubegres{}
	err := r.Client.Get(ctx, req.NamespacedName, kubegres)
	if err == nil {
		return kubegres, nil
	}

	r.Logger.Info("Kubegres resource does not exist")
	return &kubegresv1.Kubegres{}, err
}

func (r *KubegresReconciler) enforceSpec(resourcesContext *resources.ResourcesContext) error {

	err := r.enforceResourcesCountSpec(resourcesContext)
	if err != nil {
		return err
	}

	return r.enforceAllStatefulSetsSpec(resourcesContext)
}

func (r *KubegresReconciler) enforceResourcesCountSpec(resourcesContext *resources.ResourcesContext) error {
	return resourcesContext.ResourcesCountSpecEnforcer.EnforceSpec()
}

func (r *KubegresReconciler) enforceAllStatefulSetsSpec(resourcesContext *resources.ResourcesContext) error {
	return resourcesContext.AllStatefulSetsSpecEnforcer.EnforceSpec()
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubegresReconciler) SetupWithManager(mgr ctrl.Manager) error {

	ctx := context.Background()
	err := ctx2.CreateOwnerKeyIndexation(mgr, ctx)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kubegresv1.Kubegres{}).
		Owns(&apps.StatefulSet{}).
		Owns(&core.Service{}).
		Complete(r)
}
