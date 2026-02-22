package routers

import (
	"Backend/utils"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

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
	Package 	   string					`json:"package"`
	Item           string                  `json:"item"`
	Insurance      bool                    `json:"insurance"`
	Attachments    []utils.AttachmentProps `json:"attachments"`
}

func HandleStampPDF(ctx *gin.Context) {
	contractData, _, err := parseContractDataFromJSON(ctx)
	if err != nil {
		utils.HandleError(ctx, err, "invalid request body", http.StatusBadRequest)
		return
	}

	// Set UTF-8 response headers
	ctx.Header("Content-Type", "application/json; charset=utf-8")
	
	stamps, err := utils.GetStampsForRealm(contractData.Realm, contractData)
	if err != nil {
		utils.HandleError(ctx, err, "no positions configured for this contract", http.StatusInternalServerError)
		return
	}

	pdfPath := "/app/pdf/" + contractData.Realm + "/contract.pdf"
	result, err := utils.StampPDF(pdfPath, stamps)
	if err != nil {
		utils.HandleError(ctx, err, "failed to stamp pdf", http.StatusInternalServerError)
		return
	}

	encoded := base64.StdEncoding.EncodeToString(result)
	utils.RespondOK(ctx, gin.H{
		"attachments": []gin.H{{
			"filename":    fmt.Sprintf("contract-%d.pdf", contractData.NumID),
			"content":     encoded,
			"encoding":    "base64",
			"contentType": "application/pdf",
		}},
	})
}
func parseContractDataFromJSON(c *gin.Context) (utils.ContractData, []utils.AttachmentProps, error) {
	var req contractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return utils.ContractData{}, nil, err
	}

	// Validate UTF-8 encoding for Arabic text
	if !utf8.ValidString(req.Package) {
		return utils.ContractData{}, nil, fmt.Errorf("invalid UTF-8 in package field")
	}
	if !utf8.ValidString(req.UserName) {
		return utils.ContractData{}, nil, fmt.Errorf("invalid UTF-8 in userName field")
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
		Package:        req.Package, // Now validated as UTF-8
		Item:           req.Item,
		Insurance:      req.Insurance,
	}, req.Attachments, nil
}