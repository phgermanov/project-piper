//go:build unit
// +build unit

package dwc

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestNewUIUploadCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		descriptorBase DescriptorBase
		uploadBasePath string
		want           dwcCommand
	}{
		{
			name: "valid command build",
			descriptorBase: DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			uploadBasePath: "",
			want: append(uploadBaseCommand,
				uiUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with multiple app names",
			descriptorBase: DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "",
				Apps:              []App{{Name: "myApp1"}, {Name: "myApp2"}},
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			},
			uploadBasePath: "myApp",
			want: append(uploadBaseCommand,
				uiUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp1"),
				fmt.Sprintf(appNameFlag, "myApp2"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			artifact, err := NewUIArtifact(testCase.descriptorBase, testCase.uploadBasePath)
			assert.Equal(t, nil, err)
			got, err := newUIUploadCommand(artifact)
			assert.Equal(t, nil, err)
			verifyDwCCommand(t, got, testCase.want, len(uploadBaseCommand)+1)
		})
	}
}

func TestNewJavaUploadCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		artifact *JavaArtifact
		want     dwcCommand
	}{
		{
			name: "valid command build",
			artifact: NewJavaArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", nil),
			want: append(uploadBaseCommand,
				javaUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(jarURLFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with multiple app names",
			artifact: NewJavaArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "",
				Apps:              []App{{Name: "myApp1"}, {Name: "myApp2"}},
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", nil),
			want: append(uploadBaseCommand,
				javaUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp1"),
				fmt.Sprintf(appNameFlag, "myApp2"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(jarURLFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with additional download URLs",
			artifact: NewJavaArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", map[string]string{"cn20": "https://common.repositories.sapcloud.cn/very-good-srv.jar"}),
			want: append(uploadBaseCommand,
				javaUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(jarURLFlag, "myURL"),
				fmt.Sprintf(additionalDownloadUrlsFlag, "cn20=https://common.repositories.sapcloud.cn/very-good-srv.jar"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := newJavaUploadCommand(testCase.artifact)
			assert.Equal(t, nil, err)
			verifyDwCCommand(t, got, testCase.want, len(uploadBaseCommand)+1)
		})
	}
}

func TestNewMTAUploadCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		artifact *MTAArtifact
		want     dwcCommand
	}{
		{
			name: "valid command build with file patterns",
			artifact: NewMTAArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      []string{"pattern/*", "another/one/*"},
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", nil),
			want: append(uploadBaseCommand,
				mtaUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(mtaURLFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build without file patterns",
			artifact: NewMTAArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", nil),
			want: append(uploadBaseCommand,
				mtaUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(mtaURLFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with multiple app names",
			artifact: NewMTAArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "",
				Apps:              []App{{Name: "myApp1"}, {Name: "myApp2"}},
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", nil),
			want: append(uploadBaseCommand,
				mtaUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp1"),
				fmt.Sprintf(appNameFlag, "myApp2"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(mtaURLFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with additional download URLs",
			artifact: NewMTAArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", map[string]string{"cn20": "https://common.repositories.sapcloud.cn/very-good-srv.mtar"}),
			want: append(uploadBaseCommand,
				mtaUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(mtaURLFlag, "myURL"),
				fmt.Sprintf(additionalDownloadUrlsFlag, "cn20=https://common.repositories.sapcloud.cn/very-good-srv.mtar"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := newMTAUploadCommand(testCase.artifact)
			assert.Equal(t, nil, err)
			verifyDwCCommand(t, got, testCase.want, len(uploadBaseCommand)+1)
		})
	}
}

func TestNewDockerUploadCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		artifact *DockerArtifact
		want     dwcCommand
	}{
		{
			name: "valid command build",
			artifact: NewDockerArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", nil),
			want: append(uploadBaseCommand,
				dockerUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(dockerImageFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with multiple app names",
			artifact: NewDockerArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "",
				Apps:              []App{{Name: "myApp1"}, {Name: "myApp2"}},
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", nil),
			want: append(uploadBaseCommand,
				dockerUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp1"),
				fmt.Sprintf(appNameFlag, "myApp2"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(dockerImageFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with additional download URLs",
			artifact: NewDockerArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", map[string]string{"cn20": "https://common.repositories.sapcloud.cn/very-good-docker-image:latest", "ap10": "https://common.repositories.sapcloud.ap/another-very-good-docker-image:latest"}),
			want: append(uploadBaseCommand,
				dockerUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(dockerImageFlag, "myURL"),
				fmt.Sprintf(additionalDownloadUrlsFlag, "cn20=https://common.repositories.sapcloud.cn/very-good-docker-image:latest"),
				fmt.Sprintf(additionalDownloadUrlsFlag, "ap10=https://common.repositories.sapcloud.ap/another-very-good-docker-image:latest"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := newDockerUploadCommand(testCase.artifact)
			assert.Equal(t, nil, err)
			verifyDwCCommand(t, got, testCase.want, len(uploadBaseCommand)+1)
		})
	}
}

func TestNewOrbitUploadCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		artifact *OrbitArtifact
		want     dwcCommand
	}{
		{
			name: "valid command build",
			artifact: NewOrbitArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", "helm", nil),
			want: append(uploadBaseCommand,
				orbitUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(containerImageFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build using apps property",
			artifact: NewOrbitArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "",
				Apps:              []App{{Name: "myApp"}},
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", "helm", nil),
			want: append(uploadBaseCommand,
				orbitUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(containerImageFlag, "myURL"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with additional download URLs",
			artifact: NewOrbitArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, "myURL", "helm", map[string]string{"cn20": "https://common.repositories.sapcloud.cn/very-good-docker-image:latest"}),
			want: append(uploadBaseCommand,
				orbitUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(containerImageFlag, "myURL"),
				fmt.Sprintf(additionalDownloadUrlsFlag, "cn20=https://common.repositories.sapcloud.cn/very-good-docker-image:latest"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := newOrbitUploadCommand(testCase.artifact)
			assert.Equal(t, nil, err)
			verifyDwCCommand(t, got, testCase.want, len(uploadBaseCommand)+1)
		})
	}
}

func TestNewHelmUploadCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		artifact *HelmArtifact
		want     dwcCommand
	}{
		{
			name: "valid command build",
			artifact: NewHelmArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "myApp",
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, true, []string{}, "myURL", "myTag", "helm"),
			want: append(uploadBaseCommand,
				helmUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
		{
			name: "valid command build with multiple app names",
			artifact: NewHelmArtifact(DescriptorBase{
				StageWatchPolicy:  AtLeastOneSuccessfulDeploymentPolicy(),
				FileUtils:         nil,
				MetadataCollector: nil,
				AppName:           "",
				Apps:              []App{{Name: "myApp1"}, {Name: "myApp2"}},
				ResourceName:      "myResource",
				FilePatterns:      nil,
				StagesToWatch:     nil,
				WatchROI:          false,
				buildMetadata:     "{}",
			}, true, []string{}, "myURL", "myTag", "helm"),
			want: append(uploadBaseCommand,
				helmUploadSubcommand,
				waitForPromotionFlag,
				fmt.Sprintf(metadataFlag, buildMetadataKey, "{}"),
				fmt.Sprintf(appNameFlag, "myApp1"),
				fmt.Sprintf(appNameFlag, "myApp2"),
				fmt.Sprintf(filesFlag, "upload.zip"),
				fmt.Sprintf(resourceFlag, "myResource"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := newHelmUploadCommand(testCase.artifact)
			assert.Equal(t, nil, err)
			verifyDwCCommand(t, got, testCase.want, len(uploadBaseCommand)+1)
		})
	}
}

func TestAppendUploadMetadata(t *testing.T) {
	t.Parallel()
	type args struct {
		args           dwcCommand
		uploadMetadata map[string]string
		apps           []App
	}
	type testCase struct {
		name string
		args args
		want dwcCommand
	}
	tests := []testCase{
		{
			name: "multiple metadata provided",
			args: args{
				args:           dwcCommand{"dwc"},
				uploadMetadata: map[string]string{"a": "1", "b": "2"},
				apps:           nil,
			},
			want: dwcCommand{"dwc", "--metadata=a=1", "--metadata=b=2"},
		},
		{
			name: "no metadata",
			args: args{
				args:           dwcCommand{"dwc"},
				uploadMetadata: map[string]string{},
				apps:           nil,
			},
			want: dwcCommand{"dwc"},
		},
		{
			name: "no upload metadata but disabled Euporie task collection",
			args: args{
				args:           dwcCommand{"dwc"},
				uploadMetadata: map[string]string{},
				apps:           []App{{Name: "myApp", NoEuporieTaskCollection: true}},
			},
			want: dwcCommand{"dwc", "--metadata=noEuporieTaskCollection=[\"myApp\"]"},
		},
		{
			name: "no upload metadata but disabled route assignment",
			args: args{
				args:           dwcCommand{"dwc"},
				uploadMetadata: map[string]string{},
				apps:           []App{{Name: "myApp", NoRouteAssignment: true}},
			},
			want: dwcCommand{"dwc", "--metadata=noRouteAssignment=[\"myApp\"]"},
		},
		{
			name: "no upload metadata but enabled static routes",
			args: args{
				args:           dwcCommand{"dwc"},
				uploadMetadata: map[string]string{},
				apps:           []App{{Name: "myApp", AllowStaticRoutes: true}},
			},
			want: dwcCommand{"dwc", "--metadata=allowStaticRoutes=[\"myApp\"]"},
		},
		{
			name: "upload metadata provided as well as two apps that disable Euporie task collection and route assignment",
			args: args{
				args:           dwcCommand{"dwc"},
				uploadMetadata: map[string]string{"a": "1"},
				apps:           []App{{Name: "myApp1", NoEuporieTaskCollection: true, NoRouteAssignment: true}, {Name: "myApp2", NoEuporieTaskCollection: true, NoRouteAssignment: true}},
			},
			want: dwcCommand{"dwc", "--metadata=a=1", "--metadata=noEuporieTaskCollection=[\"myApp1\",\"myApp2\"]", "--metadata=noRouteAssignment=[\"myApp1\",\"myApp2\"]"},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := appendUploadMetadata(test.args.args, test.args.uploadMetadata, test.args.apps)

			sort.Strings(test.want)
			sort.Strings(got) // because uploadMetadata is a map and order is unpredictable.

			assert.Equal(t, nil, err)
			assert.Equalf(t, test.want, got, "appendUploadMetadata(%v, %v, %v)", test.args.args, test.args.uploadMetadata, test.args.apps)
		})
	}
}

func TestAppendAppNameArgs(t *testing.T) {
	t.Parallel()
	type args struct {
		args           dwcCommand
		descriptorBase *DescriptorBase
	}
	type testCase struct {
		name string
		args args
		want dwcCommand
	}
	tests := []testCase{
		{
			name: "single app name via appName property",
			args: args{
				args:           dwcCommand{"dwc"},
				descriptorBase: &DescriptorBase{AppName: "myApp"},
			},
			want: dwcCommand{"dwc", "--appname=myApp"},
		},
		{
			name: "single app name via apps property",
			args: args{
				args:           dwcCommand{"dwc"},
				descriptorBase: &DescriptorBase{Apps: []App{{Name: "myApp"}}},
			},
			want: dwcCommand{"dwc", "--appname=myApp"},
		},
		{
			name: "multiple app names",
			args: args{
				args:           dwcCommand{"dwc"},
				descriptorBase: &DescriptorBase{Apps: []App{{Name: "myApp1"}, {Name: "myApp2"}}},
			},
			want: dwcCommand{"dwc", "--appname=myApp1", "--appname=myApp2"},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, test.want, appendAppNameArgs(test.args.args, test.args.descriptorBase), "appendAppNameArgs(%v, %v)", test.args.args, test.args.descriptorBase)
		})
	}
}
