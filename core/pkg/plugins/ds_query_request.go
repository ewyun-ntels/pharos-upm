package plugins

import (
	"encoding/json"
	"io"
)

type DsQuery struct {
	ID             string `json:"id"`
	DatasourceName string `json:"datasourceName"`
	SQL            string `json:"sql"`
	Timeout        int    `json:"timeout"`
}

type DsQueryRequest struct {
	Queries []DsQuery `json:"queries"`
}

func (dsQueryRequest *DsQueryRequest) set(reader io.Reader) error {
	if body, err := io.ReadAll(reader); err != nil {
		return err
	} else if err := json.Unmarshal(body, &dsQueryRequest); err != nil {
		return err
	}

	return nil
}
