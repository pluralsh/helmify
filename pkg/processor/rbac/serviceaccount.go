package rbac

import (
	"io"

	"github.com/pluralsh/helmify/pkg/helmify"
	"github.com/pluralsh/helmify/pkg/processor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var serviceAccountGVC = schema.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "ServiceAccount",
}

// ServiceAccount creates processor for k8s ServiceAccount resource.
func ServiceAccount() helmify.Processor {
	return &serviceAccount{}
}

type serviceAccount struct{}

// Process k8s ServiceAccount object into helm template. Returns false if not capable of processing given resource type.
func (sa serviceAccount) Process(appMeta helmify.AppMetadata, obj *unstructured.Unstructured) (bool, helmify.Template, error) {
	if obj.GroupVersionKind() != serviceAccountGVC {
		return false, nil, nil
	}
	meta, err := processor.ProcessObjMeta(appMeta, obj)
	if err != nil {
		return true, nil, err
	}
	name := appMeta.TrimName(obj.GetName())

	saVals := map[string]interface{}{
		"create":      true,
		"name":        "",
		"labels":      map[string]interface{}{},
		"annotations": map[string]interface{}{},
	}

	values := helmify.Values{}
	_ = unstructured.SetNestedField(values, saVals, "serviceAccount")
	return true, &saResult{
		name:   name,
		data:   []byte(meta),
		values: values,
	}, nil
}

type saResult struct {
	name   string
	data   []byte
	values helmify.Values
}

func (r *saResult) Filename() string {
	return r.name + "-sa.yaml"
}

func (r *saResult) Values() helmify.Values {
	return r.values
}

func (r *saResult) Write(writer io.Writer) error {
	_, err := writer.Write(r.data)
	return err
}
