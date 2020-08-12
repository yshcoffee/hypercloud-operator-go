package registry

import (
	"context"
	"hypercloud-operator-go/pkg/controller/regctl"
	"reflect"

	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	//"hypercloud-operator-go/pkg/controller/regctl"

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

	// Watch Registry Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &regv1.Registry{},
	})
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

	// Fetch the Registry reg
	reg := &regv1.Registry{}
	err := r.client.Get(context.TODO(), request.NamespacedName, reg)
	if err != nil {
		reqLogger.Info("Error on get registry")
		if errors.IsNotFound(err) {
			reqLogger.Info("Not Found Error")
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = createAllSubresources(r.client, reg)
	if err != nil {
		reqLogger.Info("Subresource creation failed")
		return reconcile.Result{}, nil
	}
	/*
		svc := model.RegistryService{}
		pvc := model.RegistryPVC{}
		subreses := []model.RegistrySubresource{&svc, &pvc}

		for _, res := range subreses {
			err = res.Get(r.client, reg)
			if err != nil {
				if errors.IsNotFound(err) {
					res.Create(r.client, reg)
				} else {

					return reconcile.Result{}, err
				}
			}

			if !res.Ready(reg) {
				res.StatusUpdate(r.client, reg)
			} else {
				// res.StatusUpdate(r.client, reg)
			}
		}
	*/

	// svc := &corev1.Service{}
	// err = r.client.Get(context.TODO(), request.NamespacedName, svc)
	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		// Request object not found, could have been deleted after reconcile request.
	// 		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
	// 		// Return and don't requeue
	// 		return reconcile.Result{}, nil
	// 	}
	// 	// Error reading the object - requeue the request.
	// 	return reconcile.Result{}, err
	// }

	// // certLogger := log.WithValues("Certification Log")
	// //registryDir := createDirectory(reg.Namespace, reg.Name)

	return reconcile.Result{}, nil
}
func createAllSubresources(client client.Client, reg *regv1.Registry) error {
	// [TODO] Set Owner Reference for all sub resources
	subResourceLogger := log.WithValues("Request.Namespace", reg.Namespace, "Request.Name", reg.Name)
	subResourceLogger.Info("Making subresources")
	for _, subresource := range collectSubresources() {
		subresourceType := reflect.TypeOf(subresource).String()
		subResourceLogger.Info("Check subresource", subresourceType)
		registryCondition := &status.Condition{
			Status: corev1.ConditionFalse,
			Type:   status.ConditionType(subresource.GetTypeName()),
		}
		if err := subresource.Get(client, reg, registryCondition); err != nil {
			if errors.IsNotFound(err) {
				subResourceLogger.Info("Create subresource", subresourceType)
				subresource.Create(client, reg, registryCondition)
			} else {
				subResourceLogger.Info("Got Error in getting subresource ", subresourceType)
				return err
			}
		}
		if subresource.Ready(reg) {
			subresource.StatusPatch(client, reg, registryCondition)
		}
	}

	return nil
}

func collectSubresources() []regctl.RegistrySubresource {
	collection := []regctl.RegistrySubresource{}
	// [TODO] Add Subresources in here
	collection = append(collection, &regctl.RegistryService{})
	collection = append(collection, &regctl.RegistryPVC{})
	return collection
}
