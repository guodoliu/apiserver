package registry

import (
	"fmt"
	genericapiserver "k8s.io/apiserver/pkg/registry/generic/registry"
)

// REST implements a RESTStorage for API services against etcd
type REST struct {
	*genericapiserver.Store
}

func RESTInPeace(storage *REST, err error) *REST {
	if err != nil {
		err = fmt.Errorf("unable to create REST storage for a resource due to: %v", err)
		panic(err)
	}
	return storage
}
