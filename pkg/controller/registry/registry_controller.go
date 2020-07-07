package registry

import (
	"bytes"
	"context"
	"os"

	tmaxv1 "hypercloud-operator-go/pkg/apis/tmax/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	// [TODO] Change into public repo
	regv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
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
	err = c.Watch(&source.Kind{Type: &tmaxv1.Registry{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Registry
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.Registry{},
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
	reg := &tmaxv1.Registry{}
	err := r.client.Get(context.TODO(), request.NamespacedName, reg)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//certLogger := log.WithValues("Certification Log")
	//registryDir := createDirectory(reg.Namespace, reg.Name)

	//certificateCmd := createCertificateCmd(registryDir)

	// Define a new Pod object
	pod := newPodForCR(reg)

	// Set Registry reg as the owner and controller
	if err := controllerutil.SetControllerReference(reg, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

func createDirectory(domainId string, registryId string) string {
	certLogger := log.WithValues("Create Directory")
	certLogger.Info("Create Cert Directory")

	if _, err := os.Stat(regv1.OpenSslHomeDir); os.IsNotExist(err) {
		// [TODO] Mode should be managed
		os.Mkdir(regv1.OpenSslHomeDir, 0777)
		certLogger.Info("Directory created : " + regv1.OpenSslHomeDir)
	}

	domainDir := regv1.OpenSslHomeDir + "/" + domainId
	if _, err := os.Stat(domainDir); os.IsNotExist(err) {
		// [TODO] Mode should be managed
		os.Mkdir(domainDir, 0777)
		certLogger.Info("Directory created : " + domainDir)
	}

	registryDir := domainDir + "/" + registryId
	if _, err := os.Stat(registryDir); os.IsNotExist(err) {
		// [TODO] Mode should be managed
		os.Mkdir(registryDir, 0777)
		certLogger.Info("Directory created : " + registryDir)
	}

	dockerLoginHome := regv1.DockerLoginHomeDir
	if _, err := os.Stat(dockerLoginHome); os.IsNotExist(err) {
		// [TODO] Mode should be managed
		os.Mkdir(dockerLoginHome, 0777)
		certLogger.Info("Directory created : " + dockerLoginHome)
	}

	return registryDir
}

func createCertificateCmd(registryDir string, clusterIP string) string {
	// For Efficiency
	var buffer bytes.Buffer

	buffer.WriteString("openssl req -newkey rsa:4096 -nodes -sha256 ")
	buffer.WriteString("-keyout " + registryDir + "/" + regv1.CertKeyFile + " ")
	buffer.WriteString("-x509 -days 1000 ")
	buffer.WriteString("-subj \"/C=KR/ST=Seoul/O=tmax/CN=" + clusterIP + "\" ")                                                                       // [TODO]
	buffer.WriteString("-config <(cat /etc/ssl/openssl.cnf <(printf \"[v3_ca]\\nsubjectAltName=IP:" + clusterIP + "," + serviceTypeSubject + "\")) ") // [TODO]
	buffer.WriteString("-out " + registryDir + "/" + regv1.CertCrtFile)

	return buffer.String()
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *tmaxv1.Registry) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
