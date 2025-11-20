//go:build unit
// +build unit

package ppms

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type ppmsMockClient struct {
	username       string
	password       string
	httpMethod     string
	httpStatusCode int
	urlCalled      string
	requestBody    io.Reader
	responseBody   string
}

func (c *ppmsMockClient) SetOptions(opts piperhttp.ClientOptions) {
	c.username = opts.Username
	c.password = opts.Password
}

func (c *ppmsMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlCalled = url
	return &http.Response{StatusCode: c.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(c.responseBody)))}, nil
}

func TestGetSCV(t *testing.T) {
	myTestClient := ppmsMockClient{responseBody: `{"d":{"Id":"1","Name":"TestSCV","TechnicalName":"TechSCV","TechnicalRelease":"1","ReviewModelRiskRatings":{"__deferred":{"uri":"https://my.Uri"}},"BuildVersions":{"__deferred":{"uri":"https://my.Uri"}},"FreeOpenSourceSoftwares":{"__deferred":{"uri":"https://my.Uri"}},"Responsibles":{"__deferred":{"uri":"https://my.Uri"}}}}`}
	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
	expected := SoftwareComponentVersion{ID: "1", Name: "TestSCV", TechnicalName: "TechSCV", TechnicalRelease: "1"}
	scv, err := sys.GetSCV("1")

	assert.NoError(t, err, "Error occurred but none expected")

	assert.Equal(t, "https://my.test.server/odataint/borm/odataforosrcy/SoftwareComponentVersions('1')?$format=json", myTestClient.urlCalled, "Called url incorrect")

	assert.Equal(t, expected.ID, scv.ID, "Retrieved SCV ID incorrect")
	assert.Equal(t, expected.Name, scv.Name, "Retrieved SCV Name incorrect")
	assert.Equal(t, expected.TechnicalName, scv.TechnicalName, "Retrieved SCV  TechnicalName incorrect")
	assert.Equal(t, expected.TechnicalRelease, scv.TechnicalRelease, "Retrieved SCV TechnicalRelease incorrect")
}

func TestGetSCVFoss(t *testing.T) {
	t.Run("Return cached FOSS", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", Foss: []Foss{{ID: "1"}, {ID: "2"}}}
		result, err := scv.GetFoss(nil)

		assert.NoError(t, err, "Error received but none expected")
		assert.Equal(t, scv.Foss, result, "Foss list incorrect")
	})

	t.Run("Return fetched FOSS", func(t *testing.T) {
		myTestClient := ppmsMockClient{responseBody: `{"d":{"results":[{"Id":"1"},{"Id":"2"},{"Id":"3"}]}}`}
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		scv := SoftwareComponentVersion{ID: "1"}

		foss, err := scv.GetFoss(&sys)

		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "https://my.test.server/odataint/borm/odataforosrcy/SoftwareComponentVersions('1')/FreeOpenSourceSoftwares?$format=json", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, []Foss{{ID: "1"}, {ID: "2"}, {ID: "3"}}, foss, "Returned Foss list incorrect")

	})
}

func TestGetBVs(t *testing.T) {
	t.Run("Return cached BV", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", BuildVersions: []BuildVersion{{ID: "1"}, {ID: "2"}}}
		result, err := scv.GetBVs(nil)

		assert.NoError(t, err, "Error received but none expected")
		assert.Equal(t, scv.BuildVersions, result, "BV list incorrect")
	})

	t.Run("Return fetched BVs", func(t *testing.T) {
		myTestClient := ppmsMockClient{responseBody: `{"d":{"results":[{"Id":"1","Name":"BV1","ReviewModelRiskRatings":{"__deferred":{"uri":"https://my.Uri"}},"FreeOpenSourceSoftwares":{"__deferred":{"uri":"https://my.Uri"}}},{"Id":"2","Name":"BV2"},{"Id":"3","Name":"BV3"}]}}`}
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		scv := SoftwareComponentVersion{ID: "1"}

		bv, err := scv.GetBVs(&sys)

		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "https://my.test.server/odataint/borm/odataforosrcy/SoftwareComponentVersions('1')/BuildVersions?$format=json", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, "1", bv[0].ID, "ID of first build version incorrect")
		assert.Equal(t, "BV1", bv[0].Name, "Name of first build version incorrect")
		assert.Equal(t, BuildVersion{ID: "2", Name: "BV2"}, bv[1], "Returned Build Version list incorrect")
		assert.Equal(t, BuildVersion{ID: "3", Name: "BV3"}, bv[2], "Returned Build Version list incorrect")
	})
}

func TestGetSCVReviewModelRiskRatings(t *testing.T) {
	t.Run("Return cached ReviewModelRiskRatings", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", ReviewModelRiskRatings: []ReviewModelRiskRating{{EntityID: "1"}, {EntityID: "2"}}}
		result, err := scv.GetReviewModelRiskRatings(nil)

		assert.NoError(t, err, "Error received but none expected")
		assert.Equal(t, scv.ReviewModelRiskRatings, result, "ReviewModelRiskRatings incorrect")
	})

	t.Run("Return fetched ReviewModelRiskRatings", func(t *testing.T) {
		myTestClient := ppmsMockClient{responseBody: `{"d":{"results":[{"ReviewModelId":"A","ReviewModelName":"Model A"},{"ReviewModelId":"B","ReviewModelName":"Model B"},{"ReviewModelId":"C","ReviewModelName":"Model C"}]}}`}
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		scv := SoftwareComponentVersion{ID: "1"}

		rr, err := scv.GetReviewModelRiskRatings(&sys)

		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "https://my.test.server/odataint/borm/odataforosrcy/SoftwareComponentVersions('1')/ReviewModelRiskRatings?$format=json", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, []ReviewModelRiskRating{{ReviewModelID: "A", ReviewModelName: "Model A"}, {ReviewModelID: "B", ReviewModelName: "Model B"}, {ReviewModelID: "C", ReviewModelName: "Model C"}}, rr, "Returned ReviewModelRiskRatings incorrect")
	})
}

func TestGetResponsibles(t *testing.T) {
	t.Run("Return cached Responsibles", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", Responsibles: []Responsible{{UserID: "1"}, {UserID: "2"}}}
		result, err := scv.GetResponsibles(nil)

		assert.NoError(t, err, "Error received but none expected")
		assert.Equal(t, scv.Responsibles, result, "Responsibles list incorrect")
	})

	t.Run("Return fetched Responsibles", func(t *testing.T) {
		myTestClient := ppmsMockClient{responseBody: `{"d":{"results":[{"UserId":"1","UserName":"User1"},{"UserId":"2","UserName":"User2"},{"UserId":"3","UserName":"User3"}]}}`}
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		scv := SoftwareComponentVersion{ID: "1"}

		r, err := scv.GetResponsibles(&sys)

		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "https://my.test.server/odataint/borm/odataforosrcy/SoftwareComponentVersions('1')/Responsibles?$format=json", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, []Responsible{{UserID: "1", UserName: "User1"}, {UserID: "2", UserName: "User2"}, {UserID: "3", UserName: "User3"}}, r, "Returned Responsibles list incorrect")
	})
}

func TestGetBuildVersionByName(t *testing.T) {
	scv := SoftwareComponentVersion{ID: "1", BuildVersions: []BuildVersion{{ID: "1", Name: "BV1"}, {ID: "2", Name: "BV2"}, {ID: "3", Name: "BV3"}}}

	bv, err := scv.GetBuildVersionByName(nil, "BV2")
	assert.NoError(t, err, "Error occurred but none expected")
	assert.Equal(t, BuildVersion{ID: "2", Name: "BV2"}, bv, "got wrong build version")
}

func TestGetBuildVersionIDByName(t *testing.T) {
	scv := SoftwareComponentVersion{ID: "1", BuildVersions: []BuildVersion{{ID: "1", Name: "BV1"}, {ID: "2", Name: "BV2"}, {ID: "3", Name: "BV3"}}}

	id, err := scv.GetBuildVersionIDByName(nil, "BV2")
	assert.NoError(t, err, "Error occurred but none expected")
	assert.Equal(t, "2", id, "got wrong build version id")
}

func TestGetLatestBuildVersion(t *testing.T) {
	t.Run("No build versions", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", BuildVersions: []BuildVersion{}}

		_, err := scv.GetLatestBuildVersion(nil)
		assert.EqualError(t, err, "no build versions available")
	})

	t.Run("Success case", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", BuildVersions: []BuildVersion{{ID: "1", SortSequence: "0000000001"}, {ID: "3", SortSequence: "0000000003"}, {ID: "2", SortSequence: "0000000002"}}}

		bv, err := scv.GetLatestBuildVersion(nil)
		assert.NoError(t, err, "Error occurred but none expected")
		assert.Equal(t, "3", bv.ID, "got wrong build version")

	})
}

func TestGetBVFoss(t *testing.T) {
	t.Run("Return cached FOSS", func(t *testing.T) {
		bv := BuildVersion{ID: "1", Foss: []Foss{{ID: "1"}, {ID: "2"}}}
		result, err := bv.GetFoss(nil)

		assert.NoError(t, err, "Error received but none expected")
		assert.Equal(t, bv.Foss, result, "Foss list incorrect")
	})

	t.Run("Return fetched FOSS", func(t *testing.T) {
		myTestClient := ppmsMockClient{responseBody: `{"d":{"results":[{"Id":"1"},{"Id":"2"},{"Id":"3"}]}}`}
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		bv := BuildVersion{ID: "1"}

		foss, err := bv.GetFoss(&sys)

		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "https://my.test.server/odataint/borm/odataforosrcy/BuildVersions('1')/FreeOpenSourceSoftwares?$format=json", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, []Foss{{ID: "1"}, {ID: "2"}, {ID: "3"}}, foss, "Returned Foss list incorrect")

	})
}

func TestGetBVReviewModelRiskRatings(t *testing.T) {
	t.Run("Return cached ReviewModelRiskRatings", func(t *testing.T) {
		bv := BuildVersion{ID: "1", ReviewModelRiskRatings: []ReviewModelRiskRating{{EntityID: "1"}, {EntityID: "2"}}}
		result, err := bv.GetReviewModelRiskRatings(nil)

		assert.NoError(t, err, "Error received but none expected")
		assert.Equal(t, bv.ReviewModelRiskRatings, result, "ReviewModelRiskRatings incorrect")
	})

	t.Run("Return fetched ReviewModelRiskRatings", func(t *testing.T) {
		myTestClient := ppmsMockClient{responseBody: `{"d":{"results":[{"ReviewModelId":"A","ReviewModelName":"Model A"},{"ReviewModelId":"B","ReviewModelName":"Model B"},{"ReviewModelId":"C","ReviewModelName":"Model C"}]}}`}
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		bv := BuildVersion{ID: "1"}

		rr, err := bv.GetReviewModelRiskRatings(&sys)

		assert.NoError(t, err, "Error occurred but none expected")

		assert.Equal(t, "https://my.test.server/odataint/borm/odataforosrcy/BuildVersions('1')/ReviewModelRiskRatings?$format=json", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, []ReviewModelRiskRating{{ReviewModelID: "A", ReviewModelName: "Model A"}, {ReviewModelID: "B", ReviewModelName: "Model B"}, {ReviewModelID: "C", ReviewModelName: "Model C"}}, rr, "Returned ReviewModelRiskRatings incorrect")
	})
}

func TestGetFossIDs(t *testing.T) {
	p := []Foss{
		Foss{ID: "foss1"},
		Foss{ID: "foss2"},
		Foss{ID: "foss3"},
	}

	expected := []string{"foss1", "foss2", "foss3"}

	assert.Equal(t, expected, GetFossIDs(&p))
}

func TestSendRequest(t *testing.T) {
	myTestClient := ppmsMockClient{responseBody: "OK"}
	sys := System{ServerURL: "https://my.test.server", Username: "TestUser", Password: "TestPassword", HTTPClient: &myTestClient}

	response, err := sys.sendRequest("POST", "/test", bytes.NewReader([]byte("Test")), nil)

	assert.NoError(t, err, "Error occurred but none expected")

	assert.Equal(t, sys.Username, myTestClient.username, "Username not passed correctly")
	assert.Equal(t, sys.Password, myTestClient.password, "Password not passed correctly")
	assert.Equal(t, "POST", myTestClient.httpMethod, "Wrong HTTP method used")
	assert.Equal(t, "https://my.test.server/test", myTestClient.urlCalled, "Called url incorrect")

	assert.Equal(t, []byte("OK"), response, "Response incorrect")

}
