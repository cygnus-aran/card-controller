package value_objects

// CardData represents the card information (PAN & expiration date)
type CardData struct {
	Pan  string `json:"pan"`
	Date string `json:"date"`
}

// IsValid validates the CardData
func (c CardData) IsValid() bool {
	return c.Pan != "" && c.Date != ""
}
