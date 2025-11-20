//go:build unit
// +build unit

package reportpipelinestatus

import (
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWriteLogFile(t *testing.T) {

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("could not get current working directory")
	}
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			t.Fatal("could not change working dir")
		}
	}(currentDir)

	tempDir, err := os.MkdirTemp("", "")

	if err != nil {
		t.Fatal("could not get tempDir")
	}
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal("could not get change current working directory")
	}

	type args struct {
		logFile  *[]byte
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		content string
	}{
		{
			name: "write log file and read content",
			args: args{
				logFile:  &[]byte{80, 105, 112, 101, 108, 105, 110, 101, 115, 32, 97, 114, 101, 32, 97, 119, 101, 115, 111, 109, 101},
				fileName: "testFile.log",
			},
			content: "Pipelines are awesome",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteLogFile(tt.args.logFile, tt.args.fileName)
			if _, err := os.Stat(tt.args.fileName); errors.Is(err, os.ErrNotExist) {
				t.Error("WriteLogFile(), could not write LogFile to disk")
			}
			content, err := os.ReadFile(tt.args.fileName)
			if err != nil {
				t.Fatalf("unable to read file: %v", err)
			}
			if string(content) != tt.content {
				t.Errorf("WriteLogFile(), got %v want %v", content, tt.content)
			}
		})
	}
}

func Test_categoryInList(t *testing.T) {
	type args struct {
		categories []string
		list       []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "category in list",
			args: args{
				categories: []string{"a"},
				list:       []string{"a", "b", "c"},
			},
			want: true,
		},
		{
			name: "category not in list ",
			args: args{
				categories: []string{"d"},
				list:       []string{"a", "b", "c"},
			},
			want: false,
		},
		{
			name: "category twice in list",
			args: args{
				categories: []string{"d"},
				list:       []string{"a", "d", "d"},
			},
			want: true,
		},
		{
			name: "empty list",
			args: args{
				categories: []string{"d"},
				list:       []string{},
			},
			want: false,
		},
		{
			name: "empty categories",
			args: args{
				categories: []string{},
				list:       []string{"d"},
			},
			want: false,
		},
		{
			name: "empty categories, empty list",
			args: args{
				categories: []string{},
				list:       []string{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategoryInList(tt.args.categories, tt.args.list); got != tt.want {
				t.Errorf("CategoryInList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMatches(t *testing.T) {
	regex := `(?i)\"(\w+)":\"(.*?)\"`

	type args struct {
		text  string
		regex string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "get matches",
			args: args{
				text:  "errorDetails{\"category\":\"undefined\",\"correlationId\":\"https://some.url.sap/\"}",
				regex: regex,
			},
			want: map[string]interface{}{"category": "undefined", "correlationId": "https://some.url.sap/"},
		},
		{
			name: "no matches available",
			args: args{
				text:  "No matches here",
				regex: regex,
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMatches(tt.args.text, tt.args.regex); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalcDuration(t *testing.T) {
	type args struct {
		pipelineTime time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "check time correctness",
			args: args{pipelineTime: time.Date(2022, time.March, 26, 17, 45, 10, 0, time.UTC)},
			want: int(time.Since(time.Date(2022, time.March, 26, 17, 45, 10, 0, time.UTC)).Milliseconds()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // allow 5 seconds delta
			if got := CalcDuration(tt.args.pipelineTime); !assert.InDelta(t, strToInt(got), tt.want, 5000) {
				t.Errorf("CalcDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func strToInt(str string) int {
	parsedInt, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return parsedInt
}
