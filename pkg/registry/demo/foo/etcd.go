package foo

import (
	"github.com/guodoliu/apiserver/pkg/apis/demo"
	"github.com/guodoliu/apiserver/pkg/registry"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
)

func NewREST(scheme *runtime.Scheme, opsGetter generic.RESTOptionsGetter) (*registry.REST, error) {
	strategy := NewStrategy(scheme)

	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &demo.Foo{}
		},
		NewListFunc: func() runtime.Object {
			return &demo.FooList{}
		},
		PredicateFunc:             MatchFoo,
		DefaultQualifiedResource:  demo.Resource("foos"),
		SingularQualifiedResource: demo.Resource("foo"),
		CreateStrategy:            strategy,
		UpdateStrategy:            strategy,
		DeleteStrategy:            strategy,

		TableConvertor: rest.NewDefaultTableConvertor(demo.Resource("foos")),
	}
	options := &generic.StoreOptions{RESTOptions: opsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}
	return &registry.REST{Store: store}, nil
}
