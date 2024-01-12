package v1alpha1

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/vexxhost/stratometrics/internal/api/v1alpha1/handlers"
)

func SetupRoutes(db *gorm.DB, r *gin.RouterGroup) {
	r.GET("/instances", func(ctx *gin.Context) {
		handlers.GetInstanceUsage(ctx, db)
	})
}
