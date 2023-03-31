package pagination

import (
	"github.com/d1msk1y/simple-go-chat-server/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type PaginatedResult struct {
	Items []models.Message
	Page  int
}

func Paginate(slice []models.Message, pageSize int, c *gin.Context) PaginatedResult {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return PaginatedResult{}
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(slice) {
		end = len(slice)
	}

	items := slice[start:end]

	return PaginatedResult{
		Items: items,
		Page:  page,
	}
}
