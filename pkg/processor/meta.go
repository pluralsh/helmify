package processor

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arttor/helmify/pkg/helmify"
	yamlformat "github.com/arttor/helmify/pkg/yaml"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
)

const metaTeml = `apiVersion: %[1]s
kind: %[2]s
metadata:
  name: %[3]s
  labels:
%[5]s
  {{- include "%[4]s.labels" . | nindent 4 }}
%[6]s`

const metaAnnTeml = `apiVersion: %[1]s
kind: %[2]s
metadata:
  name: %[3]s
  labels:
%[6]s
  {{- include "%[4]s.labels" . | nindent 4 }}
  {{- with .Values.%[5]s.labels }}
  {{- toYaml . | nindent 4 }}
  {{- end }}
  annotations:
%[7]s
  {{- with .Values.%[5]s.annotations }}
  {{- toYaml . | nindent 4 }}
  {{- end }}`

const metaAnnSaTeml = `{{- if .Values.%[4]s.create }}
apiVersion: %[1]s
kind: %[2]s
metadata:
  name: {{ include "%[3]s.serviceAccountName" . }}
  labels:
%[5]s
  {{- include "%[3]s.labels" . | nindent 4 }}
  {{- with .Values.%[4]s.labels }}
  {{ toYaml . | nindent 4 }}
  {{- end }}
  annotations:
%[6]s
  {{- with .Values.%[4]s.annotations }}
  {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}`

var serviceAccountGVC = schema.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "ServiceAccount",
}

var deploymentGVC = schema.GroupVersionKind{
	Group:   "apps",
	Version: "v1",
	Kind:    "Deployment",
}

// ProcessObjMeta - returns object apiVersion, kind and metadata as helm template.
func ProcessObjMeta(appMeta helmify.AppMetadata, obj *unstructured.Unstructured) (string, error) {
	var err error
	var labels, annotations string
	if len(obj.GetLabels()) != 0 {
		l := obj.GetLabels()
		// provided by Helm
		delete(l, "app.kubernetes.io/name")
		delete(l, "app.kubernetes.io/instance")
		delete(l, "app.kubernetes.io/version")
		delete(l, "app.kubernetes.io/managed-by")
		delete(l, "helm.sh/chart")

		// Since we delete labels above, it is possible that at this point there are no more labels.
		if len(l) > 0 {
			labels, err = yamlformat.Marshal(l, 4)
			if err != nil {
				return "", err
			}
		}
	}
	if len(obj.GetAnnotations()) != 0 {
		a := obj.GetAnnotations()
		// Since we delete labels above, it is possible that at this point there are no more labels.
		if len(a) > 0 {
			annotations, err = yamlformat.Marshal(a, 4)
			if err != nil {
				logrus.Debug("Failed to marshal annotations: ", err)
				return "", err
			}
		}
	}

	name := obj.GetName()
	templatedName := appMeta.TemplatedName(name)
	apiVersion, kind := obj.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

	var metaStr string

	if obj.GroupVersionKind() == serviceAccountGVC {
		metaStr = fmt.Sprintf(metaAnnSaTeml, apiVersion, kind, appMeta.ChartName(), strcase.ToLowerCamel(kind), labels, annotations)
	} else if obj.GroupVersionKind() == deploymentGVC {
		metaStr = fmt.Sprintf(metaAnnTeml, apiVersion, kind, templatedName, appMeta.ChartName(), strcase.ToLowerCamel(appMeta.TrimName(name)), labels, annotations)
	} else {
		metaStr = fmt.Sprintf(metaTeml, apiVersion, kind, templatedName, appMeta.ChartName(), labels, annotations)
		if len(obj.GetAnnotations()) != 0 {
			annotations, err = yamlformat.Marshal(map[string]interface{}{"annotations": obj.GetAnnotations()}, 2)
			if err != nil {
				return "", err
			}
		}
	}

	metaStr = strings.Trim(metaStr, " \n")
	metaStr = strings.ReplaceAll(metaStr, "\n\n", "\n")
	return metaStr, nil
}
