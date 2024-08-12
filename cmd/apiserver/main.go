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

package main

import (
	"flag"
	"fmt"
	"github.com/guodoliu/apiserver/pkg/admission/disallow"
	"github.com/guodoliu/apiserver/pkg/apis/demo"
	"github.com/guodoliu/apiserver/pkg/apiserver"
	"github.com/guodoliu/apiserver/pkg/generated/openapi"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	openapinamer "k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/cli"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/klog/v2"
	"net"
	"os"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/demo"

type Options struct {
	SecureServing *genericoptions.SecureServingOptionsWithLoopback
	KubeConfig    string
	Features      *genericoptions.FeatureOptions

	EnableEtcdStorage bool
	Etcd              *genericoptions.EtcdOptions

	EnableAuth     bool
	Authentication *genericoptions.DelegatingAuthenticationOptions
	Authorization  *genericoptions.DelegatingAuthorizationOptions

	EnableAdmission bool
	Admission       *genericoptions.AdmissionOptions
}

func (o *Options) Flags() (fs cliflag.NamedFlagSets) {
	msfs := fs.FlagSet("demo.dev-server")
	msfs.StringVar(&o.KubeConfig, "kubeconfig", o.KubeConfig, "The path to the kubeconfig used to connect to the Kubernetes API server (defaults to in-cluster config)")

	o.SecureServing.AddFlags(fs.FlagSet("apiserver secure serving"))
	o.Features.AddFlags(fs.FlagSet("features"))

	msfs.BoolVar(&o.EnableEtcdStorage, "enable-etcd-storage", false, "If true, enable etcd storage")
	o.Etcd.AddFlags(fs.FlagSet("Etcd"))

	msfs.BoolVar(&o.EnableAuth, "enable-auth", o.EnableAuth, "If true, enable authentication")
	o.Authentication.AddFlags(fs.FlagSet("apiserver authentication"))
	o.Authorization.AddFlags(fs.FlagSet("apiserver authorization"))

	msfs.BoolVar(&o.EnableAdmission, "enable-admission", o.EnableAdmission, "If true, enable admission plugins")
	return fs
}

func (o *Options) Complete() error {
	disallow.Register(o.Admission.Plugins)
	o.Admission.RecommendedPluginOrder = append(o.Admission.RecommendedPluginOrder, "DisallowFoo")
	return nil
}

func (o Options) Validate(args []string) error {
	var errs []error
	if o.EnableEtcdStorage {
		errs = o.Etcd.Validate()
	}
	if o.EnableAuth {
		errs = append(errs, o.Authentication.Validate()...)
		errs = append(errs, o.Authorization.Validate()...)
	}
	return utilerrors.NewAggregate(errs)
}

type ServerConfig struct {
	Apiserver *genericapiserver.Config
	Rest      *rest.Config
}

func (o Options) ServerConfig() (*apiserver.Config, error) {
	apiservercfg, err := o.ApiserverConfig()
	if err != nil {
		return nil, err
	}
	if o.EnableEtcdStorage {
		storageConfigCopy := o.Etcd.StorageConfig
		if storageConfigCopy.StorageObjectCountTracker == nil {
			storageConfigCopy.StorageObjectCountTracker = apiservercfg.StorageObjectCountTracker
		}
		klog.Infof("etcd cfg: %v", o.Etcd)

		if err = o.Etcd.ApplyWithStorageFactoryTo(storage.NewDefaultStorageFactory(
			o.Etcd.StorageConfig,
			o.Etcd.DefaultStorageMediaType,
			apiserver.Codec,
			storage.NewDefaultResourceEncodingConfig(apiserver.Scheme),
			apiservercfg.MergedResourceConfig,
			nil), &apiservercfg.Config); err != nil {
			return nil, err
		}
	}
	return &apiserver.Config{
		GenericConfig: apiservercfg,
		ExtraConfig: apiserver.ExtraConfig{
			EnableEtcdStorage: o.EnableEtcdStorage,
		},
	}, nil
}

func (o Options) ApiserverConfig() (*genericapiserver.RecommendedConfig, error) {
	if err := o.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificate: %v", err)
	}
	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codec)
	if err := o.SecureServing.ApplyTo(&serverConfig.SecureServing, &serverConfig.LoopbackClientConfig); err != nil {
		return nil, err
	}

	// enable OpenAPI schemas
	namer := openapinamer.NewDefinitionNamer(apiserver.Scheme)
	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(openapi.GetOpenAPIDefinitions, namer)
	serverConfig.OpenAPIConfig.Info.Title = "demo.dev-server"
	serverConfig.OpenAPIConfig.Info.Version = "0.1"

	serverConfig.OpenAPIV3Config = genericapiserver.DefaultOpenAPIV3Config(openapi.GetOpenAPIDefinitions, namer)
	serverConfig.OpenAPIV3Config.Info.Title = "demo.dev-server"
	serverConfig.OpenAPIV3Config.Info.Version = "0.1"

	if o.EnableAuth {
		if err := o.Authentication.ApplyTo(&serverConfig.Authentication, serverConfig.SecureServing, nil); err != nil {
			return nil, err
		}
		if err := o.Authorization.ApplyTo(&serverConfig.Authorization); err != nil {
			return nil, err
		}
	}

	if o.EnableAdmission {
		(&genericoptions.CoreAPIOptions{}).ApplyTo(serverConfig)

		kubeClient, err := kubernetes.NewForConfig(serverConfig.ClientConfig)
		if err != nil {
			return nil, err
		}
		dynamicClient, err := dynamic.NewForConfig(serverConfig.ClientConfig)
		if err != nil {
			return nil, err
		}
		initializers := []admission.PluginInitializer{}
		o.Admission.ApplyTo(&serverConfig.Config, serverConfig.SharedInformerFactory, kubeClient, dynamicClient, feature.DefaultFeatureGate, initializers...)
	}

	return serverConfig, nil
}

func (o Options) restConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error
	if len(o.KubeConfig) > 0 {
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: o.KubeConfig}
		loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
		config, err = loader.ClientConfig()
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client config: %v", err)
	}
	// Use protobufs for communication with apiserver
	config.ContentType = "application/vnd.kubernetes.protobuf"
	err = rest.SetKubernetesDefaults(config)
	return config, err
}

// NewDemoServerCommand provides a CLI handler for the metrics server entrypoint
func NewDemoServerCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := &Options{
		SecureServing:  genericoptions.NewSecureServingOptions().WithLoopback(),
		Etcd:           genericoptions.NewEtcdOptions(storagebackend.NewDefaultConfig(defaultEtcdPathPrefix, nil)),
		Authentication: genericoptions.NewDelegatingAuthenticationOptions(),
		Authorization:  genericoptions.NewDelegatingAuthorizationOptions(),
		Admission:      genericoptions.NewAdmissionOptions(),
	}
	opts.Etcd.StorageConfig.EncodeVersioner = runtime.NewMultiGroupVersioner(demo.SchemeGroupVersion, schema.GroupKind{Group: demo.GroupName})
	opts.Etcd.DefaultStorageMediaType = "application/json"
	opts.SecureServing.BindPort = 6443

	cmd := &cobra.Command{
		Short: "Launch a demo server",
		Long:  "Launch a demo server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := opts.Complete(); err != nil {
				return err
			}
			if err := opts.Validate(args); err != nil {
				return err
			}
			if err := runCommand(opts, stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	fs := cmd.Flags()
	nfs := opts.Flags()
	for _, f := range nfs.FlagSets {
		fs.AddFlagSet(f)
	}
	local := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(local)
	nfs.FlagSet("logging").AddGoFlagSet(local)

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), nfs, cols)
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), nfs, cols)
	})

	return cmd
}

func runCommand(options *Options, stopCh <-chan struct{}) error {
	serverCfg, err := options.ServerConfig()
	if err != nil {
		return err
	}

	server, err := serverCfg.Complete().New()
	if err != nil {
		return err
	}

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}

//func main() {
//	err := builder.APIServer.
//		// +kubebuilder:scaffold:resource-register
//		WithResource(&demov1alpha1.Foo{}).
//		Execute()
//	if err != nil {
//		klog.Fatal(err)
//	}
//}

func main() {
	stopCh := genericapiserver.SetupSignalHandler()
	cmd := NewDemoServerCommand(stopCh)
	code := cli.Run(cmd)
	os.Exit(code)
}
