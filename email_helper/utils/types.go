package utils

import "time"

// ContractData carries the metadata needed for email subject/body context.
type ContractData struct {
	NumID          int       `json:"numId"`
	CreatedAt      time.Time `json:"createdAt"`
	ItemAmount     float64   `json:"itemAmount"`
	PaymentMethod  string    `json:"paymentMethod"`
	Realm          string    `json:"realm"`
	Currency       string    `json:"currency"`
	CurrencyAmount float64   `json:"currencyAmount"`
	UserName       string    `json:"userName"`
	UserPhone      string    `json:"userPhone"`
	UserEmail      string    `json:"userEmail"`
	Item           string    `json:"item"`
	Insurance      bool      `json:"insurance"`
}

// AttachmentProps mirrors the JSON shape the UI sends for base64 PDF attachments.
type AttachmentProps struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	ContentType string `json:"contentType"`
}
