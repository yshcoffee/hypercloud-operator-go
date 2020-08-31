package registry

import (
	"bytes"
	"context"
	"encoding/json"
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"hypercloud-operator-go/pkg/controller/regctl"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
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

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &regv1.Registry{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &regv1.Registry{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
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

	var requeueErr error = nil
	collectSubController := collectSubController(reg.Spec.RegistryService.ServiceType)
	patchReg := reg.DeepCopy() // Target to Patch object

	defer r.patch(reg, patchReg)


	// Check if subresources are created.
	for _, sctl := range collectSubController {
		subresourceType := reflect.TypeOf(sctl).String()
		subResourceLogger.Info("Check subresource", "subresourceType", subresourceType)

		// Check if subresource is Created.
		if err := sctl.Create(r.client, reg, patchReg, r.scheme, true); err != nil {
			subResourceLogger.Error(err, "Got an error in creating subresource ")
			return err
		}

		// Check if subresource is ready.
		if err := sctl.Ready(r.client, reg, patchReg, false); err != nil {
			subResourceLogger.Error(err, "Got an error in checking ready")
			if regv1.IsPodError(err) {
				requeueErr = err
			} else {
				return err
			}
		}
	}
	if requeueErr != nil {
		return requeueErr
	}

	return nil
}

func updateAllSubresources(reg *regv1.Registry) bool {
	if reg.Status.Phase != string(regv1.StatusRunning) {
		return false
	}

	// lastRegSpec := reg.Status.LastAppliedSpec
	// curRegSpec := reg.Spec
	// opts := jsondiff.DefaultJSONOptions()
	// jsondiff.Compare(lastRegSpec, curRegSpec)

	return true
}

func (r *ReconcileRegistry) patch(origin, target *regv1.Registry) error {
	subResourceLogger := log.WithValues("SubResource.Namespace", origin.Namespace, "SubResource.Name", origin.Name)

	originObject := client.MergeFrom(origin) // Set original obeject
	statusPatchTarget := target.DeepCopy()

	// Get origin data except status for compare
	originWithoutStatus := origin.DeepCopy()
	originWithoutStatus.Status = regv1.RegistryStatus{}
	originWithoutStatusByte, err := json.Marshal(*originWithoutStatus)
	if err != nil {
		subResourceLogger.Error(err, "json marshal error")
		return err
	}

	// Get target data except status for compare
	targetWithoutStatus := target.DeepCopy()
	targetWithoutStatus.Status = regv1.RegistryStatus{}
	targetWithoutStatusByte, err := json.Marshal(*targetWithoutStatus)
	if err != nil {
		subResourceLogger.Error(err, "json marshal error")
		return err
	}

	// Check whether patch is necessary or not
	if res := bytes.Compare(originWithoutStatusByte, targetWithoutStatusByte); res != 0 {
		subResourceLogger.Info("Patch registry")
		if err := r.client.Patch(context.TODO(), target, originObject); err != nil {
			subResourceLogger.Error(err, "Unknown error patching status")
			return err
		}
	}

	// Get origin status data for compare
	originStatus := origin.Status.DeepCopy()
	originStatusByte, err := json.Marshal(*originStatus)
	if err != nil {
		subResourceLogger.Error(err, "json marshal error")
		return err
	}

	// Get target status data for compare
	targetStatusByte, err := json.Marshal(*statusPatchTarget)
	if err != nil {
		subResourceLogger.Error(err, "json marshal error")
		return err
	}

	// Check whether patch is necessary or not about status
	if res := bytes.Compare(originStatusByte, targetStatusByte); res != 0 {
		subResourceLogger.Info("Patch registry status")
		if err := r.client.Status().Patch(context.TODO(), statusPatchTarget, originObject); err != nil {
			subResourceLogger.Error(err, "Unknown error patching status")
			return err
		}
	}

	return nil
}

func collectSubController(serviceType regv1.RegistryServiceType) []regctl.RegistrySubresource {
	collection := []regctl.RegistrySubresource{}
	// [TODO] Add Subresources in here
	collection = append(collection, &regctl.RegistryPVC{}, &regctl.RegistryService{}, &regctl.RegistryCertSecret{},
		&regctl.RegistryDCJSecret{}, &regctl.RegistryConfigMap{}, &regctl.RegistryDeployment{}, &regctl.RegistryPod{})
	if serviceType == "Ingress" {
		collection = append(collection, &regctl.RegistryIngress{})
	}
	return collection
}
