package utils

import (
	"sync"

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

// Claims wraps standard JWT registered claims and app-specific fields.
type Claims struct {
	Authorized bool   `json:"authorized"`
	User       string `json:"user"`
	jwt.StandardClaims
}

// StreamClaims is used for short-lived stream access tokens.
// These tokens are intended to be embedded in URLs (query string) so browsers can stream
// without setting Authorization headers.
type StreamClaims struct {
	Purpose string `json:"purpose"`
	Video   string `json:"video"`
	jwt.StandardClaims
}

var (
	cfgOnce   sync.Once
	cachedCfg *Config
)
