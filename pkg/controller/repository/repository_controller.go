package repository

import (
	"context"
	"strings"

	"hypercloud-operator-go/internal/utils"
	tmaxv1 "hypercloud-operator-go/pkg/apis/tmax/v1"
	"hypercloud-operator-go/pkg/controller/regctl"
	regApi "hypercloud-operator-go/pkg/registry"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_repository")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Repository Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRepository{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("repository-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Repository
	err = c.Watch(&source.Kind{Type: &tmaxv1.Repository{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Repository
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.Repository{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRepository implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRepository{}

// ReconcileRepository reconciles a Repository object
type ReconcileRepository struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Repository object and makes changes based on the state read
// and what is in the Repository.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRepository) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Repository")

	// Fetch the Repository instance
	repo := &tmaxv1.Repository{}
	err := r.client.Get(context.TODO(), request.NamespacedName, repo)
	if err != nil {
		if errors.IsNotFound(err) {
			reg, err := getRegistryByRequest(r.client, request)
			if err != nil {
				reqLogger.Error(err, "")
				return reconcile.Result{}, err
			}

			repoName, _ := splitRepoCRName(request.Name)
			if err := sweepRegistryRepo(r.client, reg, repoName); err != nil {
				reqLogger.Error(err, "")
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reg, err := getRegistryByRequest(r.client, request)
	if err != nil {
		reqLogger.Error(err, "")
		return reconcile.Result{}, err
	}

	if err := sweepImages(r.client, reg, repo); err != nil {
		reqLogger.Error(err, "")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func patchRepo(c client.Client, reg *tmaxv1.Registry, repo *tmaxv1.Repository, deletedTags []string) error {
	repoctl := &regctl.RegistryRepository{}
	patchRepo := repo.DeepCopy()
	patchImageList := []tmaxv1.ImageVersion{}

	for _, image := range repo.Spec.Versions {
		if !utils.Contains(deletedTags, image.Version) {
			patchImageList = append(patchImageList, image)
		}
	}

	patchRepo.Spec.Versions = patchImageList
	if err := repoctl.Patch(c, repo, patchRepo); err != nil {
		return err
	}

	return nil
}

func sweepImages(c client.Client, reg *tmaxv1.Registry, repo *tmaxv1.Repository) error {
	repoName := repo.Spec.Name
	images := repo.Spec.Versions
	delTags := []string{}

	for _, image := range images {
		if image.Delete {
			delTags = append(delTags, image.Version)
		}
	}

	deletedTags, err := deleteImagesInRepo(c, reg, repoName, delTags)
	if err != nil {
		return err
	}

	if err := patchRepo(c, reg, repo, deletedTags); err != nil {
		return err
	}

	if err := garbageCollect(c, reg); err != nil {
		return err
	}

	return nil
}

func sweepRegistryRepo(c client.Client, reg *tmaxv1.Registry, repoName string) error {
	ra := regApi.NewRegistryApi(reg)

	tags := ra.Tags(repoName).Tags
	if tags == nil {
		return nil
	}
	if _, err := deleteImagesInRepo(c, reg, repoName, ra.Tags(repoName).Tags); err != nil {
		return err
	}

	if err := garbageCollect(c, reg); err != nil {
		return err
	}

	return nil
}

func deleteImagesInRepo(c client.Client, reg *tmaxv1.Registry, repoName string, tags []string) ([]string, error) {
	ra := regApi.NewRegistryApi(reg)
	deletedTags := []string{}

	for _, tag := range tags {
		digest, err := ra.DockerContentDigest(repoName, tag)
		if err != nil {
			log.Error(err, "")
			return deletedTags, err
		}

		if err := ra.DeleteManifest(repoName, digest); err != nil {
			log.Error(err, "")
			return deletedTags, err
		}
		deletedTags = append(deletedTags, tag)
	}

	return deletedTags, nil
}

func garbageCollect(c client.Client, reg *tmaxv1.Registry) error {
	podCtl := &regctl.RegistryPod{}

	podName, err := podCtl.PodName(c, reg)
	if err != nil {
		return err
	}

	cmder := regApi.NewCommander(podName, reg.Namespace)
	out, err := cmder.GarbageCollect()
	if err != nil {
		log.Error(err, "exec")
		return err
	}

	log.Info("exec", "stdout", out.Outbuf.String(), "stderr", out.Errbuf.String())
	return nil
}

func getRegistryByRequest(c client.Client, request reconcile.Request) (*tmaxv1.Registry, error) {
	registry := &tmaxv1.Registry{}
	name, _ := splitRepoCRName(request.Name)
	namespace := request.Namespace

	err := c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func splitRepoCRName(crName string) (repoName, regName string) {
	parts := strings.Split(crName, ".")

	repoName = strings.Join(parts[:len(parts)-1], ".")
	regName = parts[len(parts)-1]

	return
}
