//go:build unit
// +build unit

package dwc

import "testing"

func TestOverallSuccessPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		generateWatchResults func() []watchResult
		wantErr              bool
	}{
		{
			name: "call that passes the policy",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(true)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(true)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			wantErr: false,
		},
		{
			name: "call that does not pass the policy",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(true)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(false)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			wantErr: true,
		},
	}
	pol := OverallSuccessPolicy()
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := pol(testCase.generateWatchResults())
			if (err != nil) != testCase.wantErr {
				t.Fatalf("OverallSuccessPolicy() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

func TestAtLeastOneSuccessfulDeploymentPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		generateWatchResults func() []watchResult
		wantErr              bool
	}{
		{
			name: "call that passes the policy",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(true)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(true)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			wantErr: false,
		},
		{
			name: "call that passes the policy 2",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(false)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(true)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			wantErr: false,
		},
		{
			name: "call that does not pass the policy",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(false)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(false)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			wantErr: true,
		},
	}
	pol := AtLeastOneSuccessfulDeploymentPolicy()
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := pol(testCase.generateWatchResults())
			if (err != nil) != testCase.wantErr {
				t.Fatalf("AtLeastOneSuccessfulDeploymentPolicy() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

func TestSubsetSuccessPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		generateWatchResults func() []watchResult
		generatePolicy       func() StageWatchPolicy
		wantErr              bool
	}{
		{
			name: "call that passes the policy",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(true)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(true)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			generatePolicy: func() StageWatchPolicy {
				subsetStages := []string{"stage1", "stage2"}
				return SubsetSuccessPolicy(subsetStages)
			},
			wantErr: false,
		},
		{
			name: "call that passes the policy 2",
			generateWatchResults: func() []watchResult {
				res1, res2, res3 := &mockWatchResult{}, &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(true)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(true)
				res2.On("getStageName").Return("stage2")
				res3.On("succeeded").Return(false)
				res3.On("getStageName").Return("stage3")
				return []watchResult{res1, res2, res3}
			},
			generatePolicy: func() StageWatchPolicy {
				subsetStages := []string{"stage1", "stage2"}
				return SubsetSuccessPolicy(subsetStages)
			},
			wantErr: false,
		},
		{
			name: "call that passes the policy 3",
			generateWatchResults: func() []watchResult {
				res1, res2, res3 := &mockWatchResult{}, &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(false)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(false)
				res2.On("getStageName").Return("stage2")
				res3.On("succeeded").Return(false)
				res3.On("getStageName").Return("stage3")
				return []watchResult{res1, res2, res3}
			},
			generatePolicy: func() StageWatchPolicy {
				var subsetStages []string
				return SubsetSuccessPolicy(subsetStages)
			},
			wantErr: false,
		},
		{
			name: "call that does not pass the policy",
			generateWatchResults: func() []watchResult {
				res1, res2, res3 := &mockWatchResult{}, &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(false)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(true)
				res2.On("getStageName").Return("stage2")
				res3.On("succeeded").Return(false)
				res3.On("getStageName").Return("stage3")
				return []watchResult{res1, res2, res3}
			},
			generatePolicy: func() StageWatchPolicy {
				subsetStages := []string{"stage1", "stage2"}
				return SubsetSuccessPolicy(subsetStages)
			},
			wantErr: true,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pol := testCase.generatePolicy()
			err := pol(testCase.generateWatchResults())
			if (err != nil) != testCase.wantErr {
				t.Fatalf("SubsetSuccessPolicy() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

func TestAlwaysPassPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		generateWatchResults func() []watchResult
		wantErr              bool
	}{
		{
			name: "call that passes the policy",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(true)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(true)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			wantErr: false,
		},
		{
			name: "call that passes the policy 2",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(true)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(false)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			wantErr: false,
		},
		{
			name: "call that passes the policy 3",
			generateWatchResults: func() []watchResult {
				res1, res2 := &mockWatchResult{}, &mockWatchResult{}
				res1.On("succeeded").Return(false)
				res1.On("getStageName").Return("stage1")
				res2.On("succeeded").Return(false)
				res2.On("getStageName").Return("stage2")
				return []watchResult{res1, res2}
			},
			wantErr: false,
		},
	}
	pol := AlwaysPassPolicy()
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := pol(testCase.generateWatchResults())
			if (err != nil) != testCase.wantErr {
				t.Fatalf("OverallSuccessPolicy() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}
