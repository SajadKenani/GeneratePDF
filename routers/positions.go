package routers

import (
	"Backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleGetPositions(c *gin.Context) {
	db, err := utils.LoadPositions()
	if err != nil {
		utils.HandleError(c, err, "failed to load positions", http.StatusInternalServerError)
		return
	}
	utils.RespondOK(c, db)
}

func HandleSavePositions(c *gin.Context) {
	var db utils.PositionsDB
	if err := c.ShouldBindJSON(&db); err != nil {
		utils.HandleError(c, err, "invalid body", http.StatusBadRequest)
		return
	}
	if err := utils.SavePositions(db); err != nil {
		utils.HandleError(c, err, "failed to save positions", http.StatusInternalServerError)
		return
	}
	utils.RespondOK(c, gin.H{"saved": true})
}