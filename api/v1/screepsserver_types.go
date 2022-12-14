/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type PullRequest struct {
	Repo  string `json:"repo,omitempty"`
	Issue int    `json:"issue,omitempty"`
}

// ScreepsServerSpec defines the desired state of ScreepsServer
type ScreepsServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name        string      `json:"name,omitempty"`
	Branch      string      `json:"branch,omitempty"`
	Tag         string      `json:"tag,omitempty"`
	PullRequest PullRequest `json:"pullRequest,omitempty"`
}

type Status string

const (
	StatusCreated Status = "Created"
	StatusRunning Status = "Running"
)

// ScreepsServerStatus defines the observed state of ScreepsServer
type ScreepsServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status      Status `json:"privateServerStatus,omitempty"`
	ServiceHost string `json:"serviceHost,omitempty"`
	ServicePort int32  `json:"servicePort,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ScreepsServer is the Schema for the screepsservers API
type ScreepsServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScreepsServerSpec   `json:"spec,omitempty"`
	Status ScreepsServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScreepsServerList contains a list of ScreepsServer
type ScreepsServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScreepsServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScreepsServer{}, &ScreepsServerList{})
}
