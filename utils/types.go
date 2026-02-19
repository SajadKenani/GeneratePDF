package utils

import (
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type StoredAccount struct {
	ID          string
	PhoneNumber string
	Password    string
	Name        string
	Salt        string
	Limits      int
	Status      string
	Action_date string
	Is_active   bool
	Email       string
}
type claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}
type adminClaims struct {
	AdminID string `json:"admin_id"`
	jwt.StandardClaims
}
type Value struct {
	Currency int `json:"currency"`
}
type Claims struct {
	Authorized bool   `json:"authorized"`
	User       string `json:"user"`
	jwt.StandardClaims
}
type StreamClaims struct {
	Purpose string `json:"purpose"`
	Video   string `json:"video"`
	jwt.StandardClaims
}
var (
	cfgOnce   sync.Once
	cachedCfg *Config
)
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
	Package        string    `json:"package"`
	Item           string    `json:"item"`
	Insurance      bool      `json:"insurance"`
}
type AttachmentProps struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	ContentType string `json:"contentType"`
}
type pdfOptions struct {
	paperWidthIn  float64
	paperHeightIn float64
	marginIn      float64
	printBG       bool
}