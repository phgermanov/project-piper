//go:build unit
// +build unit

package fosstars

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseRatings(t *testing.T) {
	invocationClient := InvocationClient{}
	sapFosstars := Fosstars{
		RatingLabelThreshold: "BAD",
		SCMUrl:               "testSCMUrl",
		Branch:               "testBranch",
		ExcludedLibraries:    []string{"org.slf4j/slf4j-api"},
	}
	ratingsJson := `{
	"org.slf4j/slf4j-api": {
		"metadata": {
			"status": "CALCULATED"
		},
		"ratings": [
			{
				"confidence": 9.879629629629628,
				"created": "2021-02-23T05:01:30.648+00:00",
				"id": "60348c2b5234745a4cc4694c",
				"label": "BAD",
				"value": 2.881527777777777,
				"modelRatingDefinitionId": "8a10913d7723ac90017723ac9c500000"
			}
		]
	},
	"org.yaml/snakeyaml": {
		"metadata": {
			"status": "CALCULATED"
		},
		"ratings": [
			{
				"confidence": 9.75925925925926,
				"created": "2021-02-22T19:29:28.764+00:00",
				"id": "6034061a5234745a4cc44da9",
				"label": "BAD",
				"value": 3.3164957264957257,
				"modelRatingDefinitionId": "8a10913d7723ac90017723ac9c500000"
			}
		]
	}
}`
	ratingsByteArray := []byte(ratingsJson)
	_, badRatingsOccured, responsesInRequestedState, _ := invocationClient.parseRatings(&sapFosstars, ratingsByteArray)
	assert.True(t, badRatingsOccured, "expected BAD ratings")
	assert.False(t, responsesInRequestedState, "All ratings are in CALCULATED state")
	allratingsFromFosstars := invocationClient.GetAllRatingsFromFosstars()
	assert.True(t, allratingsFromFosstars != nil, "expected allratingsFromFosstars to be not nil")
	//Negative Test
	invocationClient = InvocationClient{}
	ratingsByteArray = []byte("")
	_, _, _, err := invocationClient.parseRatings(&sapFosstars, ratingsByteArray)
	assert.True(t, err != nil, "Error expected for invalid ratingsJson")
	allratingsFromFosstars = invocationClient.GetAllRatingsFromFosstars()
	assert.True(t, allratingsFromFosstars == nil, "expected allratingsFromFosstars to be nil")
}

func TestGetFosstarsClientHeader(t *testing.T) {
	assert.True(t, getFosstarsClientHeader("") == "fosstars-piper", "Expected fosstars-piper")
	assert.True(t, getFosstarsClientHeader("test") == "fosstars-piper-test", "Expected fosstars-piper-test")
}

// Mock getRatingsForArtifacts for testing
var mockGetRatingsForArtifacts func(string, []string, string, []string) ([]byte, error)

func TestGetRatings(t *testing.T) {
	tests := []struct {
		name                  string
		mockRatingsResponse   []byte
		mockRatingsError      error
		expectedErr           bool
		expectedBadRatings    bool
		expectedRatingsLength int
	}{
		{
			name:                "invalid JSON",
			mockRatingsResponse: []byte(`{invalid json}`),
			expectedErr:         true,
		},
		{
			name:             "error from getRatingsForArtifacts",
			mockRatingsError: errors.New("network error"),
			expectedErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetRatingsForArtifacts = func(url string, artifacts []string, suffix string, certs []string) ([]byte, error) {
				return tt.mockRatingsResponse, tt.mockRatingsError
			}

			client := &InvocationClient{}
			sapFosstars := &Fosstars{
				RatingLabelThreshold: "BAD",
				RequestedRetryCount:  1,
			}
			ratings, badRatings, err := client.GetRatings(sapFosstars, "url", []string{"org.example/lib"}, time.Millisecond, nil)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBadRatings, badRatings)
				assert.Len(t, ratings, tt.expectedRatingsLength)
			}
		})
	}
}
