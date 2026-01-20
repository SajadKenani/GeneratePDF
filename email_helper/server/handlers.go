package server

import (
	"context"
	"emailhelper/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type contractRequest struct {
	NumID          int                     `json:"numId"`
	CreatedAt      string                  `json:"createdAt"`
	ItemAmount     float64                 `json:"itemAmount"`
	PaymentMethod  string                  `json:"paymentMethod"`
	Realm          string                  `json:"realm"`
	Currency       string                  `json:"currency"`
	CurrencyAmount float64                 `json:"currencyAmount"`
	UserName       string                  `json:"userName"`
	UserPhone      string                  `json:"userPhone"`
	UserEmail      string                  `json:"userEmail"`
	Item           string                  `json:"item"`
	Insurance      bool                    `json:"insurance"`
	Attachments    []utils.AttachmentProps `json:"attachments"`
}

func parseContractDataFromJSON(c *gin.Context) (utils.ContractData, []utils.AttachmentProps, error) {
	var req contractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return utils.ContractData{}, nil, err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(req.CreatedAt))
	if err != nil {
		createdAt, err = time.Parse(time.RFC3339, strings.TrimSpace(req.CreatedAt))
	}
	if err != nil {
		return utils.ContractData{}, nil, err
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
	}, req.Attachments, nil
}

func handleSendDepositEmail(c *gin.Context) {
	contractData, attachments, err := parseContractDataFromJSON(c)
	if err != nil {
		utils.HandleError(c, err, "invalid request body (createdAt must be RFC3339)", http.StatusBadRequest)
		return
	}

	if len(attachments) == 0 {
		utils.HandleError(c, nil, "attachments are required (pre-rendered PDFs)", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := utils.SendDepositEmail(ctx, contractData, attachments); err != nil {
		utils.HandleError(c, err, "failed to send deposit email", http.StatusInternalServerError)
		return
	}

	utils.RespondOK(c, gin.H{
		"attachments": attachments,
	})
}
