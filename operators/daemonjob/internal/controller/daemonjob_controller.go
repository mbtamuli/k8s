package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	batchv1 "github.com/mbtamuli/k8s/daemonjob/api/v1"
)

// DaemonJobReconciler reconciles a DaemonJob object
type DaemonJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.mriyam.com,resources=daemonjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.mriyam.com,resources=daemonjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.mriyam.com,resources=daemonjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=list;watch
//+kubebuilder:rbac:groups=core,resources=pods,verbs=list;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DaemonJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling Daemonjob", "Namespace", req.Namespace, "Name", req.Name)

	var daemonjob batchv1.DaemonJob
	if err := r.Get(ctx, req.NamespacedName, &daemonjob); err != nil {
		return ctrl.Result{}, nil
	}

	err := DaemonSet(ctx, r.Client, &daemonjob, r.Scheme, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func DaemonSet(ctx context.Context, client client.Client, daemonjob *batchv1.DaemonJob, scheme *runtime.Scheme, logger logr.Logger) error {
	nodes := &corev1.NodeList{}
	err := client.List(ctx, nodes)
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		logger.Info("Node", "Name", node.Name)

		pod := getPodForNode(daemonjob, node.Name)

		if err := ctrl.SetControllerReference(daemonjob, pod, scheme); err != nil {
			logger.Error(err, "unable to set controller reference")
		}

		if err := client.Create(ctx, pod); err != nil {
			logger.Error(err, "Unable to create pod")

			return err
		}
	}

	return nil
}

func getPodForNode(daemonjob *batchv1.DaemonJob, nodeName string) *corev1.Pod {
	pod := &corev1.Pod{}
	pod.Name = fmt.Sprintf("pod-%s", nodeName)
	pod.Spec.NodeName = nodeName
	pod.ObjectMeta.Namespace = daemonjob.Namespace
	pod.Spec = *daemonjob.Spec.Template.Spec.DeepCopy()
	pod.ObjectMeta.Labels = daemonjob.Spec.Template.ObjectMeta.Labels

	return pod
}

// SetupWithManager sets up the controller with the Manager.
func (r *DaemonJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.DaemonJob{}).
		Complete(r)
}
