package model

type Task struct {
	Schema       string `json:"$schema"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	FriendlyName string `json:"friendlyName"`
	Description  string `json:"description"`
	HelpMarkDown string `json:"helpMarkDown"`
	Category     string `json:"category"`
	Author       string `json:"author"`
	Version      struct {
		Major int `json:"Major"`
		Minor int `json:"Minor"`
		Patch int `json:"Patch"`
	} `json:"version"`
	InstanceNameFormat string  `json:"instanceNameFormat"`
	Inputs             []Input `json:"inputs"`
	Groups             []struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		IsExpanded  bool   `json:"isExpanded"`
	} `json:"groups"`
	Execution struct {
		Node20_1 struct {
			Target string `json:"target"`
		} `json:"Node20_1"`
	} `json:"execution"`
	Messages struct {
		DownloadPiperFailedFromLocation string `json:"DownloadPiperFailedFromLocation"`
		PiperNotFoundInFolder           string `json:"PiperNotFoundInFolder"`
		PiperDownloadFailed             string `json:"PiperDownloadFailed"`
		VerifyPiperInstallation         string `json:"VerifyPiperInstallation"`
	} `json:"messages"`
	DataSourceBindings []struct {
		Target         string `json:"target"`
		EndpointId     string `json:"endpointId"`
		DataSourceName string `json:"dataSourceName"`
		ResultTemplate string `json:"resultTemplate"`
	} `json:"dataSourceBindings"`
}

type Input struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Label        string      `json:"label"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	Required     bool        `json:"required"`
	HelpMarkDown string      `json:"helpMarkDown"`
	GroupName    string      `json:"groupName,omitempty"`
}
