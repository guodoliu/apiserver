/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FooSpec   `json:"spec,omitempty"`
	Status FooStatus `json:"status,omitempty"`
}

// FooList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FooList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Foo `json:"items"`
}

// FooSpec defines the desired state of Foo
type FooSpec struct {
	// Container image that the container is running to do our foo work
	Image string
	// Config is the configuration used by foo container
	Config FooConfig
}

type FooConfig struct {
	// Msg says hello world!
	Msg string
	// +optional
	Msg1 string
}

type FooStatus struct {
	Phase      FooPhase       `json:"phase,omitempty"`
	Conditions []FooCondition `json:"conditions,omitempty"`
}

type FooPhase string

type FooConditionType string

type FooCondition struct {
	Type   FooConditionType       `json:"type"`
	Status metav1.ConditionStatus `json:"status"`
}
