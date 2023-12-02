package router

import (
	"github.com/gin-gonic/gin"

	"github.com/vexxhost/stratometrics/internal/api/v1alpha1"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	v1alpha1Router := r.Group("/v1alpha1")
	v1alpha1.SetupRoutes(v1alpha1Router)

	return r
}
