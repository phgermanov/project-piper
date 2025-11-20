//go:build !release
// +build !release

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fosstars"
)

// DependencyTreeGeneratorInterface is an interface to mock DependencyTreeGenerator
type DependencyTreeGeneratorInterface struct {
	mock.Mock
}

func (_d *DependencyTreeGeneratorInterface) GetDependencyTreeFiles(pomFilePath string, globalSettingsFile string, buildQuality string) ([]string, error) {
	ret := _d.Called(pomFilePath, globalSettingsFile, buildQuality)
	var r0 []string
	if rf, ok := ret.Get(0).(func(string, string, string) []string); ok {
		r0 = rf(pomFilePath, globalSettingsFile, buildQuality)
	} else {
		r0 = ret.Get(0).([]string)
	}

	return r0, nil
}

func (_d *DependencyTreeGeneratorInterface) GetDependencyTreeForNPM(artifactsRatingsMap map[string]*fosstars.Rating, rootArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool, devDependencies []string) (*fosstars.DependencyTree, error) {
	return nil, nil
}

func (_d *DependencyTreeGeneratorInterface) ParseDependencyTreeFile(depGraphJsonFilePath string, artifactsRatingsMap map[string]*fosstars.Rating, pomArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool) (*fosstars.DependencyTree, error) {
	return nil, nil
}

func (_d *DependencyTreeGeneratorInterface) AddRelativeBuidDescriptorPathToJson(ratingDetailsJson []byte, buildDescriptorRelativePath string) ([]byte, error) {
	return nil, nil
}

func (_d *DependencyTreeGeneratorInterface) GetAllPomFilesToBuildDependencyTree(checkOutPath string) ([]string, error) {
	ret := _d.Called(checkOutPath)
	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(checkOutPath)
	} else {
		r0 = ret.Get(0).([]string)
	}

	return r0, nil
}
