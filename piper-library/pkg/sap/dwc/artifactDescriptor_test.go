//go:build unit
// +build unit

package dwc

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"helm.sh/helm/v3/pkg/chartutil"
	"testing"
)

func TestHelmArtifact_prepareFiles(t *testing.T) {
	tests := []struct {
		name               string
		DescriptorBase     *DescriptorBase
		ContainerImage     string
		ImageTag           string
		PatchImageSpec     bool
		ValueFiles         []string
		HelmChartDirectory string
		ConfigureUtils     func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils)
		wantErr            bool
	}{
		{
			name: "prepare files of Helm artifact without patching and multiple value files",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "",
				ResourceName:      "",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "",
			ImageTag:           "",
			PatchImageSpec:     false,
			ValueFiles:         []string{},
			HelmChartDirectory: "helm",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				return nil, nil, nil
			},
			wantErr: false,
		},
		{
			name: "prepare files of Helm artifact with patching",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     true,
			ValueFiles:         []string{},
			HelmChartDirectory: "helm",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileWrite", mock.Anything, []byte("image:\n  pullPolicy: IfNotPresent\n  repository: myImage\n  tag: latest\nreplicas: 1\n"), mock.Anything).Return(nil)
				chartUtils := newHelmChartMockUtils(t)
				chartUtils.On("IsChartDir", "helm").Return(true, nil)
				chartUtils.On("ReadValuesFile", fmt.Sprintf("helm/%s", ValuesFileName)).Return(chartutil.Values{
					"replicas": 1,
					"image": map[string]interface{}{
						"pullPolicy": "IfNotPresent",
					},
				}, nil)
				yamlUtils := newYamlMockUtils(t)
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: false,
		},
		{
			name: "prepare files of Helm artifact with multiple value files",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     false,
			ValueFiles:         []string{"helm/values-1.yaml", "helm/values-2.yaml"},
			HelmChartDirectory: "helm",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				contentFile1 := []byte("replicas: 1\n")
				contentFile2 := []byte("image:\n    pullPolicy: IfNotPresent\n")
				mergedFile := []byte("image:\n    pullPolicy: IfNotPresent\nreplicas: 1\n")
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileRead", "helm/values-1.yaml").Return(contentFile1, nil)
				fileUtils.On("FileRead", "helm/values-2.yaml").Return(contentFile2, nil)
				fileUtils.On("FileRemove", mock.Anything).Return(nil).Times(2)
				fileUtils.On("FileWrite", fmt.Sprintf("helm/%s", ValuesFileName), mergedFile, mock.Anything).Return(nil)

				chartUtils := newHelmChartMockUtils(t)
				yamlUtils := newYamlMockUtils(t)
				yamlUtils.On("Unmarshal", contentFile1, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["replicas"] = 1
				})
				yamlUtils.On("Unmarshal", contentFile2, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["image"] = map[string]interface{}{
						"pullPolicy": "IfNotPresent",
					}
				})
				yamlUtils.On("Marshal", mock.Anything).Return(mergedFile, nil)
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: false,
		},
		{
			name: "prepare files of Helm artifact with chart in custom folder",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     false,
			ValueFiles:         []string{},
			HelmChartDirectory: "chart",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				fileUtils := newDescriptorFileMockUtils(t)
				chartUtils := newHelmChartMockUtils(t)
				yamlUtils := newYamlMockUtils(t)
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: false,
		},
		{
			name: "prepare files of Helm artifact with chart in custom folder and patching",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     true,
			ValueFiles:         []string{},
			HelmChartDirectory: "chart",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileWrite", mock.Anything, []byte("image:\n  pullPolicy: IfNotPresent\n  repository: myImage\n  tag: latest\nreplicas: 1\n"), mock.Anything).Return(nil)
				chartUtils := newHelmChartMockUtils(t)
				chartUtils.On("IsChartDir", "chart").Return(true, nil)
				chartUtils.On("ReadValuesFile", fmt.Sprintf("chart/%s", ValuesFileName)).Return(chartutil.Values{
					"replicas": 1,
					"image": map[string]interface{}{
						"pullPolicy": "IfNotPresent",
					},
				}, nil)
				yamlUtils := newYamlMockUtils(t)
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: false,
		},
		{
			name: "prepare files of Helm artifact with chart in custom folder and multiple value files",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     false,
			ValueFiles:         []string{"chart/values-1.yaml", "chart/values-2.yaml"},
			HelmChartDirectory: "chart",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				contentFile1 := []byte("replicas: 1\n")
				contentFile2 := []byte("image:\n    pullPolicy: IfNotPresent\n")
				mergedFile := []byte("image:\n    pullPolicy: IfNotPresent\nreplicas: 1\n")
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileRead", "chart/values-1.yaml").Return(contentFile1, nil)
				fileUtils.On("FileRead", "chart/values-2.yaml").Return(contentFile2, nil)
				fileUtils.On("FileRemove", mock.Anything).Return(nil).Times(2)
				fileUtils.On("FileWrite", fmt.Sprintf("chart/%s", ValuesFileName), mergedFile, mock.Anything).Return(nil)

				chartUtils := newHelmChartMockUtils(t)
				yamlUtils := newYamlMockUtils(t)
				yamlUtils.On("Unmarshal", contentFile1, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["replicas"] = 1
				})
				yamlUtils.On("Unmarshal", contentFile2, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["image"] = map[string]interface{}{
						"pullPolicy": "IfNotPresent",
					}
				})
				yamlUtils.On("Marshal", mock.Anything).Return(mergedFile, nil)
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: false,
		},
		{
			name: "failed to prepare files of Helm artifact with multiple value files while reading files",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     false,
			ValueFiles:         []string{"helm/values-1.yaml", "helm/values-2.yaml"},
			HelmChartDirectory: "helm",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileRead", mock.Anything).Return(nil, errors.New("failed to read"))

				chartUtils := newHelmChartMockUtils(t)
				yamlUtils := newYamlMockUtils(t)
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: true,
		},
		{
			name: "failed to prepare files of Helm artifact with multiple value files while removing files",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     false,
			ValueFiles:         []string{"helm/values-1.yaml", "helm/values-2.yaml"},
			HelmChartDirectory: "helm",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				contentFile1 := []byte("replicas: 1\n")
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileRead", "helm/values-1.yaml").Return(contentFile1, nil)
				fileUtils.On("FileRemove", mock.Anything).Return(errors.New("failed to remove"))

				chartUtils := newHelmChartMockUtils(t)
				yamlUtils := newYamlMockUtils(t)
				yamlUtils.On("Unmarshal", contentFile1, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["replicas"] = 1
				})
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: true,
		},
		{
			name: "failed to prepare files of Helm artifact with multiple value files while writing files",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     false,
			ValueFiles:         []string{"helm/values-1.yaml", "helm/values-2.yaml"},
			HelmChartDirectory: "helm",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				contentFile1 := []byte("replicas: 1\n")
				contentFile2 := []byte("image:\n    pullPolicy: IfNotPresent\n")
				mergedFile := []byte("image:\n    pullPolicy: IfNotPresent\nreplicas: 1\n")
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileRead", "helm/values-1.yaml").Return(contentFile1, nil)
				fileUtils.On("FileRead", "helm/values-2.yaml").Return(contentFile2, nil)
				fileUtils.On("FileRemove", mock.Anything).Return(nil).Times(2)
				fileUtils.On("FileWrite", mock.Anything, mergedFile, mock.Anything).Return(errors.New("failed to write"))

				chartUtils := newHelmChartMockUtils(t)
				yamlUtils := newYamlMockUtils(t)
				yamlUtils.On("Unmarshal", contentFile1, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["replicas"] = 1
				})
				yamlUtils.On("Unmarshal", contentFile2, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["image"] = map[string]interface{}{
						"pullPolicy": "IfNotPresent",
					}
				})
				yamlUtils.On("Marshal", mock.Anything).Return(mergedFile, nil)
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: true,
		},
		{
			name: "failed to prepare files of Helm artifact with multiple value files while unmarshal YAML",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     false,
			ValueFiles:         []string{"helm/values-1.yaml", "helm/values-2.yaml"},
			HelmChartDirectory: "helm",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				contentFile1 := []byte("replicas: 1\n")
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileRead", "helm/values-1.yaml").Return(contentFile1, nil)

				chartUtils := newHelmChartMockUtils(t)
				yamlUtils := newYamlMockUtils(t)
				yamlUtils.On("Unmarshal", contentFile1, mock.Anything).Return(errors.New("failed to unmarshal"))
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: true,
		},
		{
			name: "failed to prepare files of Helm artifact with multiple value files while marshal YAML",
			DescriptorBase: &DescriptorBase{
				StageWatchPolicy:  nil,
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			ContainerImage:     "myImage",
			ImageTag:           "latest",
			PatchImageSpec:     false,
			ValueFiles:         []string{"helm/values-1.yaml", "helm/values-2.yaml"},
			HelmChartDirectory: "helm",
			ConfigureUtils: func(t *testing.T) (descriptorFileUtils, helmChartUtils, yamlUtils) {
				contentFile1 := []byte("replicas: 1\n")
				contentFile2 := []byte("image:\n    pullPolicy: IfNotPresent\n")
				fileUtils := newDescriptorFileMockUtils(t)
				fileUtils.On("FileRead", "helm/values-1.yaml").Return(contentFile1, nil)
				fileUtils.On("FileRead", "helm/values-2.yaml").Return(contentFile2, nil)
				fileUtils.On("FileRemove", mock.Anything).Return(nil).Times(2)

				chartUtils := newHelmChartMockUtils(t)
				yamlUtils := newYamlMockUtils(t)
				yamlUtils.On("Unmarshal", contentFile1, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["replicas"] = 1
				})
				yamlUtils.On("Unmarshal", contentFile2, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					targetMap := args.Get(1).(*map[string]interface{})
					(*targetMap)["image"] = map[string]interface{}{
						"pullPolicy": "IfNotPresent",
					}
				})
				yamlUtils.On("Marshal", mock.Anything).Return(nil, errors.New("failed to marshal"))
				return fileUtils, chartUtils, yamlUtils
			},
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			fileUtils, chartUtils, yamlUtils := testCase.ConfigureUtils(t)
			helm := &HelmArtifact{
				DescriptorBase:     testCase.DescriptorBase,
				ContainerImage:     testCase.ContainerImage,
				ImageTag:           testCase.ImageTag,
				PatchImageSpec:     testCase.PatchImageSpec,
				ValueFiles:         testCase.ValueFiles,
				ChartUtils:         chartUtils,
				YAMLUtils:          yamlUtils,
				HelmChartDirectory: testCase.HelmChartDirectory,
			}
			helm.DescriptorBase.FileUtils = fileUtils
			err := helm.prepareFiles()
			if (err != nil) != testCase.wantErr {
				t.Fatalf("helm.prepareFiles() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}
