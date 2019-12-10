package myapp

import (
	"context"
	"fmt"

	myappv1alpha1 "github.com/brunovlucena/myapp-operator/pkg/apis/myapp/v1alpha1"
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
)

var log = logf.Log.WithName("controller_myapp")

// Add creates a new MyApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMyApp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("myapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MyApp
	err = c.Watch(&source.Kind{Type: &myappv1alpha1.MyApp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner MyApp
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &myappv1alpha1.MyApp{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMyApp implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMyApp{}

// ReconcileMyApp reconciles a MyApp object
type ReconcileMyApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MyApp object and makes changes based on the state read
// and what is in the MyApp.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMyApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MyApp")

	// Fetch the MyApp instance
	instance := &myappv1alpha1.MyApp{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set MyApp instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
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

func getLabelsForCR(name string) map[string]string {
	return map[string]string{
		"app":                          name,
		"app.kubernetes.io/managed-by": "myapp-operator",
		"app.kubernetes.io/version":    "v0.0.1",
	}
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *myappv1alpha1.MyApp) *corev1.Pod {
	labels := getLabelsForCR(cr.Name)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "myapp",
					Image:   "localhost:5000/myapp:v0.0.1",
					Command: []string{"./main"},
					Ports: []corev1.ContainerPort{{
						Name:          "http",
						ContainerPort: cr.Spec.ContainerPort,
					}},
					Env: []corev1.EnvVar{
						{Name: "API_CONTAINER_PORT", Value: fmt.Sprint(cr.Spec.ContainerPort)},
						{Name: "DATABASE_TYPE", Value: cr.Spec.DatabaseType},
						{Name: "DATABASE_HOST", Value: cr.Spec.Host},
						{Name: "DATABASE_PORT", Value: fmt.Sprint(cr.Spec.Port)},
						{Name: "DATABASE_USER", Value: cr.Spec.User},
						{Name: "DATABASE_PASS", Value: cr.Spec.Pass},
						{Name: "DATABASE_NAME", Value: cr.Spec.DatabaseName},
					},
				},
			},
		},
	}
}
