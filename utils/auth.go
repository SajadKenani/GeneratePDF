package utils

import (
	"crypto/subtle"
	"encoding/base64"
	"math/rand"

	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/scrypt"
)

func getPDFUIPassword() string {
	// Prefer a dedicated password for the PDF UI, but fall back to CLIENT_PASSWORD.
	if v := strings.TrimSpace(os.Getenv("PDF_UI_PASSWORD")); v != "" {
		return v
	}

	cfg := getConfig()
	if cfg != nil {
		// Config currently doesn't define PDF_UI_PASSWORD; keep for future-proofing.
		// Intentionally empty.
	}

	return strings.TrimSpace(os.Getenv("CLIENT_PASSWORD"))
}

// PDFUILoginHandler authenticates with a single password from env and returns a JWT.
// This is intended for the built-in /pdf/ui page.
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
func GenerateStreamToken(idOrName string, ttl time.Duration) (string, error) {
	idOrName = strings.TrimSpace(idOrName)
	if idOrName == "" {
		return "", fmt.Errorf("stream id is empty")
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	expiresAt := time.Now().Add(ttl).Unix()
	claims := &StreamClaims{
		Purpose: "stream",
		Video:   idOrName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecretKey())
}
func ValidateStreamToken(tokenString string, expectedIDOrName string) error {
	tokenString = strings.TrimSpace(tokenString)
	expectedIDOrName = strings.TrimSpace(expectedIDOrName)
	if tokenString == "" {
		return fmt.Errorf("missing token")
	}
	if expectedIDOrName == "" {
		return fmt.Errorf("missing stream id")
	}

	claims := &StreamClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return getJWTSecretKey(), nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("invalid token")
	}
	if claims.Purpose != "stream" {
		return fmt.Errorf("invalid token purpose")
	}
	if claims.Video != expectedIDOrName {
		return fmt.Errorf("token does not match requested video")
	}
	return nil
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
func LoginHandler(ctx *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Bind the incoming JSON body to the `loginData` struct
	if err := ctx.BindJSON(&loginData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON provided"})
		return
	}

	username := loginData.Username
	password := loginData.Password

	cfg := getConfig()
	var clientUsername = os.Getenv("CLIENT_USERNAME")
	var clientPassword = os.Getenv("CLIENT_PASSWORD")
	if cfg != nil {
		if cfg.ClientUsername != "" {
			clientUsername = cfg.ClientUsername
		}
		if cfg.ClientPassword != "" {
			clientPassword = cfg.ClientPassword
		}
	}

	// Dummy username and password, replace with database check in real apps
	if username == clientUsername && password == clientPassword {
		token, err := GenerateJWT(username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError,
				gin.H{"error": "Failed to generate token"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"token": token})
	} else {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}
func HashPassword(password string) (string, error) {
	encoded, _, err := HashPasswordArgon2id(password)
	return encoded, err
}
func HashPasswordArgon2id(password string) (encodedHash string, saltBase64 string, err error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}

	// Keep parameters aligned with VerifyPassword.
	iters := uint32(1)
	memory := uint32(64 * 1024)
	threads := uint8(4)
	keyLen := uint32(32)

	hash := argon2.IDKey([]byte(password), salt, iters, memory, threads, keyLen)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, iters, threads, saltB64, hashB64)
	return encoded, base64.StdEncoding.EncodeToString(salt), nil
}
func VerifyPasswordScrypt(password, saltBase64, hashBase64 string) (bool, error) {
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return false, err
	}

	computed, err := scrypt.Key([]byte(password), salt, 16384, 8, 1, len(expectedHash))
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(expectedHash, computed) == 1, nil
}
func VerifyPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
	hash, _ := base64.RawStdEncoding.DecodeString(parts[5])

	computed := argon2.IDKey(
		[]byte(password),
		salt,
		1,       // iterations
		64*1024, // memory
		4,       // threads
		uint32(len(hash)),
	)

	return subtle.ConstantTimeCompare(hash, computed) == 1
}
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
