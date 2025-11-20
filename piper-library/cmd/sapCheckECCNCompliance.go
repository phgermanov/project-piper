package cmd

import (
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	eccn "github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/eccn"
)

func sapCheckECCNCompliance(config sapCheckECCNComplianceOptions, telemetryData *telemetry.CustomData) {
	telemetryData.ServerURL = config.ServerURL

	ECCNSystem := eccn.System{
		ServerURL:  config.ServerURL,
		Username:   config.Username,
		Password:   config.Password,
		HTTPClient: &piperhttp.Client{},
	}

	err := runSapCheckECCNCompliance(&config, telemetryData, &ECCNSystem, piperutils.Files{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapCheckECCNCompliance(config *sapCheckECCNComplianceOptions, telemetryData *telemetry.CustomData, eccnsystem *eccn.System, fileUtils piperutils.FileUtils) error {
	eccnData, err := eccnsystem.GetECCNData(config.PpmsID, config.EccnDetails)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve Eccn data")
	}

	telemetryData.ECCNMessageStatus = eccnData.Data.MessageStatus

	// write links to JSON for later retrieval
	links := []piperutils.Path{
		{
			Target: eccnData.Data.QuestionnaireLink,
			Name:   "ECCN Questionnaire",
		},
		{
			Target: eccnData.Data.AllQuestionnairesLink,
			Name:   "ECCN Questionnaire with comprised components",
		},
	}
	// ignore potential errors
	_ = piperutils.PersistReportsAndLinks("sapCheckECCNCompliance", ".", fileUtils, nil, links)

	return nil
}
