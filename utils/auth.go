package utils

import (
	"crypto/subtle"

	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func getPDFUIPassword() string {
	if v := strings.TrimSpace(os.Getenv("PDF_UI_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("CLIENT_PASSWORD"))
}
func PDFUILoginHandler(ctx *gin.Context) {
	var body struct {
		Password string `json:"password"`
	}
	if err := ctx.BindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON provided"})
		return
	}

	want := getPDFUIPassword()
	got := strings.TrimSpace(body.Password)
	if want == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "PDF UI password is not configured"})
		return
	}

	if subtle.ConstantTimeCompare([]byte(want), []byte(got)) != 1 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	token, err := GenerateJWT("pdf-ui")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
func getConfig() *Config {
	cfgOnce.Do(func() {
		c, err := LoadConfig()
		if err != nil {
			log.Printf("failed to load config: %v", err)
		}
		cachedCfg = c
	})
	return cachedCfg
}
func getJWTSecretKey() []byte {
	cfg := getConfig()
	secretKey := ""
	if cfg != nil {
		secretKey = cfg.JWTSecret
	}
	if secretKey == "" {
		secretKey = os.Getenv("JWT_SECRET_KEY")
	}
	if secretKey == "" {
		log.Panic("JWT secret is not configured")
	} else if len(secretKey) < 32 {
		fmt.Println("WARNING: JWT_SECRET_KEY should be at least 32 characters long")
	}
	return []byte(secretKey)
}
func GenerateJWT(username string) (string, error) {
	expiresAt := time.Now().Add(time.Hour * 99999).Unix()
	claims := &Claims{
		Authorized: true,
		User:       username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getJWTSecretKey())
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func AuthenticateMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized,
				gin.H{"error": "Authorization header format must be Bearer {token}"})
			ctx.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return getJWTSecretKey(), nil
		})
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		ctx.Set("user", claims.User)
		ctx.Next()
	}
}

