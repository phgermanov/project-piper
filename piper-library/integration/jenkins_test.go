// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SAP/jenkins-library/pkg/jenkins"
)

func TestTriggerJenkinsJob(t *testing.T) {
	os.Setenv("PIPER_INTEGRATION_JENKINS_HOST", "https://gkepipervalidation.jaas-gcp.cloud.sap.corp")
	os.Setenv("PIPER_INTEGRATION_JENKINS_JOB_NAME", "Jenkins-Integration-Test")

	host := os.Getenv("PIPER_INTEGRATION_JENKINS_HOST")
	jobName := os.Getenv("PIPER_INTEGRATION_JENKINS_JOB_NAME")
	user := os.Getenv("PIPER_INTEGRATION_JENKINS_USER_NAME")
	token := os.Getenv("PIPER_INTEGRATION_JENKINS_TOKEN")
	
	require.NotEmpty(t, host, "Jenkins host url is missing")
	require.NotEmpty(t, jobName, "Jenkins job name is missing")
	require.NotEmpty(t, user, "Jenkins user name is missing")
	require.NotEmpty(t, token, "Jenkins token is missing")

	ctx := context.Background()
	jenx, err := jenkins.Instance(ctx, &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}, host, user, token)
	require.NoError(t, err)
	require.NotNil(t, jenx, "could not connect to Jenkins instance")
	// test
	job, getJobErr := jenkins.GetJob(ctx, jenx, jobName)
	build, triggerJobErr := jenkins.TriggerJob(ctx, jenx, job, nil)
	// asserts
	assert.NoError(t, getJobErr)
	assert.NoError(t, triggerJobErr)
	assert.NotNil(t, build)
	assert.True(t, build.IsRunning(ctx) || build.IsGood(ctx))
}
