package routers

import (
	"Backend/utils"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type contractRequest struct {
	NumID          int     `json:"numId"`
	CreatedAt      string  `json:"createdAt"`
	ItemAmount     float64 `json:"itemAmount"`
	PaymentMethod  string  `json:"paymentMethod"`
	Realm          string  `json:"realm"`
	Currency       string  `json:"currency"`
	CurrencyAmount float64 `json:"currencyAmount"`
	UserName       string  `json:"userName"`
	UserPhone      string  `json:"userPhone"`
	UserEmail      string  `json:"userEmail"`
	Item           string  `json:"item"`
	Insurance      bool    `json:"insurance"`
}

func parseContractDataFromJSON(ctx *gin.Context) (utils.ContractData, error) {
	var req contractRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return utils.ContractData{}, err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(req.CreatedAt))
	if err != nil {
		createdAt, err = time.Parse(time.RFC3339, strings.TrimSpace(req.CreatedAt))
	}
	if err != nil {
		return utils.ContractData{}, err
	}

	return utils.ContractData{
		NumID:          req.NumID,
		CreatedAt:      createdAt,
		ItemAmount:     req.ItemAmount,
		PaymentMethod:  req.PaymentMethod,
		Realm:          req.Realm,
		Currency:       req.Currency,
		CurrencyAmount: req.CurrencyAmount,
		UserName:       req.UserName,
		UserPhone:      req.UserPhone,
		UserEmail:      req.UserEmail,
		Item:           req.Item,
		Insurance:      req.Insurance,
	}, nil
}

func handleGeneratePDF(c *gin.Context) {
	contractData, err := parseContractDataFromJSON(c)
	if err != nil {
		utils.HandleError(c, err, "invalid request body (createdAt must be RFC3339)", http.StatusBadRequest)
		return
	}

	// Use a background-based context so a client disconnect does not cancel Chrome mid-start.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	attachments, err := utils.GetDepositAttachments(ctx, contractData)
	if err != nil {
		utils.HandleError(c, err, "failed to generate pdf", http.StatusInternalServerError)
		return
	}

	utils.RespondOK(c, gin.H{
		"attachments": attachments,
	})
}

func handlePreviewContractHTML(c *gin.Context) {
	contractType := strings.TrimSpace(c.Param("type"))
	if contractType == "" {
		utils.HandleError(c, nil, "missing contract type", http.StatusBadRequest)
		return
	}

	contractData, err := parseContractDataFromJSON(c)
	if err != nil {
		utils.HandleError(c, err, "invalid request body (createdAt must be RFC3339)", http.StatusBadRequest)
		return
	}

	html, err := utils.RenderContractHTML(contractType, contractData)
	if err != nil {
		utils.HandleError(c, err, "failed to render contract html", http.StatusInternalServerError)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
