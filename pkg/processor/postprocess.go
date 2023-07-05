package processor

// import (
// 	"fmt"
// 	"strings"

// 	"github.com/pluralsh/helmify/pkg/helmify"
// 	"github.com/iancoleman/strcase"
// 	"github.com/sirupsen/logrus"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"k8s.io/apimachinery/pkg/runtime/serializer"
// 	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
// 	"sigs.k8s.io/yaml"
// )

// // unstructuredDecoder is used to decode byte arrays into Unstructured objects.
// var unstructuredDecoder = serializer.NewCodecFactory(nil).UniversalDeserializer()

// // New creates processor for k8s Deployment resource.
// func New() helmify.PostProcessor {
// 	return &postprocessor{}
// }

// type postprocessor struct{}

// func PostProcess(obj *unstructured.Unstructured) (*unstructured.Unstructured, helmify.Values, error) {

// 	yamlProcessor := yamlprocessor.NewSimpleProcessor()

// 	objectBytes, err := yaml.Marshal(obj)
// 	varMap, err := yamlProcessor.GetVariableMap(objectBytes)
// 	if err != nil {
// 		return obj, nil, err
// 	}
// 	logrus.Debug("Env vars: ", varMap)

// 	values := helmify.Values{}
// 	err = createValues(values, varMap)
// 	if err != nil {
// 		logrus.Debug("Error creating values: ", err)
// 		return obj, nil, err
// 	}
// 	logrus.Debug("Values: ", values)

// 	processed, err := yamlProcessor.Process([]byte(objectBytes), getVarString)
// 	if err != nil {
// 		logrus.Debug("Error preprocessing object: ", err)
// 		return obj, nil, err
// 	}

// 	processedObject, err := bytesToUnstructured(processed)
// 	if err != nil {
// 		logrus.Debug("Error converting to unstructured: ", err)
// 		return obj, nil, nil
// 	}

// 	// logrus.Debug("Processed: ", string(processed))

// 	return processedObject, values, nil
// }

// // bytesToUnstructured provides a utility method that converts a (JSON) byte array into an Unstructured object.
// func bytesToUnstructured(b []byte) (*unstructured.Unstructured, error) {
// 	// Unmarshal the JSON.
// 	u := &unstructured.Unstructured{}
// 	if _, _, err := unstructuredDecoder.Decode(b, nil, u); err != nil {
// 		return nil, err
// 	}

// 	return u, nil
// }

// func getVarString(variable string) (string, error) {
// 	var templateString string
// 	splitStr := strings.Split(variable, "_")

// 	for i, str := range splitStr {
// 		splitStr[i] = strings.ToLower(str)
// 	}

// 	if len(splitStr) >= 2 && splitStr[0] == "exp" {
// 		templateString = fmt.Sprintf("exprimental.%s", strcase.ToLowerCamel(strings.Join(splitStr[1:], "-")))
// 	} else {
// 		templateString = strcase.ToLowerCamel(strings.Join(splitStr, "-"))
// 	}

// 	logrus.Debug("Template string: ", templateString)

// 	return fmt.Sprintf(`{{ .Values.configVariables.%s }}`, templateString), nil
// 	// return "test", nil
// }

// // GetEnvVars returns a map of environment variables from a yaml artifact.
// func createValues(values helmify.Values, varMap map[string]*string) error {

// 	// envVars := make(map[string]interface{})

// 	for k, v := range varMap {
// 		splitStr := strings.Split(k, "_")
// 		value := ""

// 		if v != nil {
// 			value = *v
// 		}

// 		for i, str := range splitStr {
// 			splitStr[i] = strings.ToLower(str)
// 		}
// 		if len(splitStr) >= 2 && splitStr[0] == "exp" {
// 			_ = unstructured.SetNestedField(values, value, "configVariables", "exprimental", strcase.ToLowerCamel(strings.Join(splitStr[1:], "-")))
// 			// templateString = fmt.Sprintf("exprimental.%s", strcase.ToLowerCamel(strings.Join(splitStr[1:], "-")))
// 		} else {
// 			_ = unstructured.SetNestedField(values, value, "configVariables", strcase.ToLowerCamel(strings.Join(splitStr[1:], "-")))
// 			// templateString = strcase.ToLowerCamel(strings.Join(splitStr, "-"))
// 		}
// 	}

// 	return nil
// }
