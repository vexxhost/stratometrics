package v1alpha1

import (
	"github.com/gin-gonic/gin"

	"github.com/vexxhost/stratometrics/internal/api/v1alpha1/handlers"
	"github.com/vexxhost/stratometrics/internal/clickhousedb"
)

func SetupRoutes(db *clickhousedb.Database, r *gin.RouterGroup) {
	r.GET("/instances", func(ctx *gin.Context) {
		handlers.GetInstanceUsage(ctx, db)
	})
}
