package demo

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Foo is an example type with a spec and a status
type Foo struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   FooSpec
	Status FooStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FooList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Foo
}

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

type FooPhase string

const (
	FooPhaseProcessing FooPhase = "Processing"
	FooPhaseReady      FooPhase = "Ready"
)

type FooStatus struct {
	Phase      FooPhase
	Conditions []FooCondition
}

type FooConditionType string

const (
	FooConditionTypeWorker FooConditionType = "Worker"
	FooConditionTypeConfig FooConditionType = "Config"
)

type FooCondition struct {
	Type   FooConditionType
	Status metav1.ConditionStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Config struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec ConfigSpec
}

type ConfigSpec struct {
	Msg  string
	Msg1 string
}
