package cmd

import (
	"fmt"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/daster"
)

type sapDasterExecuteScanUtils interface {
	command.ExecRunner
	piperutils.FileUtils
}

type sapDasterExecuteScanUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newSapDasterExecuteScanUtils() sapDasterExecuteScanUtils {
	utils := sapDasterExecuteScanUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func sapDasterExecuteScan(config sapDasterExecuteScanOptions, telemetryData *telemetry.CustomData) {
	utils := newSapDasterExecuteScanUtils()

	err := runSapDasterExecuteScan(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("daster execution failed")
	}
}

func initDasterInstance(config *sapDasterExecuteScanOptions) (daster.Daster, error) {
	var dasterInstance daster.Daster
	switch config.ScanType {
	case daster.FioriDASTScanType:
		dasterInstance = daster.NewFioriDASTScan(config.ServiceURL, config.Verbose, config.MaxRetries, time.Duration(config.RetryDelay)*time.Second)
	default:
		return nil, fmt.Errorf("scan type %s is currently unavailable", config.ScanType)
	}
	return dasterInstance, nil
}

func prepareSettings(config *sapDasterExecuteScanOptions, dasterInstance daster.Daster) (map[string]interface{}, error) {
	settings := config.Settings
	if settings == nil {
		settings = map[string]interface{}{}
	}

	if config.OAuthServiceURL != "" && config.ClientID != "" && config.ClientSecret != "" && settings["parameterRules"] == nil {
		token, err := dasterInstance.FetchOAuthToken(config.OAuthServiceURL, config.OAuthGrantType,
			config.OAuthSource, config.ClientID, config.ClientSecret, config.Verbose)
		if err != nil {
			log.Entry().WithError(err).Error("failed to fetch OAuth token")
			return nil, err
		}
		settings["parameterRules"] = map[string]interface{}{
			"name":     "Authorization",
			"location": "head",
			"inject":   true,
			"value":    fmt.Sprintf("Bearer %s", token),
		}
	}

	if config.TargetURL != "" && settings["targetUrl"] == nil {
		settings["targetUrl"] = config.TargetURL
	}
	if config.DasterToken != "" && settings["dasterToken"] == nil {
		settings["dasterToken"] = config.DasterToken
	}
	if config.UserCredentials != "" && settings["userCredentials"] == nil {
		settings["userCredentials"] = config.UserCredentials
	}
	return settings, nil
}

func runSapDasterExecuteScan(config *sapDasterExecuteScanOptions, telemetryData *telemetry.CustomData, utils sapDasterExecuteScanUtils) error {
	log.Entry().Debug("runSapDasterExecuteScan")
	dasterInstance, err := initDasterInstance(config)
	if err != nil {
		return err
	}
	settings, err := prepareSettings(config, dasterInstance)
	if err != nil {
		return err
	}
	scanId, err := dasterInstance.TriggerScan(settings)
	if err != nil {
		log.Entry().WithError(err).Error("failed to trigger scan")
		return err
	}
	log.Entry().Infof("Triggered scan of type %s: %s and waiting for it to complete", config.ScanType, scanId)

	if scanId == "" {
		log.Entry().Warn("Received empty scan id.")
		return nil
	}

	if !config.Synchronous {
		log.Entry().Info("Param synchronous is false, scan result will not be waited.")
		return nil
	}

	var scan *daster.Scan
	for {
		scan, err = dasterInstance.GetScan(scanId)
		if err != nil {
			log.Entry().WithError(err).Error("failed to get scan")
			return err
		}
		if scan.Terminated {
			break
		}
		time.Sleep(30 * time.Second)
	}

	if scan.ExitCode > 0 {
		return fmt.Errorf("scan failed with code %d, reason %s", scan.ExitCode, scan.Reason)
	}

	violations := checkThresholdViolations(getThresholdsConfig(config), scan.Summary)
	if violations != nil {
		return fmt.Errorf("threshold(s) %+v violated by findings %+v", *violations, *scan.Summary)
	}

	if config.DeleteScan {
		err = dasterInstance.DeleteScan(scanId)
		if err != nil {
			log.Entry().WithError(err).Warn("failed to delete scan")
		}
	}

	return nil
}

func getThresholdsConfig(config *sapDasterExecuteScanOptions) *daster.Violations {
	thresholdsConfig := &daster.Violations{}
	if config.Thresholds == nil {
		return thresholdsConfig
	}
	thresholdsConfig.High = getThresholdValue(config, "high")
	thresholdsConfig.Medium = getThresholdValue(config, "medium")
	thresholdsConfig.Low = getThresholdValue(config, "low")
	thresholdsConfig.Info = getThresholdValue(config, "info")
	thresholdsConfig.All = getThresholdValue(config, "all")
	return thresholdsConfig
}

func getThresholdValue(config *sapDasterExecuteScanOptions, level string) int {
	threshold, ok := config.Thresholds[level].(float64) // when reading int from json into interface it becomes float
	if !ok {
		log.Entry().Warnf("no %s threshold set", level)
	}
	return int(threshold)
}

func checkThresholdViolations(thresholdsConfig, scanResults *daster.Violations) *daster.Violations {
	violations := &daster.Violations{}
	if thresholdsConfig.High < scanResults.High {
		violations.High = thresholdsConfig.High
	}
	if thresholdsConfig.Medium < scanResults.Medium {
		violations.Medium = thresholdsConfig.Medium
	}
	if thresholdsConfig.Low < scanResults.Low {
		violations.Low = thresholdsConfig.Low
	}
	if thresholdsConfig.Info < scanResults.Info {
		violations.Info = thresholdsConfig.Info
	}
	if thresholdsConfig.All < (scanResults.High + scanResults.Medium + scanResults.Low + scanResults.Info) {
		violations.All = thresholdsConfig.All
	}
	return violations
}
