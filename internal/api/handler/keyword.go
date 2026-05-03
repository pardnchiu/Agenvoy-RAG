package apiHandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/agenvoy/kuradb/internal/database"
	databaseHandler "github.com/agenvoy/kuradb/internal/database/handler"
	"github.com/agenvoy/kuradb/internal/utils/segmenter"
)

const (
	defaultLimit = 10
	maxLimit     = 100
)

type Match struct {
	Chunk   int    `json:"chunk"`
	Content string `json:"content"`
}

type Group struct {
	Source  string  `json:"source"`
	Matches []Match `json:"matches"`
}

func Keyword(dbs map[string]*database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.GetString("db")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "db is required",
			})
			return
		}

		q := c.Query("q")
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "q is required",
			})
			return
		}

		db := dbs[name]
		limit := queryLimit(c)

		keywords, err := segmenter.Tokenize(q)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		var results []databaseHandler.FileRow
		if len(keywords) != 0 {
			results, err = databaseHandler.SearchKeyword(db, c.Request.Context(), keywords, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"results": group(results),
		})
	}
}

func group(flat []databaseHandler.FileRow) []Group {
	if len(flat) == 0 {
		return []Group{}
	}
	idx := make(map[string]int, len(flat))
	groups := make([]Group, 0)
	for _, h := range flat {
		i, ok := idx[h.Source]
		if !ok {
			idx[h.Source] = len(groups)
			groups = append(groups, Group{Source: h.Source})
			i = idx[h.Source]
		}
		groups[i].Matches = append(groups[i].Matches, Match{
			Chunk:   h.Chunk,
			Content: h.Content,
		})
	}
	return groups
}

func queryLimit(c *gin.Context) int {
	raw := c.Query("limit")
	if raw == "" {
		return defaultLimit
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 || v > maxLimit {
		return defaultLimit
	}
	return v
}
