package registry

import (
	"context"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"hypercloud-operator-go/pkg/controller/regctl"
	"reflect"

	"github.com/operator-framework/operator-sdk/pkg/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_registry")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Registry Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRegistry{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("registry-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Registry
	err = c.Watch(&source.Kind{Type: &regv1.Registry{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Registry
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &regv1.Registry{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &regv1.Registry{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRegistry implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRegistry{}

// ReconcileRegistry reconciles a Registry object
type ReconcileRegistry struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Registry object and makes changes based on the state read
// and what is in the Registry.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRegistry) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Registry")

	// [TODO] Compare spec with annotation spec

	// Fetch the Registry reg
	reg := &regv1.Registry{}
	err := r.client.Get(context.TODO(), request.NamespacedName, reg)
	if err != nil {
		reqLogger.Info("Error on get registry")
		if errors.IsNotFound(err) {
			reqLogger.Info("Not Found Error")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if regctl.UpdateRegistryStatus(r.client, reg) {
		return reconcile.Result{}, nil
	}

	if err = r.createAllSubresources(reg); err != nil {
		reqLogger.Error(err, "Subresource creation failed")
		return reconcile.Result{}, err
	}

	// [TODO] Store spec in annotation spec
	return reconcile.Result{}, nil
}

func (r *ReconcileRegistry) createAllSubresources(reg *regv1.Registry) error { // if want to requeue, return true
	subResourceLogger := log.WithValues("SubResource.Namespace", reg.Namespace, "SubResource.Name", reg.Name)
	subResourceLogger.Info("Creating all Subresources")
	for _, subresource := range collectSubresources() {
		subresourceType := reflect.TypeOf(subresource).String()
		subResourceLogger.Info("Check subresource", "subresourceType", subresourceType)
		registryCondition := &status.Condition{
			Status: corev1.ConditionFalse,
			Type:   status.ConditionType(subresource.GetTypeName()),
		}

		if err := subresource.Create(r.client, reg, registryCondition, r.scheme, true); err != nil {
			subResourceLogger.Info("Got Error in creating subresource ")
			subresource.StatusPatch(r.client, reg, registryCondition, true)
			return err
		}

		err := subresource.Ready(reg, true)
		if err != nil && err.Error() == regv1.NotReady {
			registryCondition.Status = corev1.ConditionFalse
			subresource.StatusPatch(r.client, reg, registryCondition, false)
			return err
		} else {
			registryCondition.Status = corev1.ConditionTrue
			subresource.StatusPatch(r.client, reg, registryCondition, false)
		}
	}

	return nil
}

func collectSubresources() []regctl.RegistrySubresource {
	collection := []regctl.RegistrySubresource{}
	// [TODO] Add Subresources in here
	// collection = append(collection, &regctl.RegistryService{})
	collection = append(collection, &regctl.RegistryPVC{}, &regctl.RegistryService{})
	return collection
}
