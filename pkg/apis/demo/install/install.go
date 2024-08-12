package install

import (
	"github.com/guodoliu/apiserver/pkg/apis/demo"
	"github.com/guodoliu/apiserver/pkg/apis/demo/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func Install(scheme *runtime.Scheme) {
	utilruntime.Must(demo.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}
