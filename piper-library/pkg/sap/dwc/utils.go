package dwc

import (
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chartutil"
)

type helmChartUtils interface {
	IsChartDir(dirName string) (bool, error)
	ReadValuesFile(filename string) (chartutil.Values, error)
}

type DefaultHelmChartUtils struct{}

func (utils *DefaultHelmChartUtils) IsChartDir(dirName string) (bool, error) {
	return chartutil.IsChartDir(dirName)
}

func (utils *DefaultHelmChartUtils) ReadValuesFile(filename string) (chartutil.Values, error) {
	return chartutil.ReadValuesFile(filename)
}

type yamlUtils interface {
	Marshal(in interface{}) (out []byte, err error)
	Unmarshal(in []byte, out interface{}) (err error)
}

type DefaultYAMLUtils struct{}

func (utils *DefaultYAMLUtils) Marshal(in interface{}) (out []byte, err error) {
	return yaml.Marshal(in)
}

func (utils *DefaultYAMLUtils) Unmarshal(in []byte, out interface{}) (err error) {
	return yaml.Unmarshal(in, out)
}

func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
