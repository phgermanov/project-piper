package cumulus

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type fileWriter interface {
	FileWrite(filename string, data []byte, perm os.FileMode) error
}

// ReleaseStatus defines release status for Cumulus
// as per https://github.tools.sap/P4TEAM/cumulus-file-schema/blob/main/schema/v1/.status-log/release/release-status.schema.json
type ReleaseStatus struct {
	Status                  string `json:"releaseStatus" validate:"required"`
	SAPJiraMicroDeliveryKey string `json:"sapJiraMicrodeliveryKey,omitempty"`
	SiriusDeliveryId        string `json:"siriusDeliveryId,omitempty" validate:"min=32,max=32"`
	SelectionType           string `json:"selectionType,omitempty"`
}

func (r *ReleaseStatus) ToFile(files fileWriter, now time.Time) error {
	fileName := fmt.Sprintf("release-status-%v.json", now.Format("20060102150405"))

	// ignore error since format is in our hands
	data, _ := json.Marshal(r)

	if err := files.FileWrite(fileName, data, 0o666); err != nil {
		return fmt.Errorf("failed to write %v: %w", fileName, err)
	}

	return nil
}
