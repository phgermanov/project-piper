//go:build unit
// +build unit

package versioning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitVersion(t *testing.T) {
	tt := []struct {
		version       string
		expected      Version
		expectedError string
	}{
		{version: "1", expected: Version{Major: "1", Minor: "0", Patch: "0"}},
		{version: "1", expected: Version{Major: "1", Minor: "0", Patch: "0"}},
		{version: "1.2", expected: Version{Major: "1", Minor: "2", Patch: "0"}},
		{version: "1.2.3", expected: Version{Major: "1", Minor: "2", Patch: "3"}},
		{version: "1.2.3-20200202", expected: Version{Major: "1", Minor: "2", Patch: "3", Timestamp: "20200202"}},
		{version: "1.2.3.20200202", expected: Version{Major: "1", Minor: "2", Patch: "3", Timestamp: "20200202"}},
		{version: "1.2.3.20200202+5c02f9d9", expected: Version{Major: "1", Minor: "2", Patch: "3", Timestamp: "20200202", CommitID: "5c02f9d9"}},
		{version: "1.2.3.20200202_5c02f9d9", expected: Version{Major: "1", Minor: "2", Patch: "3", Timestamp: "20200202", CommitID: "5c02f9d9"}},
		{version: "1.2.3.20200202.5c02f9d9", expected: Version{Major: "1", Minor: "2", Patch: "3", Timestamp: "20200202", CommitID: "5c02f9d9"}},
		{version: "1.2.3-2020-02-0304T010203UTC+5c02f9d9", expected: Version{Major: "1", Minor: "2", Patch: "3", Timestamp: "2020-02-0304T010203UTC", CommitID: "5c02f9d9"}},
		{version: "1.2.3-2020-02-0304T010203UTC+5c02f9d9", expected: Version{Major: "1", Minor: "2", Patch: "3", Timestamp: "2020-02-0304T010203UTC", CommitID: "5c02f9d9"}},
		{version: "master", expectedError: "invalid version: master"},
	}

	for _, test := range tt {

		version, err := SplitVersion(test.version)

		if len(test.expectedError) > 0 {
			assert.EqualError(t, err, test.expectedError)
		} else {
			assert.Equalf(t, test.expected, version, test.version)
			assert.NoError(t, err)
		}
	}
}
