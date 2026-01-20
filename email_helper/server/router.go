package server

import (
	"emailhelper/utils"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Start(port string) error {
	if _, err := utils.LoadConfig(); err != nil {
		log.Printf("config load warning: %v", err)
	}

	r := gin.Default()
	r.Use(gin.Logger(), gin.Recovery(), utils.RecoveryWithLogger())

	allowedOrigin := strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN"))
	corsCfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "accept", "origin", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range", "Accept-Ranges"},
		AllowCredentials: false,
	}
	if allowedOrigin == "" || allowedOrigin == "*" {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = []string{allowedOrigin}
	}
	r.Use(cors.New(corsCfg))

	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.POST("/api/email/deposit", handleSendDepositEmail)

	addr := ":" + port
	log.Printf("email helper listening on %s", addr)
	return http.ListenAndServe(addr, r)
}
