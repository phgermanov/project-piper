//go:build !release
// +build !release

package mocks

import (
	"time"

	mock "github.com/stretchr/testify/mock"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fosstars"
)

type InvocationClientInterface struct {
	mock.Mock
}

func (_m *InvocationClientInterface) GetRatings(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration, customTLSCerts []string) (map[string]*fosstars.Rating, bool, error) {
	ret := _m.Called(sapFosstars, fosstarsQueryServiceURL, artifacts)
	var r0 map[string]*fosstars.Rating
	if rf, ok := ret.Get(0).(func(*fosstars.Fosstars, string, []string, time.Duration) map[string]*fosstars.Rating); ok {
		r0 = rf(sapFosstars, fosstarsQueryServiceURL, artifacts, duration)
	} else {
		r0 = ret.Get(0).(map[string]*fosstars.Rating)
	}

	return r0, false, nil
}

func (_m *InvocationClientInterface) GetAllRatingsFromFosstars() map[string]interface{} {
	return nil
}

func (_m *InvocationClientInterface) GetModelRatingDefinitionDetails(fosstarsModelRatingsDefinitionsURL string, fosstarsClientSuffix string, customTLSCerts []string) (map[string]string, error) {
	return nil, nil
}
