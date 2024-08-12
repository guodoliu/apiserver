package v1alpha1

func SetDefaults_Foo(obj *Foo) {
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	obj.Labels["demo.k8s.io/metadata.name"] = obj.Name
}
