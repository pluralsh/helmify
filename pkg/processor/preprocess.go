package processor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arttor/helmify/pkg/helmify"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"
)

// unstructuredDecoder is used to decode byte arrays into Unstructured objects.
var unstructuredDecoder = serializer.NewCodecFactory(nil).UniversalDeserializer()

// New creates processor for k8s Deployment resource.
func New() helmify.PreProcessor {
	return &preprocessor{}
}

type preprocessor struct{}

func (p preprocessor) Process(obj *unstructured.Unstructured) (*unstructured.Unstructured, helmify.Values, error) {

	yamlProcessor := yamlprocessor.NewSimpleProcessor()

	objectBytes, err := yaml.Marshal(obj)
	varMap, err := yamlProcessor.GetVariableMap(objectBytes)
	if err != nil {
		return obj, nil, err
	}

	tmpMap := make(map[string]string)
	for k, v := range varMap {

		if v != nil {
			tmpMap[k] = *v
		} else {
			tmpMap[k] = ""
		}
	}
	logrus.Debug("Env vars: ", tmpMap)

	values := helmify.Values{}
	err = createValues(values, varMap)
	if err != nil {
		logrus.Debug("Error creating values: ", err)
		return obj, nil, err
	}
	logrus.Debug("Values: ", values)

	// processed, err := yamlProcessor.Process(objectBytes, GetVarString)
	// // processed, err := yamlProcessor.Process(objectBytes, func(in string) (string, error) {
	// // 	return in, nil
	// // })
	// if err != nil {
	// 	logrus.Debug("Error preprocessing object: ", err)
	// 	return obj, nil, err
	// }

	// processedObject, err := bytesToUnstructured(processed)
	// logrus.Debug("Processed: ", string(processed))
	// if err != nil {
	// 	logrus.Debug("Error converting to unstructured: ", err)
	// 	return processedObject, nil, err
	// }

	return obj, values, nil
}

// bytesToUnstructured provides a utility method that converts a (JSON) byte array into an Unstructured object.
func bytesToUnstructured(b []byte) (*unstructured.Unstructured, error) {
	// Unmarshal the JSON.
	u := &unstructured.Unstructured{}
	if _, _, err := unstructuredDecoder.Decode(b, nil, u); err != nil {
		return u, err
	}

	return u, nil
}

func GetVarString(variable string) (string, error) {

	// tmp, _ := envsubst.Eval(variable, func(in string) string {
	// 	// v, _ := "test"
	// 	return "test"
	// })
	// logrus.Debug("Envsubst: ", tmp)

	var templateString string
	splitStr := strings.Split(variable, "_")

	for i, str := range splitStr {
		splitStr[i] = strings.ToLower(str)
	}

	if len(splitStr) >= 2 && splitStr[0] == "exp" {
		templateString = fmt.Sprintf("exprimental.%s", strcase.ToLowerCamel(strings.Join(splitStr[1:], "-")))
	} else {
		templateString = strcase.ToLowerCamel(strings.Join(splitStr, "-"))
	}

	logrus.Debug("Template string: ", templateString)

	return fmt.Sprintf(`{{ .Values.configVariables.%s }}`, templateString), nil
	// return "test", nil
}

// GetEnvVars returns a map of environment variables from a yaml artifact.
func createValues(values helmify.Values, varMap map[string]*string) error {

	// envVars := make(map[string]interface{})

	for k, v := range varMap {
		splitStr := strings.Split(k, "_")
		var value interface{}

		if v != nil {
			bool, err := strconv.ParseBool(*v)
			if err == nil {
				value = bool
			} else if intVal, err := strconv.Atoi(*v); err == nil {
				value = intVal
			} else if floatVal, err := strconv.ParseFloat(*v, 64); err == nil {
				value = floatVal
			} else {
				value = *v
			}
		} else {
			value = ""
		}

		for i, str := range splitStr {
			splitStr[i] = strings.ToLower(str)
		}
		if len(splitStr) >= 2 && splitStr[0] == "exp" {
			_ = unstructured.SetNestedField(values, value, "configVariables", "exprimental", strcase.ToLowerCamel(strings.Join(splitStr[1:], "-")))
		} else {
			_ = unstructured.SetNestedField(values, value, "configVariables", strcase.ToLowerCamel(strings.Join(splitStr, "-")))
		}
	}

	return nil
}
