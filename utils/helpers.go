package utils

import (
	"fmt"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"

)
func HandleError(ctx *gin.Context, err error, message string, statusCode ...int) {
	code := http.StatusInternalServerError
	if len(statusCode) > 0 {
		code = statusCode[0] // Use custom status code if provided
	}

	type pgError interface{ Code() string }
	if err != nil {
		if e, ok := err.(pgError); ok {
			switch e.Code() {
			case "23505":
				code = http.StatusConflict
			}
		}
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile) // Add this line to configure log output
	log.Printf("ERROR: %s | %v", message, err)

	// Return a JSON response with error information
	var details string
	if err != nil {
		details = err.Error()
	}

	ctx.AbortWithStatusJSON(code, gin.H{
		"error":   message,
		"details": details,
		"status":  false,
	})
}
func RespondOK(ctx *gin.Context, payload any) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   payload,
	})
}
func RecoveryWithLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("🔥 Panic recovered: %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"details": fmt.Sprintf("%v", r),
					"status":  false,
				})
			}
		}()
		c.Next()
	}
}
