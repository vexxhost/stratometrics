package v1alpha1

import (
	"github.com/gin-gonic/gin"

	"github.com/vexxhost/stratometrics/internal/api/v1alpha1/handlers"
)

func SetupRoutes(r *gin.RouterGroup) {
	r.GET("/instances", handlers.GetInstanceUsage)
}
