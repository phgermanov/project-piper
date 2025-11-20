package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/daster"
)

type sapDasterExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newSapDasterExecuteScanTestsUtils() sapDasterExecuteScanMockUtils {
	utils := sapDasterExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

type DasterMock struct{}

func NewDasterMock() *DasterMock {
	return &DasterMock{}
}

func (d *DasterMock) FetchOAuthToken(oAuthServiceURL, oAuthGrantType,
	oAuthSource, clientID, clientSecret string, verbose bool) (string, error) {
	return "token", nil
}

func (d *DasterMock) TriggerScan(request map[string]interface{}) (string, error) {
	return "scanId", nil
}

func (d *DasterMock) GetScan(scanId string) (*daster.Scan, error) {
	return &daster.Scan{Terminated: true}, nil
}

func (d *DasterMock) DeleteScan(scanId string) error {
	return nil
}

type DasterMockError struct{}

func NewDasterMockError() *DasterMockError {
	return &DasterMockError{}
}

func (d *DasterMockError) FetchOAuthToken(oAuthServiceURL, oAuthGrantType,
	oAuthSource, clientID, clientSecret string, verbose bool) (string, error) {
	return "", fmt.Errorf("error")
}

func (d *DasterMockError) TriggerScan(request map[string]interface{}) (string, error) {
	return "", fmt.Errorf("error")
}

func (d *DasterMockError) GetScan(scanId string) (*daster.Scan, error) {
	return nil, fmt.Errorf("error")
}

func (d *DasterMockError) DeleteScan(scanId string) error {
	return fmt.Errorf("error")
}

func TestPrepareSettings(t *testing.T) {
	t.Parallel()

	t.Run("empty config", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{}
		dasterMock := NewDasterMock()

		settings, err := prepareSettings(config, dasterMock)
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.Empty(t, settings)
	})

	t.Run("not empty config & empty settings", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{
			TargetURL:   "https://target.url",
			DasterToken: "token",
		}
		dasterMock := NewDasterMock()

		settings, err := prepareSettings(config, dasterMock)
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.NotEmpty(t, settings)
		assert.Equal(t, settings["targetUrl"], config.TargetURL)
		assert.Equal(t, settings["dasterToken"], config.DasterToken)
		assert.Equal(t, settings["userCredentials"], nil)
	})

	t.Run("not empty settings", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{
			TargetURL:   "https://target.url",
			DasterToken: "token",
			Settings: map[string]interface{}{
				"targetUrl": "http://another.target.url",
			},
			UserCredentials: "credentials",
		}
		dasterMock := NewDasterMock()

		settings, err := prepareSettings(config, dasterMock)
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.NotEmpty(t, settings)
		assert.Equal(t, settings["targetUrl"], "http://another.target.url")
		assert.Equal(t, settings["dasterToken"], config.DasterToken)
		assert.Equal(t, settings["userCredentials"], "credentials")
	})

	t.Run("with fetching oAuth token", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{
			TargetURL:       "https://target.url",
			DasterToken:     "token",
			ClientID:        "clientId",
			OAuthServiceURL: "http://oauth.service",
			ClientSecret:    "secret",
		}
		dasterMock := NewDasterMock()

		settings, err := prepareSettings(config, dasterMock)
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.NotEmpty(t, settings)
		assert.Equal(t, settings["targetUrl"], config.TargetURL)
		assert.Equal(t, settings["dasterToken"], config.DasterToken)
		assert.Equal(t, settings["userCredentials"], nil)
		assert.NotEmpty(t, settings["parameterRules"])
		paramRules, ok := settings["parameterRules"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, paramRules["name"], "Authorization")
		assert.Equal(t, paramRules["location"], "head")
		assert.Equal(t, paramRules["inject"], true)
		assert.Equal(t, paramRules["value"], "Bearer token")
	})

	t.Run("failed to fetch oAuth token", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{
			TargetURL:   "https://target.url",
			DasterToken: "token",
			Settings: map[string]interface{}{
				"targetUrl": "http://another.target.url",
			},
			ClientID:        "clientId",
			OAuthServiceURL: "http://oauth.service",
			ClientSecret:    "secret",
		}
		dasterMock := NewDasterMockError()

		settings, err := prepareSettings(config, dasterMock)
		assert.Error(t, err)
		assert.Nil(t, settings)
	})
}

func TestCheckThresholdViolations(t *testing.T) {
	t.Parallel()
	t.Run("Not violated", func(t *testing.T) {
		scanResults := &daster.Violations{
			High:   0,
			Medium: 0,
			Low:    0,
			Info:   0,
			All:    0,
		}
		thresholds := &daster.Violations{
			High:   1,
			Medium: 1,
			Low:    1,
			Info:   1,
			All:    4,
		}
		violations := checkThresholdViolations(thresholds, scanResults)
		assert.Equal(t, violations.High, 0)
		assert.Equal(t, violations.Medium, 0)
		assert.Equal(t, violations.Low, 0)
		assert.Equal(t, violations.Info, 0)
		assert.Equal(t, violations.All, 0)
	})
	t.Run("Violated high threshold", func(t *testing.T) {
		scanResults := &daster.Violations{
			High:   10,
			Medium: 1,
			Low:    1,
			Info:   1,
		}
		thresholds := &daster.Violations{
			High:   1,
			Medium: 1,
			Low:    1,
			Info:   1,
			All:    4,
		}
		violations := checkThresholdViolations(thresholds, scanResults)
		assert.Equal(t, violations.High, 1)
		assert.Equal(t, violations.Medium, 0)
		assert.Equal(t, violations.Low, 0)
		assert.Equal(t, violations.Info, 0)
		assert.Equal(t, violations.All, 4)
	})
	t.Run("Violated medium threshold", func(t *testing.T) {
		scanResults := &daster.Violations{
			High:   1,
			Medium: 10,
			Low:    1,
			Info:   1,
		}
		thresholds := &daster.Violations{
			High:   1,
			Medium: 1,
			Low:    1,
			Info:   1,
			All:    4,
		}
		violations := checkThresholdViolations(thresholds, scanResults)
		assert.Equal(t, violations.High, 0)
		assert.Equal(t, violations.Medium, 1)
		assert.Equal(t, violations.Low, 0)
		assert.Equal(t, violations.Info, 0)
		assert.Equal(t, violations.All, 4)
	})
	t.Run("Violated low and info thresholds", func(t *testing.T) {
		scanResults := &daster.Violations{
			High:   1,
			Medium: 1,
			Low:    10,
			Info:   10,
		}
		thresholds := &daster.Violations{
			High:   1,
			Medium: 1,
			Low:    1,
			Info:   1,
			All:    4,
		}
		violations := checkThresholdViolations(thresholds, scanResults)
		assert.Equal(t, violations.High, 0)
		assert.Equal(t, violations.Medium, 0)
		assert.Equal(t, violations.Low, 1)
		assert.Equal(t, violations.Info, 1)
		assert.Equal(t, violations.All, 4)
	})
}

func TestGetThresholdValue(t *testing.T) {
	t.Parallel()
	t.Run("No thresholds is set", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{}
		threshold := getThresholdValue(config, "high")
		assert.Equal(t, 0, threshold)
	})
	t.Run("No high threshold is set", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{
			Thresholds: map[string]interface{}{
				"medium": 10,
			},
		}
		threshold := getThresholdValue(config, "high")
		assert.Equal(t, 0, threshold)
	})
	t.Run("High threshold is set", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{
			Thresholds: map[string]interface{}{
				"high": 10.0,
			},
		}
		threshold := getThresholdValue(config, "high")
		assert.Equal(t, 10, threshold)
	})
}

func TestGetThresholdsConfig(t *testing.T) {
	t.Parallel()
	t.Run("Empty config", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{}
		thresholdsConfig := getThresholdsConfig(config)
		assert.Empty(t, thresholdsConfig)
	})
	t.Run("Not empty thresholds config", func(t *testing.T) {
		config := &sapDasterExecuteScanOptions{
			Thresholds: map[string]interface{}{
				"high":   1.0,
				"medium": 1.0,
				"low":    1.0,
				"info":   1.0,
				"all":    1.0,
			},
		}
		thresholdsConfig := getThresholdsConfig(config)
		assert.NotEmpty(t, thresholdsConfig)
		assert.Equal(t, 1, thresholdsConfig.High)
		assert.Equal(t, 1, thresholdsConfig.Medium)
		assert.Equal(t, 1, thresholdsConfig.Low)
		assert.Equal(t, 1, thresholdsConfig.Info)
		assert.Equal(t, 1, thresholdsConfig.All)
	})
}

func TestInitDasterInstance(t *testing.T) {
	t.Run("Fiori DAST scan", func(t *testing.T) {
		scanType := "fioriDASTScan"
		daster, err := initDasterInstance(&sapDasterExecuteScanOptions{
			ScanType: scanType,
		})
		assert.NoError(t, err)
		assert.NotNil(t, daster)
	})
	t.Run("Unavailable scan type", func(t *testing.T) {
		scanType := "unavailable"
		daster, err := initDasterInstance(&sapDasterExecuteScanOptions{
			ScanType: scanType,
		})
		assert.Error(t, err)
		assert.Nil(t, daster)
	})
}

func TestNewSapDasterExecuteScanUtils(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		utils := newSapDasterExecuteScanUtils()
		assert.NotEmpty(t, utils)
		assert.NotEmpty(t, utils.GetStdout())
		assert.NotEmpty(t, utils.GetStderr())
	})
}

func TestRunSapExecuteScan(t *testing.T) {
	t.Run("empty config", func(t *testing.T) {
		utils := newSapDasterExecuteScanTestsUtils()
		err := runSapDasterExecuteScan(&sapDasterExecuteScanOptions{}, &telemetry.CustomData{}, utils)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "scan type  is currently unavailable")
	})

	t.Run("unavailable scan type", func(t *testing.T) {
		utils := newSapDasterExecuteScanTestsUtils()
		err := runSapDasterExecuteScan(&sapDasterExecuteScanOptions{
			ScanType: "test",
		}, &telemetry.CustomData{}, utils)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "scan type test is currently unavailable")
	})

	t.Run("failed while preparing settings", func(t *testing.T) {
		utils := newSapDasterExecuteScanTestsUtils()
		err := runSapDasterExecuteScan(&sapDasterExecuteScanOptions{
			ScanType:        "fioriDASTScan",
			OAuthServiceURL: "http://localhost/oauth",
			ClientID:        "client-id",
			ClientSecret:    "client-secret",
		}, &telemetry.CustomData{}, utils)
		assert.Error(t, err)
	})

	t.Run("failed while triggering scan", func(t *testing.T) {
		utils := newSapDasterExecuteScanTestsUtils()
		err := runSapDasterExecuteScan(&sapDasterExecuteScanOptions{
			ScanType: "fioriDASTScan",
		}, &telemetry.CustomData{}, utils)
		assert.Error(t, err)
	})
}
