//go:build unit
// +build unit

package dwc

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestCLIBinaryResolver_InstallDwCCli(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name            string
		authToken       string
		prepareResolver func(self *testCase, t *testing.T) CLIBinaryResolver
		wantErr         bool
	}
	tests := []*testCase{
		{
			name:      "canonical install",
			authToken: "90234jasd",
			prepareResolver: func(self *testCase, t *testing.T) CLIBinaryResolver {
				// CLI release Resolver will be called with the supplied auth token
				releaseResolver := newMockCliReleaseResolver(t)
				releaseResolver.On("GetCLIReleaseURL", self.authToken).Return("myURL", nil)
				// the file will be downloaded next. The download location on the fs is an implementation detail. Download URL should match the resolved one.
				fileDownloader := newMockFileDownloader(t)
				fileDownloader.On("DownloadFile", "myURL", mock.Anything, mock.Anything, mock.Anything).Return(
					func(url string, filename string, header http.Header, cookies []*http.Cookie) error {
						// The binary should match the targetBinary constant as it is the command invoke base
						// we presuppose the target download path is part of the $PATH
						if !strings.HasSuffix(filename, targetBinary) {
							t.Errorf("Expected download location to end with %s but it does not: %s", targetBinary, fileDownloader)
						}
						// The http header should accept an octet-stream and Authorization should match the supplied token
						if header.Get("Authorization") != fmt.Sprintf("token %s", self.authToken) {
							t.Errorf("Authorization header for cli download was supposed to be %s but is is %s", fmt.Sprintf("token %s", self.authToken), header.Get("Authorization"))
						}
						if header.Get("Accept") != "application/octet-stream" {
							t.Errorf("Accept header for cli download was supposed to be application/octet-stream but it is %s", header.Get("Accept"))
						}
						return nil
					},
				)
				// file permissions should be set to -rwxr-xr-x after download
				permEditor := newMockFilePermissionEditor(t)
				permEditor.On("Chmod", mock.Anything, os.FileMode(0755)).Return(
					func(path string, mode os.FileMode) error {
						if !strings.HasSuffix(path, targetBinary) {
							t.Errorf("expected -rwxr-xr-x file permissions to be set on the binary %s but it was set on %s", targetBinary, path)
						}
						return nil
					},
				)
				return CLIBinaryResolver{
					FileDownloader:     fileDownloader,
					PermEditor:         permEditor,
					CLIReleaseResolver: releaseResolver,
				}
			},
			wantErr: false,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			resolver := testCase.prepareResolver(testCase, t)
			err := resolver.InstallDwCCli(testCase.authToken)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("InstallDwCCli() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}
