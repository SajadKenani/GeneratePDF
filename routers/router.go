package routers

import (
	"Backend/utils"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Logger(), gin.Recovery(), utils.RecoveryWithLogger())
	r.Use(func(c *gin.Context) {
		log.Printf("→ [%s] %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
		log.Printf("← %d %s", c.Writer.Status(), c.FullPath())
	})

	r.Use(func(c *gin.Context) {
		p := c.Request.URL.Path
		if after, ok := strings.CutPrefix(p, "//uploads/"); ok {
			c.Request.URL.Path = "/uploads/" + after
		} else if after0, ok0 := strings.CutPrefix(p, "//compressed/"); ok0 {
			c.Request.URL.Path = "/compressed/" + after0
		}
		c.Next()
	})

	r.Static("/uploads", utils.UploadsDir())
	r.Static("/compressed", utils.CompressedDir())

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

	r.GET("/pdf/ui", servePDFUI)
	r.POST("/api/pdf-ui/login", utils.PDFUILoginHandler)
	auth := r.Group("/api")
	auth.Use(utils.AuthenticateMiddleware())
	{
		auth.POST("/pdf/generatePDF", handleGeneratePDF)
		auth.POST("/pdf/preview/:type", handlePreviewContractHTML)
		// auth.POST("/email/deposit", handleSendDepositEmail)
	}

	return r
}
