package apiserver

import (
	"github.com/guodoliu/apiserver/pkg/apis/demo"
	"github.com/guodoliu/apiserver/pkg/apis/demo/install"
	"github.com/guodoliu/apiserver/pkg/registry"
	foostorage "github.com/guodoliu/apiserver/pkg/registry/demo/foo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	clientrest "k8s.io/client-go/rest"
)

var (
	Scheme = runtime.NewScheme()
	Codec  = serializer.NewCodecFactory(Scheme)
)

func init() {
	install.Install(Scheme)
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Group: "", Version: "v1"})

	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

type ExtraConfig struct {
	Rest              *clientrest.Config
	EnableEtcdStorage bool
}

type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

type DemoServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

type CompletedConfig struct {
	*completedConfig
}

func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}
	return CompletedConfig{&c}
}

func (c completedConfig) New() (*DemoServer, error) {
	genericServer, err := c.GenericConfig.New("demo-apiserver", genericapiserver.NewEmptyDelegate())
	if err != nil {
	}

	s := &DemoServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(demo.GroupName, Scheme, metav1.ParameterCodec, Codec)

	v1alpha1storage := map[string]rest.Storage{}
	v1alpha1storage["foos"] = registry.RESTInPeace(foostorage.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter))
	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}
	return s, nil
}
