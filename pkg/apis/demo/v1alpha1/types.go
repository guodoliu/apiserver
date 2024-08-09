
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
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
 	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Foo
// +k8s:openapi-gen=true
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FooSpec   `json:"spec,omitempty"`
	Status FooStatus `json:"status,omitempty"`
}

// FooList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FooList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Foo `json:"items"`
}

// FooSpec defines the desired state of Foo
type FooSpec struct {
}

var _ resource.Object = &Foo{}
var _ resourcestrategy.Validater = &Foo{}

func (in *Foo) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Foo) NamespaceScoped() bool {
	return false
}

func (in *Foo) New() runtime.Object {
	return &Foo{}
}

func (in *Foo) NewList() runtime.Object {
	return &FooList{}
}

func (in *Foo) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "demo.guodoliu.bytedance.com",
		Version:  "v1alpha1",
		Resource: "foos",
	}
}

func (in *Foo) IsStorageVersion() bool {
	return true
}

func (in *Foo) Validate(ctx context.Context) field.ErrorList {
	// TODO(user): Modify it, adding your API validation here.
	return nil
}

var _ resource.ObjectList = &FooList{}

func (in *FooList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
// FooStatus defines the observed state of Foo
type FooStatus struct {
}

func (in FooStatus) SubResourceName() string {
	return "status"
}

// Foo implements ObjectWithStatusSubResource interface.
var _ resource.ObjectWithStatusSubResource = &Foo{}

func (in *Foo) GetStatus() resource.StatusSubResource {
	return in.Status
}

// FooStatus{} implements StatusSubResource interface.
var _ resource.StatusSubResource = &FooStatus{}

func (in FooStatus) CopyTo(parent resource.ObjectWithStatusSubResource) {
	parent.(*Foo).Status = in
}
