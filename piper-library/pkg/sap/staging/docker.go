package staging

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
)

type AuthEntry struct {
	Auth string `json:"auth,omitempty"`
}

func CreateDockerConfigJSON(registryURL, username, password, configPath string, utils ServiceUtils) (string, error) {

	filePath := ".pipeline/docker/config.json" // filename must be 'config.json' according to the docker CLI docs

	err := utils.MkdirAll(filepath.Dir(filePath), 0777)

	if err != nil {
		return "", err
	}

	dockerConfig := map[string]interface{}{}
	if exists, _ := utils.FileExists(configPath); exists {
		dockerConfigContent, err := utils.FileRead(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file '%v': %w", configPath, err)
		}

		err = json.Unmarshal(dockerConfigContent, &dockerConfig)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal json file '%v': %w", configPath, err)
		}
	}

	credentialsBase64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", username, password)))
	dockerAuth := AuthEntry{Auth: credentialsBase64}

	if dockerConfig["auths"] == nil {
		dockerConfig["auths"] = map[string]AuthEntry{registryURL: dockerAuth}
	} else {
		authEntries, ok := dockerConfig["auths"].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("failed to read authentication entries from file '%v': format invalid", configPath)
		}
		authEntries[registryURL] = dockerAuth
		dockerConfig["auths"] = authEntries
	}

	jsonResult, err := json.Marshal(dockerConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Docker config.json: %w", err)
	}

	err = utils.FileWrite(filePath, jsonResult, 0666)
	if err != nil {
		return "", fmt.Errorf("failed to write Docker config.json: %w", err)
	}

	return filePath, nil
}
