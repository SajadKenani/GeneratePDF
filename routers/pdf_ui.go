package routers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func servePDFUI(c *gin.Context) {
	baseDir, _ := os.Getwd()

	pdfUIHTML, err := os.ReadFile(
		filepath.Join(baseDir, "index.html"),
	)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(pdfUIHTML))
}
