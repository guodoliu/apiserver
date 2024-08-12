package foo

import (
	"context"
	"fmt"
	"github.com/guodoliu/apiserver/pkg/apis/demo"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
)

func NewStrategy(typer runtime.ObjectTyper) fooStrategy {
	return fooStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	apiserver, ok := obj.(*demo.Foo)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Foo")
	}
	return apiserver.ObjectMeta.Labels, SelectableFields(apiserver), nil
}

type fooStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func SelectableFields(obj *demo.Foo) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

func MatchFoo(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func (fooStrategy) NamespaceScoped() bool                                         { return true }
func (fooStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object)      {}
func (fooStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {}

func (fooStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (fooStrategy) Canonicalize(obj runtime.Object) {}

func (fooStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (fooStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (fooStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (fooStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (fooStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
