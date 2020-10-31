package http

import (
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
)

func (s *Server) ticketUploadHandler(ctx *gin.Context) {
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	guild, err := strconv.ParseUint(ctx.Query("guild"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": "missing guild ID",
		})
		return
	}

	id, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": "missing ticket ID",
		})
		return
	}

	_, premium := ctx.GetQuery("premium")

	if err := s.UploadTicket(os.Getenv("S3_BUCKET"), premium, guild, id, body); err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
	} else {
		ctx.JSON(200, gin.H{})
	}
}
