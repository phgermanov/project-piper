package ocm

import (
	"io"
	"net/http"

	jHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

// DebugContainerRegistry retrieves the catalog of a given container registry if verbose logging is enabled and a container registry URL is set.
func DebugContainerRegistry(registryUrl string, user string, pass string) {
	if !log.IsVerbose() {
		return
	}
	if registryUrl == "" {
		log.Entry().Warn("No container registry URL found, cannot get catalog")
		return
	}

	url := registryUrl + "/v2/_catalog"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Entry().Error(err)
		return
	}
	req.SetBasicAuth(user, pass)

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
	log.Entry().Infof("repositoryCatalog: %s", string(body))
}
