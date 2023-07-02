package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonJobSpec defines the desired state of DaemonJob
type DaemonJobSpec struct {
	Template corev1.PodTemplateSpec `json:"template,omitempty"`
}

// DaemonJobStatus defines the observed state of DaemonJob
type DaemonJobStatus struct {
	DesiredNumberScheduled int32 `json:"desiredNumberScheduled"`
	CurrentNumberScheduled int32 `json:"currentNumberScheduled"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DaemonJob is the Schema for the daemonjobs API
type DaemonJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DaemonJobSpec   `json:"spec,omitempty"`
	Status DaemonJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DaemonJobList contains a list of DaemonJob
type DaemonJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DaemonJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DaemonJob{}, &DaemonJobList{})
}
