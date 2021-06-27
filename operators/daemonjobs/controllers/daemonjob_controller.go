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
	"context"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appv1alpha1 "github.com/mbtamuli/k8s/operators/daemonjobs/api/v1alpha1"
)

// DaemonJobReconciler reconciles a DaemonJob object
type DaemonJobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.mriyam.com,resources=daemonjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.mriyam.com,resources=daemonjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.mriyam.com,resources=daemonjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DaemonJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DaemonJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("daemonjob", req.NamespacedName)

	reqLogger.Info("Reconciling DaemonJob...")

	// Fetch the DaemonJob instance
	dj := &appv1alpha1.DaemonJob{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, dj)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	reqLogger.Info("Get pods for the current DaemonJob...")
	pods, err := r.getPodsForJob(dj, req.Namespace, reqLogger)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, p := range pods {
		reqLogger.Info("pod", "namespace", p.Namespace, "name", p.Name, "state ", p.Status.Phase)
	}

	activePods := filterActivePods(pods, reqLogger)
	active := int32(len(activePods))
	succeeded, failed := getStatus(pods)

	status := appv1alpha1.DaemonJobStatus{
		Active:    active,
		Succeeded: succeeded,
		Failed:    failed,
	}
	if !reflect.DeepEqual(dj.Status, status) {
		dj.Status = status
		err := r.Client.Status().Update(context.TODO(), dj)
		if err != nil {
			reqLogger.Error(err, "failed to update the dj")
			return ctrl.Result{}, err
		}
	}

	if len(pods) <= 0 {
		reqLogger.Info("No pods found for the current DaemonJob, creating a new one...")
		pod := newPodForDaemonJob(dj)
		pod.Spec = *dj.Spec.Template.Spec.DeepCopy()
		if err := controllerutil.SetControllerReference(dj, pod, r.Scheme); err != nil {
			reqLogger.Error(err, "unable to set owner reference on new Workflow")
			return ctrl.Result{}, err
		}
		reqLogger.Info("new pod", "name", pod.ObjectMeta.Name, "spec", pod.Spec)
		err = r.Client.Create(context.TODO(), pod)
		if err != nil {
			reqLogger.Error(err, "failed to create a pod")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{Requeue: true}, nil
}

// newPodForDaemonJob returns a new pod with the same name/namespace as the DaemonJob
func newPodForDaemonJob(dj *appv1alpha1.DaemonJob) *corev1.Pod {
	labels := map[string]string{
		"app":     dj.Name,
		"version": "v0.1",
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: dj.Name + "-pod",
			Namespace:    dj.Namespace,
			Labels:       labels,
		},
	}
}

// getPodsForJob returns the set of pods that this Job should manage.
// It also reconciles ControllerRef by adopting/orphaning.
// Note that the returned Pods are pointers into the cache.
func (r *DaemonJobReconciler) getPodsForJob(dj *appv1alpha1.DaemonJob, namespace string, reqLogger logr.Logger) ([]*corev1.Pod, error) {
	var pods []*corev1.Pod
	// List all pods owned by this PodSet instance
	lbls := labels.Set{
		"app":     dj.Name,
		"version": "v0.1",
	}
	podList := &corev1.PodList{}
	err := r.Client.List(context.TODO(),
		podList,
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		})
	if err != nil {
		reqLogger.Error(err, "failed to list existing pods in the podSet")
		return nil, err
	}
	for _, pod := range podList.Items {
		pods = append(pods, &pod)
	}
	return pods, nil
}

// filterActivePods returns pods that have not terminated.
func filterActivePods(pods []*corev1.Pod, reqLogger logr.Logger) []*corev1.Pod {
	var result []*corev1.Pod
	for _, p := range pods {
		if isPodActive(p) {
			result = append(result, p)
		} else {
			reqLogger.Info("Ignoring inactive pod", "namespace", p.Namespace, "name", p.Name, "state ", p.Status.Phase, "deletion time", p.DeletionTimestamp)
		}
	}
	return result
}

func isPodActive(p *corev1.Pod) bool {
	return corev1.PodSucceeded != p.Status.Phase &&
		corev1.PodFailed != p.Status.Phase &&
		p.DeletionTimestamp == nil
}

// getStatus returns no of succeeded and failed pods running a job
func getStatus(pods []*corev1.Pod) (succeeded, failed int32) {
	succeeded = int32(countPodsByPhase(pods, corev1.PodSucceeded))
	failed = int32(countPodsByPhase(pods, corev1.PodFailed))
	return
}

// countPodsByPhase returns pods based on their phase.
func countPodsByPhase(pods []*corev1.Pod, phase corev1.PodPhase) int {
	result := 0
	for _, p := range pods {
		if phase == p.Status.Phase {
			result++
		}
	}
	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *DaemonJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.DaemonJob{}).
		Complete(r)
}
