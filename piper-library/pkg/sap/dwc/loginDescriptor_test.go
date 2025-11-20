//go:build unit
// +build unit

package dwc

import (
	"fmt"
	piperOrchestrator "github.com/SAP/jenkins-library/pkg/orchestrator"
	"testing"
)

func TestThemistoLoginDescriptor_buildLoginCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ThemistoLoginDescriptor
		want    dwcCommand
		wantErr bool
	}{
		{
			name: "valid login command invoke",
			ThemistoLoginDescriptor: ThemistoLoginDescriptor{
				ThemistoURL:         "themisto.dwc.tools.sap",
				CertificateFilePath: "/dev/random",
			},
			want: append(loginBaseCommand,
				fmt.Sprintf(certFlag, "/dev/random"),
				fmt.Sprintf(keyFlag, "/dev/random"),
				fmt.Sprintf(authFlag, certAuthKey),
				fmt.Sprintf(apiFlag, "themisto.dwc.tools.sap"),
			),
			wantErr: false,
		},
		{
			name: "invalid login command with missing themisto url",
			ThemistoLoginDescriptor: ThemistoLoginDescriptor{ //nolint:exhaustruct,exhaustivestruct
				CertificateFilePath: "/dev/random",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid login command with missing certificate file path",
			ThemistoLoginDescriptor: ThemistoLoginDescriptor{ //nolint:exhaustruct,exhaustivestruct
				ThemistoURL: "themisto.dwc.tools.sap",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := testCase.buildLoginCommand()
			if (err != nil) != testCase.wantErr {
				t.Fatalf("buildLoginCommand() error = %v, wantErr %v", err, testCase.wantErr)
			}
			verifyDwCCommand(t, got, testCase.want, len(loginBaseCommand))
		})
	}
}

func TestGatewayLoginDescriptor_buildLoginCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		GatewayLoginDescriptor
		want    dwcCommand
		wantErr bool
	}{
		{
			name: "valid login command invoke",
			GatewayLoginDescriptor: GatewayLoginDescriptor{
				GatewayURL:          "api.dwc.tools.sap",
				CertificateFilePath: "/dev/random",
				Project:             "project",
			},
			want: append(loginBaseCommand,
				fmt.Sprintf(certFlag, "/dev/random"),
				fmt.Sprintf(keyFlag, "/dev/random"),
				fmt.Sprintf(authFlag, certAuthKey),
				fmt.Sprintf(gatewayFlag, "api.dwc.tools.sap"),
				fmt.Sprintf(projectFlag, "project"),
			),
			wantErr: false,
		},
		{
			name: "valid login command invoke using GitHub Actions OIDC authentication",
			GatewayLoginDescriptor: GatewayLoginDescriptor{
				GatewayURL:                 "api.dwc.tools.sap",
				Project:                    "project",
				ActionsIdTokenRequestToken: "some-token",
				ActionsIdTokenRequestUrl:   "https://the-issuer.url",
				Orchestrator:               piperOrchestrator.GitHubActions,
			},
			want: append(loginBaseCommand,
				fmt.Sprintf(githubActionTokenFlag, "some-token"),
				fmt.Sprintf(oidcIssuerUrlFlag, "https://the-issuer.url"),
				fmt.Sprintf(authFlag, oidcAuthKey),
				fmt.Sprintf(gatewayFlag, "api.dwc.tools.sap"),
				fmt.Sprintf(projectFlag, "project"),
			),
			wantErr: false,
		},
		{
			name: "invalid login command with missing gateway url",
			GatewayLoginDescriptor: GatewayLoginDescriptor{ //nolint:exhaustruct,exhaustivestruct
				CertificateFilePath: "/dev/random",
				Project:             "project",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid login command with missing certificate file path",
			GatewayLoginDescriptor: GatewayLoginDescriptor{ //nolint:exhaustruct,exhaustivestruct
				GatewayURL: "gateway.dwc.tools.sap",
				Project:    "project",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid login command with missing project",
			GatewayLoginDescriptor: GatewayLoginDescriptor{ //nolint:exhaustruct,exhaustivestruct
				GatewayURL:          "gateway.dwc.tools.sap",
				CertificateFilePath: "/dev/random",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid login command with missing ActionsIdTokenRequestToken",
			GatewayLoginDescriptor: GatewayLoginDescriptor{ //nolint:exhaustruct,exhaustivestruct
				GatewayURL:               "gateway.dwc.tools.sap",
				Project:                  "project",
				ActionsIdTokenRequestUrl: "https://the-issuer.url",
				Orchestrator:             piperOrchestrator.GitHubActions,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid login command with missing ActionsIdTokenRequestUrl",
			GatewayLoginDescriptor: GatewayLoginDescriptor{ //nolint:exhaustruct,exhaustivestruct
				GatewayURL:                 "gateway.dwc.tools.sap",
				Project:                    "project",
				ActionsIdTokenRequestToken: "some-token",
				Orchestrator:               piperOrchestrator.GitHubActions,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := testCase.buildLoginCommand()
			if (err != nil) != testCase.wantErr {
				t.Fatalf("buildLoginCommand() error = %v, wantErr %v", err, testCase.wantErr)
			}
			verifyDwCCommand(t, got, testCase.want, len(loginBaseCommand))
		})
	}
}
