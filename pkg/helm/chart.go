package helm

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/pluralsh/helmify/pkg/cluster"
	"github.com/pluralsh/helmify/pkg/helmify"
	"github.com/pluralsh/helmify/pkg/processor"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"
)

// NewOutput creates interface to dump processed input to filesystem in Helm chart format.
func NewOutput() helmify.Output {
	return &output{}
}

type output struct{}

// Create a helm chart in the current directory:
// chartName/
//
//	├── .helmignore   	# Contains patterns to ignore when packaging Helm charts.
//	├── Chart.yaml    	# Information about your chart
//	├── values.yaml   	# The default values for your templates
//	└── templates/    	# The template files
//	    └── _helpers.tp   # Helm default template partials
//
// Overwrites existing values.yaml and templates in templates dir on every run.
func (o output) Create(chartDir, chartName string, crd bool, certManagerAsSubchart bool, preVals helmify.Values, templates []helmify.Template, filenames []string) error {
	err := initChartDir(chartDir, chartName, crd, certManagerAsSubchart)
	if err != nil {
		return err
	}
	// group templates into files
	files := map[string][]helmify.Template{}
	values := preVals
	values[cluster.DomainKey] = cluster.DefaultDomain
	for i, template := range templates {
		file := files[filenames[i]]
		file = append(file, template)
		files[filenames[i]] = file

		if template.HelpersFilename() != "" {
			helperfile := files[template.HelpersFilename()]
			helperfile = append(helperfile, template)
			files[template.HelpersFilename()] = file
		}

		err = values.Merge(template.Values())
		if err != nil {
			return err
		}
	}
	cDir := filepath.Join(chartDir, chartName)
	for filename, tpls := range files {
		err = overwriteTemplateFile(filename, cDir, crd, tpls)
		if err != nil {
			return err
		}
	}
	err = overwriteValuesFile(cDir, values, certManagerAsSubchart)
	if err != nil {
		return err
	}
	return nil
}

func overwriteTemplateFile(filename, chartDir string, crd bool, templates []helmify.Template) error {
	// pull in crd-dir setting and siphon crds into folder
	yamlProcessor := yamlprocessor.NewSimpleProcessor()
	var subdir string
	if strings.Contains(filename, "crd") && crd {
		subdir = "crds"
		// create "crds" if not exists
		if _, err := os.Stat(filepath.Join(chartDir, "crds")); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Join(chartDir, "crds"), 0750)
			if err != nil {
				return errors.Wrap(err, "unable create crds dir")
			}
		}
	} else {
		subdir = "templates"
	}
	file := filepath.Join(chartDir, subdir, filename)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "unable to open "+file)
	}
	defer f.Close()
	for i, t := range templates {
		logrus.WithField("file", file).Debug("writing a template into")

		buf := new(bytes.Buffer)
		if strings.Contains(filename, "tpl") {
			err = t.HelpersWrite(buf)
			_, err = f.Write(buf.Bytes())
		} else {
			err = t.Write(buf)

			processed, err := yamlProcessor.Process(buf.Bytes(), processor.GetVarString)
			if err != nil {
				logrus.Debug("Error post processing object: ", err)
				return err
			}

			_, err = f.Write(processed)
		}
		// err = t.Write(f)
		if err != nil {
			return errors.Wrap(err, "unable to write into "+file)
		}
		if i != len(templates)-1 {
			_, err = f.Write([]byte("\n---\n"))
			if err != nil {
				return errors.Wrap(err, "unable to write into "+file)
			}
		}
	}
	logrus.WithField("file", file).Info("overwritten")
	return nil
}

func overwriteValuesFile(chartDir string, values helmify.Values, certManagerAsSubchart bool) error {
	if certManagerAsSubchart {
		_, err := values.Add(true, "certmanager", "installCRDs")
		if err != nil {
			return errors.Wrap(err, "unable to add cert-manager.installCRDs")
		}

		_, err = values.Add(true, "certmanager", "enabled")
		if err != nil {
			return errors.Wrap(err, "unable to add cert-manager.enabled")
		}
	}
	res, err := yaml.Marshal(values)
	if err != nil {
		return errors.Wrap(err, "unable to write marshal values.yaml")
	}

	file := filepath.Join(chartDir, "values.yaml")
	err = ioutil.WriteFile(file, res, 0600)
	if err != nil {
		return errors.Wrap(err, "unable to write values.yaml")
	}
	logrus.WithField("file", file).Info("overwritten")
	return nil
}
