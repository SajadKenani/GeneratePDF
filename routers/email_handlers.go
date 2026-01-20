package routers

// import (
// 	"Backend/utils"
// 	"context"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// func handleSendDepositEmail(c *gin.Context) {
// 	contractData, err := parseContractDataFromJSON(c)
// 	if err != nil {
// 		utils.HandleError(c, err, "invalid request body (createdAt must be RFC3339)", http.StatusBadRequest)
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Minute)
// 	defer cancel()

// 	attachments, err := utils.SendDepositEmail(ctx, contractData)
// 	if err != nil {
// 		utils.HandleError(c, err, "failed to send deposit email", http.StatusInternalServerError)
// 		return
// 	}

// 	utils.RespondOK(c, gin.H{
// 		"attachments": attachments,
// 	})
// }
