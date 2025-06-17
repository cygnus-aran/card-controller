package types

type CheckCardStatusResponse struct {
	BlockType  string `json:"blockType"`
	Blocked    bool   `json:"blocked"`
	HasRetries bool   `json:"hasRetries"`
}
