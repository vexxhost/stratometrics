package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vexxhost/stratometrics/internal/clickhousedb"
)

type InstancesRequest struct {
	From    time.Time `form:"from"`
	To      time.Time `form:"to"`
	GroupBy []string  `form:"group_by,default=type"`
}

func GetInstanceUsage(c *gin.Context, db *clickhousedb.Database) {
	var req InstancesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.From.IsZero() {
		req.From = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	if req.To.IsZero() {
		req.To = time.Date(req.From.Year(), req.From.Month()+1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	}

	projectID, ok := c.Get("project_id")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing project_id"})
		return
	}

	evts, err := db.GetInstancesUsageForProject(
		c.Request.Context(),
		req.From,
		req.To,
		projectID.(string),
		req.GroupBy,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if evts == nil {
		evts = []clickhousedb.InstanceUsage{}
	}

	c.JSON(http.StatusOK, gin.H{
		"period": gin.H{
			"from": req.From,
			"to":   req.To,
		},
		"units":   "seconds",
		"results": evts,
	})
}
