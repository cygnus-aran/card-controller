package types

type CheckCardStatusRequest struct {
	CardID             string `json:"cardId"`
	MerchantIdentifier string `json:"merchantIdentifier"`
}
