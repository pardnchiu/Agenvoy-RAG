package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pardnchiu/KuraDB/internal/database"
)

func List(running string, reg *database.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		entries, err := reg.Load()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if entries == nil {
			entries = []database.Entry{}
		}
		c.JSON(http.StatusOK, gin.H{
			"running":    running,
			"registered": entries,
		})
	}
}
