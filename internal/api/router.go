package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apiHandler "github.com/pardnchiu/KuraDB/internal/api/handler"
	"github.com/pardnchiu/KuraDB/internal/database"
	"github.com/pardnchiu/KuraDB/internal/openai"
)

func Router(reg *database.Registry, dbs map[string]*database.DB, embedder openai.Embedder, qCache *openai.Cache) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api")
	api.GET("/health", apiHandler.Health())
	api.GET("/list", apiHandler.List(reg, dbs))
	api.GET("/semantic", queryDB(dbs), apiHandler.Semantic(dbs, embedder, qCache))
	api.GET("/keyword", queryDB(dbs), apiHandler.Keyword(dbs))

	return router
}

func queryDB(dbs map[string]*database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.Query("db")
		if db == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "db is required",
			})
			return
		}
		if _, ok := dbs[db]; !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("%q not exist", db),
			})
			return
		}

		c.Set("db", db)
		c.Next()
	}
}
