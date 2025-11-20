//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type sapURLScanUtilsMock struct {
	*mock.FilesMock
}

func (m *sapURLScanUtilsMock) FileOpen(name string, flag int, perm os.FileMode) (uFile, error) {
	return m.OpenFile(name, flag, perm)
}

func newSapURLScanUtilsMock() *sapURLScanUtilsMock {
	return &sapURLScanUtilsMock{
		FilesMock: &mock.FilesMock{},
	}
}

func TestNewSapURLScanUtils(t *testing.T) {
	t.Parallel()
	config := sapURLScanOptions{}
	utils := newSapURLScanUtils(&config)

	assert.NotNil(t, utils.Files)
}

func TestSapUrlScan(t *testing.T) {
	t.Parallel()

	t.Run("Simple run through", func(t *testing.T) {

		dir, err := os.MkdirTemp("", "")
		if err != nil {
			t.Fatal("failed to get temporary directory")
		}
		defer os.RemoveAll(dir) // clean up

		t.Run("test non existing proxy.log", func(t *testing.T) {
			// init
			config := &sapURLScanOptions{
				ProxyLogFile: "non_existing_proxy.log",
			}
			utilsMock := newSapURLScanUtilsMock()

			// test
			err := runSapURLScan(config, utilsMock, dir)

			// assert
			assert.Errorf(t, err, "Failed to read proxy log file '%s", config.ProxyLogFile)
		})

		t.Run("test all URLs are secure", func(t *testing.T) {
			// init
			config := &sapURLScanOptions{
				ProxyLogFile:        "secure_proxy.log",
				InsecureURLPatterns: []string{".*http://.*", ".*:8080[^0-9].*", ".*:80[^0-9].*"},
				ReportLogFile:       "report_secure.log",
			}
			utilsMock := newSapURLScanUtilsMock()

			// generate input proxy log file containing only secure urls
			secure_url1 := "1574775613.951  28765 172.17.0.3 TCP_TUNNEL/200 11748 CONNECT registry.npmjs.org:443 - HIER_DIRECT/104.16.21.35 -\n"
			secure_url2 := "1576189310.408      8 172.17.0.2 TCP_MISS/200 871 GET https://smtproxy.wdf.sap.corp:1080/repo/SUSE/Products/SLE-Module-CAP-Tools/15-SP1/x86_64/product/repodata/repomd.xml.asc - HIER_DIRECT/10.17.214.205 text/plain\n"

			testdata := secure_url1 + secure_url2

			utilsMock.AddFile(config.ProxyLogFile, []byte(testdata))

			// test
			err := runSapURLScan(config, utilsMock, dir)

			// assert
			assert.NoError(t, err)

			exists, err := utilsMock.FileExists(config.ReportLogFile)
			assert.NoError(t, err)
			assert.True(t, exists)

			report, err := utilsMock.FileRead(config.ReportLogFile)
			assert.NoError(t, err)

			reportHeader := fmt.Sprintf("URLScan Report\n"+
				"Scanning file \"%s\" for insecure URLs\n"+
				"URLs matching the following patterns will be marked as insecure: %v \n\n"+
				"The following insecure URLs found:\n",
				config.ProxyLogFile,
				config.InsecureURLPatterns)

			assert.Contains(t, string(report), reportHeader)
			assert.NotContains(t, string(report), secure_url1)
			assert.NotContains(t, string(report), secure_url2)
		})

		t.Run("test insecure URLs are found", func(t *testing.T) {
			// init
			config := &sapURLScanOptions{
				ProxyLogFile:        "insecure_proxy.log",
				InsecureURLPatterns: []string{".*http://.*", ".*:8080[^0-9].*", ".*:80[^0-9].*"},
				ReportLogFile:       "report_insecure.log",
			}
			utilsMock := newSapURLScanUtilsMock()

			// generate input proxy log file
			secure_url1 := "1574775613.951  28765 172.17.0.3 TCP_TUNNEL/200 11748 CONNECT registry.npmjs.org:443 - HIER_DIRECT/104.16.21.35 - \n"
			secure_url2 := "1576189310.408      8 172.17.0.2 TCP_MISS/200 871 GET https://smtproxy.wdf.sap.corp:1080/repo/SUSE/Products/SLE-Module-CAP-Tools/15-SP1/x86_64/product/repodata/repomd.xml.asc - HIER_DIRECT/10.17.214.205 text/plain\n"

			insecure_url1 := "1574775613.951  28765 172.17.0.3 TCP_TUNNEL/200 11748 CONNECT registry.npmjs.org:80 - HIER_DIRECT/104.16.21.35 -\n"
			insecure_url2 := "1576189310.415      6 172.17.0.2 TCP_MISS/404 274 GET http://test.com:1080/repo/SUSE/Products/SLE-Module-CAP-Tools/15-SP1/x86_64/product/repodata/repomd.xml.key - HIER_DIRECT/10.17.214.205 -	\n"
			insecure_url3 := "1574775613.951  28765 172.17.0.3 TCP_TUNNEL/200 11748 CONNECT registry.npmjs.org:8080 - HIER_DIRECT/104.16.21.35 -\n"

			testdata := insecure_url1 +
				secure_url1 +
				insecure_url2 +
				secure_url2 +
				insecure_url3

			utilsMock.AddFile(config.ProxyLogFile, []byte(testdata))

			// test
			err := runSapURLScan(config, utilsMock, dir)

			// assert
			assert.Errorf(t, err, "Found %d insecure URLs in %s", 3, config.ProxyLogFile)

			exists, err := utilsMock.FileExists(config.ReportLogFile)
			assert.NoError(t, err)
			assert.True(t, exists)

			report, err := utilsMock.FileRead(config.ReportLogFile)
			assert.NoError(t, err)

			reportHeader := fmt.Sprintf("URLScan Report\n"+
				"Scanning file \"%s\" for insecure URLs\n"+
				"URLs matching the following patterns will be marked as insecure: %v \n\n"+
				"The following insecure URLs found:\n",
				config.ProxyLogFile,
				config.InsecureURLPatterns)

			assert.Contains(t, string(report), reportHeader)

			assert.Contains(t, string(report), insecure_url1)
			assert.Contains(t, string(report), insecure_url2)
			assert.Contains(t, string(report), insecure_url3)

			assert.NotContains(t, string(report), secure_url1)
			assert.NotContains(t, string(report), secure_url2)
		})
	})
}
