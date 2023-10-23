/*
Copyright 2023.

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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dummyv1alpha1 "github.com/marmar771/dummy-kubernetes-operator/api/v1alpha1"
)

const (
	namespace        = "default"
	podPendingStatus = "Pending"
	dummyFinalizer   = "dummy/finalizer"
	dummyKind        = "dummy"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.interview.com,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.interview.com,resources=dummies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// if request is not in the same namespace ignore it
	if req.Namespace != namespace {
		return ctrl.Result{}, nil
	}

	clientset, err := getClientset()
	if err != nil {
		panic(err)
	}

	object := &dummyv1alpha1.Dummy{}
	err = r.Get(ctx, types.NamespacedName{Name: dummyKind, Namespace: req.Namespace}, object)
	if err != nil {
		return ctrl.Result{}, nil
	}
	// when pod's status changes, update the dummy's status.podStatus field
	// if name is not dummy must get the pod and update the status
	// step 4 -> when pod's status changed update the status.podStatus
	if req.Name != dummyKind {
		existingPod, existingPodErr := clientset.CoreV1().Pods(namespace).Get(context.Background(), req.Name, metav1.GetOptions{})
		if existingPodErr == nil && existingPod.Status.Phase != podPendingStatus {
			object.Status.PodStatus = string(existingPod.Status.Phase)
			r.Status().Update(ctx, object)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// step 2 -> log name, namespace and spec's message
	l.Info("Incomming request: ",
		"Name: ", req.Name,
		"Namespace: ", req.Namespace,
		"Message: ", object.Spec.Message)

	// step 3 -> update status.specEcho
	if object.Spec.Message != object.Status.SpecEcho {
		object.Status.SpecEcho = object.Spec.Message
		r.Status().Update(ctx, object)
	}

	// add finalizer to handle API object deletation
	// step 4 -> delete pod when API object is deleted
	if object.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted,registering our finalizer.
		if !controllerutil.ContainsFinalizer(object, dummyFinalizer) {
			controllerutil.AddFinalizer(object, dummyFinalizer)
			if err := r.Update(ctx, object); err != nil {
				return ctrl.Result{}, err
			}

			l.Info("Finalizer is added")
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(object, dummyFinalizer) {

			err = clientset.CoreV1().Pods(namespace).Delete(context.Background(), string(object.GetUID()), metav1.DeleteOptions{})
			if err != nil {
				l.Info("Pod cannot be deleted", "Error", err)
			} else {
				l.Info("Pod is deleted")
			}

			controllerutil.RemoveFinalizer(object, dummyFinalizer)
			if err := r.Update(ctx, object); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// step 4 -> create a new pod, if existing update the status.podStatus property
	pod, podError := clientset.CoreV1().Pods(namespace).Get(context.Background(), string(object.GetUID()), metav1.GetOptions{})
	if podError != nil {
		desiredPod := createPodObject(string(object.GetUID()))
		pod, err = clientset.CoreV1().Pods(namespace).Create(context.Background(), desiredPod, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}

		l.Info("Pod is created", "pod", pod)
	}

	object.Status.PodStatus = string(pod.Status.Phase)
	r.Status().Update(ctx, object)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dummyv1alpha1.Dummy{}).
		Watches(&corev1.Pod{},
			&handler.EnqueueRequestForObject{}).
		Complete(r)
}

func getClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig :=
			clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}

func createPodObject(podName string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    map[string]string{"Name": podName}},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				corev1.Container{
					Name:  "main",
					Image: "nginx",
				},
			},
		},
	}
}
