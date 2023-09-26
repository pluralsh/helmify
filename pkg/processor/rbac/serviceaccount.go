package rbac

import (
	"fmt"
	"io"

	"github.com/iancoleman/strcase"
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

const SaHelperTeml = `{{/*
Create the name of the service account to use
*/}}
{{- define "%[1]s.%[2]sServiceAccountName" -}}
{{- if .Values.serviceAccounts.%[2]s.create }}
{{- default (include "cluster-api-provider-azure.fullname" .) .Values.serviceAccounts.%[2]s.name }}
{{- else }}
{{- default "default" .Values.serviceAccounts.%[2]s.name }}
{{- end }}
{{- end }}`

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

	helperTpl := fmt.Sprintf(SaHelperTeml, appMeta.ChartName(), strcase.ToLowerCamel(name))

	saVals := map[string]interface{}{
		"create":      true,
		"name":        "",
		"labels":      map[string]interface{}{},
		"annotations": map[string]interface{}{},
	}

	values := helmify.Values{}
	_ = unstructured.SetNestedField(values, saVals, "serviceAccounts", strcase.ToLowerCamel(name))
	return true, &saResult{
		name:    name,
		data:    []byte(meta),
		values:  values,
		helpers: []byte(helperTpl),
	}, nil
}

type saResult struct {
	name    string
	data    []byte
	values  helmify.Values
	helpers []byte
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

func (r *saResult) HelpersFilename() string {
	return "_" + r.name + "-helpers.tpl"
}

func (r *saResult) HelpersWrite(writer io.Writer) error {
	_, err := writer.Write(r.helpers)
	return err
}
