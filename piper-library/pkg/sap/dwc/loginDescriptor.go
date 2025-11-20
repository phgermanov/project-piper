package dwc

import (
	"encoding/base64"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	piperOrchestrator "github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/pkg/errors"
)

// Flag collection for dwc login command

var loginBaseCommand = []string{"config", "login"}

const (
	apiFlag               = "--api=%s"
	authFlag              = "--auth=%s"
	certAuthKey           = "cert"
	oidcAuthKey           = "oidc"
	certFlag              = "--cert=%s"
	gatewayFlag           = "--gateway=%s"
	keyFlag               = "--key=%s"
	projectFlag           = "--project=%s"
	githubActionTokenFlag = "--github-action-token=%s"
	oidcIssuerUrlFlag     = "--oidc-issuer-url=%s"
)

type LoginDescriptor interface {
	buildLoginCommand() (dwcCommand, error)
}

type ThemistoLoginDescriptor struct {
	ThemistoURL         string
	CertificateFilePath string
}

type GatewayLoginDescriptor struct {
	GatewayURL                 string
	CertificateFilePath        string
	Project                    string
	ActionsIdTokenRequestToken string
	ActionsIdTokenRequestUrl   string
	Orchestrator               piperOrchestrator.Orchestrator
	UseCertLogin               bool
}

func (descriptor ThemistoLoginDescriptor) buildLoginCommand() (dwcCommand, error) {
	if descriptor.ThemistoURL == "" {
		return nil, errors.New("property ThemistoUrl of ThemistoLoginDescriptor not set")
	}
	if descriptor.CertificateFilePath == "" {
		return nil, errors.New("property CertificateFilePath of ThemistoLoginDescriptor not set")
	}
	log.Entry().Debugf("certificate file path: %s", descriptor.CertificateFilePath)
	log.Entry().Debugf("certificate file path (base64): %s", base64.StdEncoding.EncodeToString([]byte(descriptor.CertificateFilePath)))

	return append(loginBaseCommand, []string{fmt.Sprintf(apiFlag, descriptor.ThemistoURL), fmt.Sprintf(authFlag, certAuthKey), fmt.Sprintf(certFlag, descriptor.CertificateFilePath), fmt.Sprintf(keyFlag, descriptor.CertificateFilePath)}...), nil
}

func NewThemistoLoginDescriptor(themistoURL, cert string) ThemistoLoginDescriptor {
	return ThemistoLoginDescriptor{ThemistoURL: themistoURL, CertificateFilePath: cert}
}

func (descriptor GatewayLoginDescriptor) buildLoginCommand() (dwcCommand, error) {
	if descriptor.GatewayURL == "" {
		return nil, errors.New("property GatewayURL of GatewayLoginDescriptor not set")
	}
	var loginMethodSpecificParts []string
	if descriptor.Orchestrator != piperOrchestrator.GitHubActions || descriptor.UseCertLogin {
		if descriptor.CertificateFilePath == "" {
			return nil, errors.New("property CertificateFilePath of GatewayLoginDescriptor not set")
		}
		loginMethodSpecificParts = []string{fmt.Sprintf(authFlag, certAuthKey), fmt.Sprintf(certFlag, descriptor.CertificateFilePath), fmt.Sprintf(keyFlag, descriptor.CertificateFilePath)}
	} else {
		if descriptor.ActionsIdTokenRequestToken == "" {
			return nil, errors.New("property ActionsIdTokenRequestToken of GatewayLoginDescriptor not set")
		}
		if descriptor.ActionsIdTokenRequestUrl == "" {
			return nil, errors.New("property ActionsIdTokenRequestUrl of GatewayLoginDescriptor not set")
		}
		loginMethodSpecificParts = []string{fmt.Sprintf(authFlag, oidcAuthKey), fmt.Sprintf(githubActionTokenFlag, descriptor.ActionsIdTokenRequestToken), fmt.Sprintf(oidcIssuerUrlFlag, descriptor.ActionsIdTokenRequestUrl)}
	}
	if descriptor.Project == "" {
		return nil, errors.New("property Project of GatewayLoginDescriptor not set")
	}
	log.Entry().Debugf("project: %s", descriptor.Project)
	log.Entry().Debugf("certificate file path: %s", descriptor.CertificateFilePath)
	log.Entry().Debugf("certificate file path (base64): %s", base64.StdEncoding.EncodeToString([]byte(descriptor.CertificateFilePath)))

	loginCommand := append(loginBaseCommand, fmt.Sprintf(gatewayFlag, descriptor.GatewayURL), fmt.Sprintf(projectFlag, descriptor.Project))
	return append(loginCommand, loginMethodSpecificParts...), nil
}

func NewGatewayLoginDescriptor(gatewayURL, cert, project, actionsIdTokenRequestToken, actionsIdTokenRequestUrl string, orchestrator piperOrchestrator.Orchestrator, useCertLogin bool) GatewayLoginDescriptor {
	return GatewayLoginDescriptor{GatewayURL: gatewayURL, CertificateFilePath: cert, Project: project, ActionsIdTokenRequestToken: actionsIdTokenRequestToken, ActionsIdTokenRequestUrl: actionsIdTokenRequestUrl, Orchestrator: orchestrator, UseCertLogin: useCertLogin}
}
