//go:build unit
// +build unit

package cmd

import (
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	eccn "github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/eccn"
)

func TestSAPCheckECCNCompliance(t *testing.T) {
	eccnsystem := eccn.System{
		ServerURL:  "https://ifp.wdf.sap.corp",
		Username:   "test",
		Password:   "test",
		HTTPClient: &piperhttp.Client{},
	}

	telemetryCustomData := telemetry.CustomData{}

	t.Run("master mode", func(t *testing.T) {
		config := sapCheckECCNComplianceOptions{
			ServerURL: "https://ifp.wdf.sap.corp",
			PpmsID:    "01200314690900001785",
			Username:  "test",
			Password:  "test",
		}
		err := runSapCheckECCNCompliance(&config, &telemetryCustomData, &eccnsystem, &mock.FilesMock{})
		// wrapped reason, actual is 401 unauthorized
		assert.Error(t, err, "failed to retrieve Eccn data")
	})

}
