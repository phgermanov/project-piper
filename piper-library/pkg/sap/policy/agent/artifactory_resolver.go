package agent

import (
	"fmt"
	"net/http"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	artifactoryBaseURL = "https://int.repositories.cloud.sap/artifactory/cumulus-binary/cumulus-policy-cli/bin/"
)

type artifactoryResolver struct {
	baseResolver
}

func newArtifactoryResolver(downloader piperHttp.Downloader) (*artifactoryResolver, error) {
	return &artifactoryResolver{baseResolver: baseResolver{downloader: downloader}}, nil
}

func (resolver *artifactoryResolver) ResolveUrl(version string) (string, error) {
	binary := resolver.resolveBinary()
	return fmt.Sprintf("%s%s/%s", artifactoryBaseURL, version, binary), nil
}

func (resolver *artifactoryResolver) Download(version, targetFile string) error {
	url, err := resolver.ResolveUrl(version)
	log.Entry().Debugf("Download cumulus policy agent from artifactory url \"%s\"...", url)
	if err != nil {
		return err
	}
	err = resolver.downloader.DownloadFile(url, targetFile, make(http.Header), nil)
	if err != nil {
		return err
	}
	log.Entry().Debugf("Download of cumulus policy agent successful!")

	return nil
}
