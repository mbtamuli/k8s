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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonJobSpec defines the desired state of DaemonJob
type DaemonJobSpec struct {
	// Optional number of retries before marking this job failed.
	// Defaults to 6
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// Describes the pod that will be created when executing a job.
	Template corev1.PodTemplateSpec `json:"template"`
}

// DaemonJobStatus defines the observed state of DaemonJob
type DaemonJobStatus struct {
	// Represents time when the job controller started processing a job. This field is reset every time a Job is resumed
	// from suspension. It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// The completion time is only set when the job finishes successfully.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// The number of actively running pods.
	// +optional
	Active int32 `json:"active,omitempty"`

	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32 `json:"succeeded,omitempty"`

	// The number of pods which reached phase Failed.
	// +optional
	Failed int32 `json:"failed,omitempty"`
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
