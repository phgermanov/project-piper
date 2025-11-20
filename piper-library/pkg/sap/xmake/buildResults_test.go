//go:build unit
// +build unit

package xmake

import (
	"context"
	"fmt"
	"testing"

	"github.com/bndr/gojenkins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
)

func TestFetchBuildResultJSON(t *testing.T) {
	ctx := context.Background()

	t.Run("artifact found", func(t *testing.T) {
		// init
		build := &mocks.Build{}
		build.On("IsRunning", ctx).Return(false)
		build.On("GetArtifacts").Return(
			[]gojenkins.Artifact{
				{FileName: mock.Anything},
				{FileName: BuildResultJSONFilename},
			},
		)
		// test
		artifact, fetchErr := FetchBuildResultJSON(ctx, build)
		// asserts
		build.AssertExpectations(t)
		assert.NoError(t, fetchErr)
		assert.Equal(t, BuildResultJSONFilename, artifact.FileName())
	})
}

func TestFetchBuildResultJSON2(t *testing.T) {
	ctx := context.Background()

	t.Run("error - failed to fetch content", func(t *testing.T) {
		// init
		artifact := &mocks.Artifact{}
		artifact.
			On("FileName").Return(BuildResultJSONFilename).
			On("GetData", ctx).Return(nil, fmt.Errorf(""))
		// test
		dataMap, fetchErr := FetchBuildResultContent(ctx, artifact)
		// asserts
		artifact.AssertExpectations(t)
		assert.Empty(t, dataMap)
		assert.EqualError(t, fetchErr, "Failed to fetch content of 'build-results.json': ")
	})
	t.Run("error - failed to unmarshal content", func(t *testing.T) {
		// init
		artifact := &mocks.Artifact{}
		artifact.
			On("FileName").Return(BuildResultJSONFilename).
			On("GetData", ctx).Return([]byte(mock.Anything), nil)
		// test
		dataMap, fetchErr := FetchBuildResultContent(ctx, artifact)
		// asserts
		artifact.AssertExpectations(t)
		assert.Empty(t, dataMap)
		assert.EqualError(t, fetchErr, "Failed to unmarshal content of 'build-results.json': invalid character 'm' looking for beginning of value")
	})
	t.Run("success - no content", func(t *testing.T) {
		// init
		artifact := &mocks.Artifact{}
		artifact.
			On("GetData", ctx).Return([]byte("{}"), nil)
		// test
		dataMap, fetchErr := FetchBuildResultContent(ctx, artifact)
		// asserts
		artifact.AssertExpectations(t)
		assert.Empty(t, dataMap)
		assert.NoError(t, fetchErr)
	})
	t.Run("success", func(t *testing.T) {
		// init
		artifact := &mocks.Artifact{}
		artifact.
			On("GetData", ctx).Return([]byte(`{"mock.Anything":"mock.Anything"}`), nil)
		// test
		dataMap, fetchErr := FetchBuildResultContent(ctx, artifact)
		// asserts
		artifact.AssertExpectations(t)
		assert.NotEmpty(t, dataMap)
		assert.Contains(t, dataMap, mock.Anything)
		assert.Equal(t, mock.Anything, dataMap[mock.Anything])
		assert.NoError(t, fetchErr)
	})
}

func TestFetchStageJSON(t *testing.T) {
	ctx := context.Background()

	t.Run("success - with filter", func(t *testing.T) {
		// init
		artifact := &mocks.Artifact{}
		artifact.
			On("GetData", ctx).Return([]byte(`{
				"TAG_EXTENSION": null,
				"TAG_PREFIX": null,
				"build_options": [
				  "--productive",
				  "-I",
				  "Common=http://nexus.wdf.sap.corp:8081/nexus/content/groups/build.milestones/",
				  "-I",
				  "DOCKER=docker.wdf.sap.corp:50001",
				  "-I",
				  "PyPi=http://nexus.wdf.sap.corp:8081/nexus/content/groups/build.milestones.pypi/simple/",
				  "-I",
				  "NPM=http://nexus.wdf.sap.corp:8081/nexus/content/groups/build.milestones.npm/",
				  "-I",
				  "GEMS=http://nexus.wdf.sap.corp:8081/nexus/content/groups/build.milestones.rubygems/",
				  "-I",
				  "HELM=https://docker.wdf.sap.corp:10443/artifactory/build.milestones.helm/",
				  "-I",
				  "NUGET=http://nexus.wdf.sap.corp:8081/nexus/content/groups/build.milestones.nuget/",
				  "-I",
				  "NUGETAPI=http://nexus.wdf.sap.corp:8081/nexus/service/local/nuget/build.milestones.nuget/",
				  "-I",
				  "APT=http://nexus.wdf.sap.corp:8081/nexus/content/groups/build.releases.apt/",
				  "-E",
				  "Common=http://nexus.wdf.sap.corp:8081/nexus/service/local/staging/profiles/should_not_be_used",
				  "-E",
				  "Docker=docker.wdf.sap.corp:51116",
				  "-E",
				  "DockerPromote=https://docker.wdf.sap.corp:10443/artifactory?sourceRepo=xmake_milestone_staging&amp;targetRepo=xmake_milestone&amp;targetNetworkLocation=docker.wdf.sap.corp:50001&amp;dmzRepo=deploy-milestones.docker.repositories.sap.ondemand.com&amp;dmzCred=deploy_milestones_docker_dmz",
				  "-E",
				  "PyPi=http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones.pypi/",
				  "-E",
				  "NPM=http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones.npm/",
				  "-E",
				  "GEMS=http://nexus.wdf.sap.corp:8081/nexus/repository/deploy.milestones.rubygems/",
				  "-E",
				  "HELM=https://docker.wdf.sap.corp:10443/artifactory/deploy.milestones.helm/",
				  "-E",
				  "NUGET=http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones.xmake/",
				  "--deploy-credentials-key",
				  "Common=deploy_milestones_xmake",
				  "--deploy-credentials-key",
				  "Docker=docker_deploy",
				  "--deploy-credentials-key",
				  "PyPi=deploy_milestones_pypi",
				  "--deploy-credentials-key",
				  "NPM=deploy_milestones_npm",
				  "--deploy-credentials-key",
				  "HELM=deploy_milestones_helm",
				  "--deploy-credentials-key",
				  "GEMS=deploy_milestones_gems",
				  "--deploy-credentials-key",
				  "NUGET=deploy_milestones_xmake"
				],
				"downstreams": {
				  "piper_validation_piper_validation_maven_SP_MS_linuxx86_64": "https://prod-build10300.wdf.sap.corp:443/job/piper-validation/job/piper-validation-maven-SP-MS-linuxx86_64/17/"
				},
				"projectArchive": "http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/Piper-Validation/Maven/f7a5d4751171941b66a1d9ce035dabfa4f1e301e/2021_07_09__04_06_31/deployPackage.tar.gz",
				"projectArchiveFiles": {
				  "piper_validation_piper_validation_maven_SP_MS_linuxx86_64": "http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/Piper-Validation/Maven/f7a5d4751171941b66a1d9ce035dabfa4f1e301e/2021_07_09__04_06_31/deployPackage.tar.gz"
				},
				"project_version": "0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d",
				"stage-bom": {
				  "dd365b588081-20210709-040623172-474": {
					"components": [
					  {
						"artifact": "piper-validation-maven",
						"assets": [
						  {
							"classifier": "releaseMetadata",
							"extension": "zip",
							"fileName": "piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d-releaseMetadata.zip",
							"relativePath": "com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d-releaseMetadata.zip",
							"url": "https://nexus.wdf.sap.corp:8443/stage/repository/dd365b588081-20210709-040623172-474/com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d-releaseMetadata.zip"
						  },
						  {
							"classifier": "releaseMetadataReference",
							"extension": "xml",
							"fileName": "piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d-releaseMetadataReference.xml",
							"relativePath": "com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d-releaseMetadataReference.xml",
							"url": "https://nexus.wdf.sap.corp:8443/stage/repository/dd365b588081-20210709-040623172-474/com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d-releaseMetadataReference.xml"
						  },
						  {
							"extension": "pom",
							"fileName": "piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.pom",
							"relativePath": "com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.pom",
							"url": "https://nexus.wdf.sap.corp:8443/stage/repository/dd365b588081-20210709-040623172-474/com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.pom"
						  },
						  {
							"extension": "war",
							"fileName": "piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.war",
							"relativePath": "com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.war",
							"url": "https://nexus.wdf.sap.corp:8443/stage/repository/dd365b588081-20210709-040623172-474/com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.war"
						  },
						  {
							"extension": "zip",
							"fileName": "piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.zip",
							"relativePath": "com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.zip",
							"url": "https://nexus.wdf.sap.corp:8443/stage/repository/dd365b588081-20210709-040623172-474/com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d/piper-validation-maven-0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d.zip"
						  }
						],
						"group": "com.sap.cc.devopscourse",
						"version": "0.0.1-20210709040309_3369dff06bd808b6803781212edc559060c3fa7d"
					  }
					],
					"credentials": {
					  "password": "4O1dQMa5x62z5cW",
					  "repository": "dd365b588081-20210709-040623172-474",
					  "repositoryURL": "https://nexus.wdf.sap.corp:8443/stage/repository/dd365b588081-20210709-040623172-474/",
					  "user": "bTwbKtM4dfuvHGP"
					},
					"format": "maven2"
				  },
				  "dd3a01b18081-20210709-040630878-903": {
					"components": [
					  {
						"artifact": "target/bom.xml",
						"assets": [
						  {
							"fileName": "bom.xml",
							"relativePath": "target/bom.xml",
							"url": "https://nexus.wdf.sap.corp:8443/stage/repository/dd3a01b18081-20210709-040630878-903/target/bom.xml"
						  }
						],
						"group": "/target",
						"version": null
					  }
					],
					"credentials": {
					  "password": "wrqa4GAevxFGAT3",
					  "repository": "dd3a01b18081-20210709-040630878-903",
					  "repositoryURL": "https://nexus.wdf.sap.corp:8443/stage/repository/dd3a01b18081-20210709-040630878-903/",
					  "user": "BjoQGtRfRlwfRQp"
					},
					"format": "raw"
				  }
				},
				"stage_gitrepo": "ssh://git@github.wdf.sap.corp/Piper-Validation/Maven.git",
				"stage_gittreeish": "f7a5d4751171941b66a1d9ce035dabfa4f1e301e",
				"stage_repourl": "https://nexus.wdf.sap.corp:8443/stage/api/repository/group-20210709-0404430970-944/",
				"staging_repo_id": "group-20210709-0404430970-944",
				"version_extension": ""
			  }`), nil)
		// test
		dataObj, fetchErr := FetchStageJSON(ctx, artifact, "*.zip")
		// asserts
		numberOfAssets := 0
		artifact.AssertExpectations(t)
		assert.NoError(t, fetchErr)
		assert.NotEmpty(t, dataObj)
		// find a better way to check this!
		for _, repository := range dataObj.StageBom {
			for key1, value1 := range repository.(map[string]interface{}) {
				if key1 == "components" {
					for _, component := range value1.([]interface{}) {
						for key2, value2 := range component.(map[string]interface{}) {
							if key2 == "assets" {
								numberOfAssets += len(value2.([]interface{}))
							}
						}
					}
				}
			}
		}
		assert.Equal(t, 2, numberOfAssets)
	})
}

func TestFilterStagesAssets(t *testing.T) {
	t.Run("", func(t *testing.T) {
		// init
		data := detailedStageJSON{
			StageBom: map[string]*stageRepository{
				mock.Anything: {
					Components: []*component{
						{
							Assets: []*asset{
								{FileName: mock.Anything + ".zip"},
								{FileName: mock.Anything + ".war"},
								{FileName: mock.Anything + ".tgz"},
								{FileName: mock.Anything + ".pom"},
							},
						},
					},
				},
			},
		}
		// test
		err := filterStagedAssets(&data, "*.zip")
		// asserts
		assert.NoError(t, err)
		require.NotEmpty(t, data.StageBom[mock.Anything].Components)
		assert.Equal(t, 1, len(data.StageBom[mock.Anything].Components[0].Assets))
		require.NotEmpty(t, data.StageBom[mock.Anything].Components[0].Assets)
		assert.Equal(t, data.StageBom[mock.Anything].Components[0].Assets[0].FileName, mock.Anything+".zip")
	})
}

func TestFilterPromotedAssets(t *testing.T) {
	t.Run("", func(t *testing.T) {
		// init
		data := PromoteJSON{
			PromoteBom: &PromoteBom{
				Repositories: []*Repository{
					{
						Success: true,
						Result: []string{
							"something",
							"anything",
						},
					},
				},
			},
		}
		// test
		err := filterPromotedAssets(&data, "something")
		// asserts
		assert.NoError(t, err)
		assert.Equal(t, 1, len(data.PromoteBom.Repositories[0].Result))
		assert.Contains(t, data.PromoteBom.Repositories[0].Result, "something")
	})

	t.Run("filtering - multi pattern", func(t *testing.T) {
		// init
		baseName := "https://common.repositories.cloud.sap/deploy.milestones/com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20220630101533_98962c9c12bd55f0510c9930edc572130e1930b4/piper-validation-maven-0.0.1-20220630101533_98962c9c12bd55f0510c9930edc572130e1930b4"
		baseName2 := "http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/com/sap/cc/devopscourse/piper-validation-maven/0.0.1-20220630101533_98962c9c12bd55f0510c9930edc572130e1930b4/piper-validation-maven-0.0.1-20220630101533_98962c9c12bd55f0510c9930edc572130e1930b4"
		data := PromoteJSON{
			PromoteBom: &PromoteBom{
				Repositories: []*Repository{
					{
						Success: true,
						Result: []string{
							baseName + "-releaseMetadata.zip",
							baseName + "-releaseMetadataReference.xml",
							baseName + ".pom",
							baseName + ".war",
							baseName + ".zip",
						},
					},
					{
						Success: true,
						Result: []string{
							baseName2 + "-releaseMetadata.zip",
							baseName2 + "-releaseMetadataReference.xml",
							baseName2 + ".pom",
							baseName2 + ".war",
							baseName2 + ".zip",
						},
					},
				},
			},
		}
		// test
		err := filterPromotedAssets(&data, "{*.war,*.zip}")
		// asserts
		assert.NoError(t, err)
		assert.Equal(t, 3, len(data.PromoteBom.Repositories[0].Result))
		assert.Equal(t, 3, len(data.PromoteBom.Repositories[1].Result))
		//TODO: could lead to flaky tests as position in the repositories list is not guaranteed
		assert.Contains(t, data.PromoteBom.Repositories[0].Result, baseName+".war")
		assert.Contains(t, data.PromoteBom.Repositories[0].Result, baseName+".zip")
		assert.Contains(t, data.PromoteBom.Repositories[1].Result, baseName2+".war")
		assert.Contains(t, data.PromoteBom.Repositories[1].Result, baseName2+".zip")
	})
}

func TestFetchPromoteJSON(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// init
		artifact := &mocks.Artifact{}
		artifact.
			On("GetData", ctx).Return([]byte(`{
				"downstreams": {
				  "anything": "nothing"
				},
				"promote-bom": {
				  "group": "group-20220503-1009090772-49",
				  "released": true,
				  "repositories": [
					{
					  "repository": "2a42cf6a8081-20220503-101101535-439",
					  "result": [
						"something",
						"anything"
					  ],
					  "success": true
					},
					{
					  "promoted": false,
					  "reason": "Target repository of same format as repository is not configured.",
					  "repository": "2d4524068081-20220503-101108886-174"
					}
				  ]
				}
			  }`), nil)
		// test
		dataObj, fetchErr := FetchPromoteJSON(ctx, artifact, "something")
		// asserts
		artifact.AssertExpectations(t)
		assert.NotEmpty(t, dataObj)
		assert.Equal(t, 1, len(dataObj.PromoteBom.Repositories[0].Result))
		assert.Contains(t, dataObj.PromoteBom.Repositories[0].Result, "something")
		assert.NoError(t, fetchErr)
	})

	t.Run("failure - invalid artifact pattern", func(t *testing.T) {
		// init
		artifact := &mocks.Artifact{}
		artifact.
			On("GetData", ctx).Return([]byte(`{
				"promote-bom": {
				  "repositories": [
					{
					  "result": [
						"anything"
					  ],
					  "success": true
					}
				  ]
				}
			  }`), nil)
		// test
		_, fetchErr := FetchPromoteJSON(ctx, artifact, "{{")
		// asserts
		artifact.AssertExpectations(t)
		assert.EqualError(t, fetchErr, "failed to match artifact: syntax error in pattern")
	})
}
