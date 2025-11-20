package ocm

import (
	"io"
	"net/http"

	jHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/staging"
)

func DebugStaging(stagingObj *staging.Staging) {
	if !log.IsVerbose() || stagingObj == nil {
		return
	}
	if stagingObj.Token == "" {
		stagingObj.LoginAndReceiveAuthToken()
	}

	url := "https://staging.repositories.cloud.sap/stage/api/repository/BOM/" + stagingObj.RepositoryId
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Entry().Error(err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+stagingObj.Token)

	client := &jHttp.Client{}
	resp, err := client.Send(req)
	if err != nil {
		log.Entry().Error(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Entry().Error(err)
		return
	}
	log.Entry().Infof("stagingRepositoryBOM: %s", string(body))
}
