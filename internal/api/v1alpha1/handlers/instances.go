package handlers

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	"gorm.io/gorm"

	"github.com/vexxhost/stratometrics/internal/database"
	"github.com/vexxhost/stratometrics/internal/database/types"
)

type InstancesRequest struct {
	ProjectID string    `form:"project_id"`
	From      time.Time `form:"from"`
	To        time.Time `form:"to"`
	GroupBy   []string  `form:"group_by,default=type" binding:"dive,oneof=type project_id state image"`
}

type InstancesPeriodResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type InstancesResponse struct {
	Period  InstancesPeriodResponse  `json:"period"`
	Results []database.InstanceUsage `json:"results"`
	Units   string                   `json:"units"`
}

func GetInstanceUsage(c *gin.Context, db *gorm.DB) {
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
	projectId := types.ParseUUID(project.ID)

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

		projectId = types.ParseUUID(req.ProjectID)
	}

	if slices.Contains(req.GroupBy, "project_id") {
		if !slices.Contains(roleNames, "admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is not an admin of the project"})
			return
		}

		projectId = types.EmptyUUID
	}

	projectId = types.ParseUUID("651c7592-6ebc-40f4-9c20-71d43b1e41b1")

	evts, err := database.GetInstancesUsageForProject(
		db,
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
		evts = []database.InstanceUsage{}
	}

	c.JSON(http.StatusOK, InstancesResponse{
		Period: InstancesPeriodResponse{
			From: req.From,
			To:   req.To,
		},
		Results: evts,
		Units:   "seconds",
	})
}
