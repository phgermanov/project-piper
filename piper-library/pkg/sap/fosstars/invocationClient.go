package fosstars

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/mitchellh/mapstructure"
)

const customTLSCertsObtainedMessage = "The customTLSCerts obtained are %v"

// InvocationClientInterface is an interface to mock InvocationClient
type InvocationClientInterface interface {
	GetRatings(sapFosstars *Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration, customTLSCerts []string) (map[string]*Rating, bool, error)
	GetAllRatingsFromFosstars() map[string]interface{}
	GetModelRatingDefinitionDetails(fosstarsModelRatingsDefinitionsURL string, fosstarsClientSuffix string, customTLSCerts []string) (map[string]string, error)
}

// Struct to store the Rating
type Rating struct {
	Id                      string
	RatingDefinitionId      string
	ModelRatingDefinitionId string
	Created                 string
	Value                   float64
	Confidence              float32
	Label                   string
	NameSpace               string
	Name                    string
	RepositoryType          string
	CoordinateValue         string
}

// Struct to store the ModelRatingDefinition
type ModelRatingDefinition struct {
	Uuid                 string
	Name                 string
	Details              string
	ModelFunctionVersion string
}

type InvocationClient struct {
	allRatingsFromFosstars      map[string]interface{}
	modelRatingsDefinitionIdMap map[string]struct{}
}

func (i *InvocationClient) GetRatings(sapFosstars *Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration, customTLSCerts []string) (map[string]*Rating, bool, error) {
	// get the ratings for the artifacts
	log.Entry().Infof(customTLSCertsObtainedMessage, customTLSCerts)
	var artifactsRatingsMap map[string]*Rating
	var badRatingsOccured bool
	var responsesInRequestedState bool
	numberOfRetries := 1
	for {
		ratings, err := getRatingsForArtifacts(fosstarsQueryServiceURL, artifacts, sapFosstars.FosstarsClientSuffix, customTLSCerts)
		if err != nil {
			return nil, false, err
		}
		//parse the ratings
		artifactsRatingsMap, badRatingsOccured, responsesInRequestedState, err = i.parseRatings(sapFosstars, ratings)
		if err != nil {
			return nil, badRatingsOccured, err
		}
		if !responsesInRequestedState || numberOfRetries <= sapFosstars.RequestedRetryCount {
			log.Entry().Infof("All the ratingResponses are not in REQUESTED status or the numberOfRetries have exceeded the limit.")
			break
		}
		time.Sleep(30 * duration)
		numberOfRetries = numberOfRetries + 1
	}

	return artifactsRatingsMap, badRatingsOccured, nil
}

func (i *InvocationClient) GetModelRatingDefinitionDetails(fosstarsModelRatingsDefinitionsURL string, fosstarsClientSuffix string, customTLSCerts []string) (map[string]string, error) {
	modelratingsDefinitionDetails := make(map[string]string)
	for key, _ := range i.modelRatingsDefinitionIdMap {
		var modelRatingDefinition ModelRatingDefinition
		client := piperhttp.Client{}
		log.Entry().Infof(customTLSCertsObtainedMessage, customTLSCerts)
		client.SetOptions(piperhttp.ClientOptions{TransportTimeout: 300 * time.Second, MaxRetries: 10, TrustedCerts: customTLSCerts})
		header := make(map[string][]string)
		header["Accept"] = []string{"application/json"}
		header["X-Fosstars-Client"] = []string{getFosstarsClientHeader(fosstarsClientSuffix)}
		response, err := client.SendRequest("GET", fmt.Sprintf(fosstarsModelRatingsDefinitionsURL, key), nil, header, nil)
		if err != nil {
			log.Entry().Debugf("The HTTP request failed with error %s\n", err)
			return nil, err
		}

		defer response.Body.Close()
		data, _ := io.ReadAll(response.Body)
		err = json.Unmarshal(data, &modelRatingDefinition)
		if err != nil {
			println(err)
			return nil, err
		}
		modelratingsDefinitionDetails[key] = modelRatingDefinition.Details
	}
	log.Entry().Infof("fosstarsClientSuffix: %v ", getFosstarsClientHeader(fosstarsClientSuffix))
	return modelratingsDefinitionDetails, nil
}

func getFosstarsClientHeader(fosstarsClientSuffix string) string {
	if fosstarsClientSuffix != "" {
		return "fosstars-piper-" + fosstarsClientSuffix
	}
	return "fosstars-piper"
}

func (i *InvocationClient) GetAllRatingsFromFosstars() map[string]interface{} {
	return i.allRatingsFromFosstars
}

func getRatingsForArtifacts(fosstarsQueryServiceURL string, artifacts []string, fosstarsClientSuffix string, customTLSCerts []string) ([]byte, error) {
	jsonValue, _ := json.Marshal(artifacts)
	client := piperhttp.Client{}
	log.Entry().Infof(customTLSCertsObtainedMessage, customTLSCerts)
	client.SetOptions(piperhttp.ClientOptions{TransportTimeout: 600 * time.Second, TrustedCerts: customTLSCerts})
	header := make(map[string][]string)
	header["Content-Type"] = []string{"application/json"}
	header["X-Fosstars-Client"] = []string{getFosstarsClientHeader(fosstarsClientSuffix)}

	response, err := client.SendRequest("POST", fosstarsQueryServiceURL, bytes.NewReader(jsonValue), header, nil)
	if err != nil {
		log.Entry().Debugf("The HTTP request failed with error %s\n", err)
		return nil, err
	}

	defer response.Body.Close()
	data, _ := io.ReadAll(response.Body)
	return data, nil
}

func createUnknownRating() *Rating {
	return &Rating{Id: "", Label: "UNKNOWN", Created: ""}
}

func processRatingResponse(i *InvocationClient, key string, ratingResponse map[string]interface{}, sapFosstars *Fosstars, artifactsRatingsMap map[string]*Rating,
	modelRatingsDefinitionIdMap map[string]struct{}) (bool, bool) {
	var badRatingsOccured, responsesInRequestedState bool
	metadata := ratingResponse["metadata"].(map[string]interface{})
	status := metadata["status"]
	log.Entry().Infof("Response status is %v", status)
	if status == "REQUESTED" {
		log.Entry().Infof("Rating request for the artifact: %v is in REQUESTED status trying again.", key)
		Value := createUnknownRating()
		artifactsRatingsMap[key] = Value
		responsesInRequestedState = true
		return badRatingsOccured, responsesInRequestedState
	}
	Values := ratingResponse["ratings"].([]interface{})
	if len(Values) > 0 {
		ratingMap := Values[0].(map[string]interface{})
		Value := &Rating{}
		mapstructure.Decode(ratingMap, &Value)
		log.Entry().Infof("artifact: %v, rating Value: %v, rating Label: %v, ModelRatingDefinitionId: %v", key, Value.Value, Value.Label, Value.ModelRatingDefinitionId)
		artifactsRatingsMap[key] = Value

		if len(Value.ModelRatingDefinitionId) > 0 {
			modelRatingsDefinitionIdMap[Value.ModelRatingDefinitionId] = struct{}{}
		}

		if checkForBadRatings(sapFosstars, key, Value) {
			badRatingsOccured = true
		}
	} else {
		Value := createUnknownRating()
		artifactsRatingsMap[key] = Value
		log.Entry().Infof("artifact: %v , rating Value: none , rating Label: UNKNOWN", key)
	}
	return badRatingsOccured, responsesInRequestedState
}

func (i *InvocationClient) parseRatings(sapFosstars *Fosstars, ratings []byte) (map[string]*Rating, bool, bool, error) {
	// We only know our rating-level keys are strings
	allRatingsMap := make(map[string]interface{})
	// Decode JSON into our map
	err := json.Unmarshal(ratings, &allRatingsMap)
	if err != nil {
		return nil, false, false, err
	}
	var badRatingsOccured bool
	var responsesInRequestedState bool
	artifactsRatingsMap := make(map[string]*Rating)
	modelRatingsDefinitionIdMap := make(map[string]struct{})
	for key, value := range allRatingsMap {
		if ratingResponse, ok := value.(map[string]interface{}); ok {
			badRatings, responses := processRatingResponse(i, key, ratingResponse, sapFosstars, artifactsRatingsMap, modelRatingsDefinitionIdMap)
			badRatingsOccured = badRatingsOccured || badRatings
			responsesInRequestedState = responsesInRequestedState || responses
		} else {
			log.Entry().Infof("artifact: %v, Value : %v", key, value)
		}
	}

	i.modelRatingsDefinitionIdMap = modelRatingsDefinitionIdMap
	i.allRatingsFromFosstars = allRatingsMap

	return artifactsRatingsMap, badRatingsOccured, responsesInRequestedState, nil

}

func checkForBadRatings(sapFosstars *Fosstars, artifact string, Value *Rating) bool {
	var isArtifactExcluded bool
	var badRatingsOccured bool

	isArtifactExcluded = checkForArtifactExclusion(sapFosstars, artifact)
	// skip the artifact if it is in the exclude list
	if isArtifactExcluded {
		return false
	}

	// check the label threshold
	failedLabelThreshold := false
	switch sapFosstars.RatingLabelThreshold {
	case "BAD":
		failedLabelThreshold = Value.Label == "BAD"
	case "MODERATE":
		failedLabelThreshold = Value.Label == "BAD" || Value.Label == "MODERATE"
	}
	if failedLabelThreshold {
		log.Entry().Errorf("%v rating found for the artifact %v but the threshold is %v", Value.Label, artifact, sapFosstars.RatingLabelThreshold)
		badRatingsOccured = true
	}

	// check the score threshold
	if Value.Value < float64(sapFosstars.RatingValueThreshold) {
		log.Entry().Errorf("The rating score %v is less than the configured threshold %v for the artifact %v", Value.Value, sapFosstars.RatingValueThreshold, artifact)
		badRatingsOccured = true
	}

	// check if unclear ratings are not acceptable
	if sapFosstars.FailOnUnclearRatings && Value.Label == "UNCLEAR" {
		log.Entry().Errorf("UNCLEAR rating found for the artifact %v", artifact)
		badRatingsOccured = true
	}

	return badRatingsOccured
}

func checkForArtifactExclusion(sapFosstars *Fosstars, artifact string) bool {
	var isArtifactExcluded bool
	if len(sapFosstars.ExcludedLibraries) > 0 && slices.Contains(sapFosstars.ExcludedLibraries, artifact) {
		log.Entry().Infof("%v is part of the excludedLibraries list", artifact)
		isArtifactExcluded = true

	}
	return isArtifactExcluded
}
