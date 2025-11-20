//go:build !release
// +build !release

package mocks

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fastlane"
)

// NewFastlaneMockUtilsBundle creates an instance of FastlaneMockUtilsBundle
func NewFastlaneMockUtilsBundle() fastlane.UtilsBundle {
	utils := fastlane.UtilsBundle{
		FileUtils:  &mock.FilesMock{},
		ExecRunner: &mock.ExecMockRunner{},
	}
	return utils
}
