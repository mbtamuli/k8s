/*
Copyright 2021.

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
	"math/rand"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	demov1alpha1 "github.com/mbtamuli/hello-world/api/v1alpha1"
)

// CustomDeploymentReconciler reconciles a CustomDeployment object
type CustomDeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=demo.mriyam.dev,resources=customdeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=demo.mriyam.dev,resources=customdeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=demo.mriyam.dev,resources=customdeployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *CustomDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("customdeployment", req.NamespacedName)

	// Fetch the CustomDeployment instance
	deployment := &demov1alpha1.CustomDeployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("CustomDeployment resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CustomDeployment")
		return ctrl.Result{}, err
	}

	// Check if the required number of pods already exists, if not create new ones
	podList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(deployment.Namespace),
		client.MatchingLabels(labelsForCustomDeployment(deployment.Name)),
	}
	log.Info("Trying to List pods")
	if err = r.List(ctx, podList, opts...); err != nil {
		log.Info("testing")
		log.Error(err, "Failed to list pods", "CustomDeployment.Namespace", deployment.Namespace, "CustomDeployment.Name", deployment.Name)
		return ctrl.Result{}, err
	}

	// podNames := getPodNames(podList.Items)
	// for _, pod := range podNames {
	// 	log.Info(pod)
	// }

	// Ensure the CustomDeployment size is the same as the spec
	size := deployment.Spec.Replicas
	if len(podList.Items) != size {
		pod := r.getPodForCustomDeployment(deployment)
		log.Info(pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			log.Error(err, "Failed to create new Pod", "CustomDeployment.Namespace", pod.Namespace, "CustomDeployment.Name", pod.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	log.Info("Nothing happened?")
	return ctrl.Result{}, nil
}

func (r *CustomDeploymentReconciler) getPodForCustomDeployment(cd *demov1alpha1.CustomDeployment) *corev1.Pod {
	ls := labelsForCustomDeployment(cd.Name)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getRandomPodName(cd.Name),
			Namespace: cd.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image:   cd.Spec.Image,
				Name:    getRandomPodName(cd.Name),
				Command: []string{"sleep", "3600"},
			}},
		},
	}

	// Set CustomDeployment instance as the owner and controller
	ctrl.SetControllerReference(cd, pod, r.Scheme)
	return pod
}

// podName returns randomized name for pods of a given customdeployment CR name.
func getRandomPodName(name string) string {
	randomizedName := name + "-" + randomStringGenerator(5) + "-" + randomStringGenerator(5)
	return randomizedName
}

// labelsForCustomDeployment returns the labels for selecting the resources
// belonging to the given customdeployment CR name.
func labelsForCustomDeployment(name string) map[string]string {
	return map[string]string{"app": "customdeployment", "customdeployment_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// randomStringGenerator generates a string of fixed length
func randomStringGenerator(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	var src = rand.NewSource(time.Now().UnixNano())

	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1alpha1.CustomDeployment{}).
		Complete(r)
}
