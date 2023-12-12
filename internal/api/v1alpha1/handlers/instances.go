package handlers

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	"github.com/vexxhost/stratometrics/internal/clickhousedb"
)

type InstancesRequest struct {
	ProjectID string    `form:"project_id"`
	From      time.Time `form:"from"`
	To        time.Time `form:"to"`
	GroupBy   []string  `form:"group_by,default=type"`
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

	tokenData, ok := c.Get("token")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing token"})
		return
	}
	token := tokenData.(tokens.GetResult)

	project, err := token.ExtractProject()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	projectId := project.ID

	roles, err := token.ExtractRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	roleNames := []string{}
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	if !slices.Contains(roleNames, "member") {
		c.JSON(http.StatusForbidden, gin.H{"error": "user is not a member of the project"})
		return
	}

	if req.ProjectID != "" {
		if !slices.Contains(roleNames, "admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is not an admin of the project"})
			return
		}

		projectId = req.ProjectID
	}

	if slices.Contains(req.GroupBy, "project_id") {
		if !slices.Contains(roleNames, "admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is not an admin of the project"})
			return
		}

		req.ProjectID = ""
	}

	evts, err := db.GetInstancesUsageForProject(
		c.Request.Context(),
		req.From,
		req.To,
		projectId,
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
