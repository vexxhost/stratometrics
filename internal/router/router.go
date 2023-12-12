package router

import (
	"github.com/gin-gonic/gin"

	"github.com/vexxhost/stratometrics/internal/api/v1alpha1"
	"github.com/vexxhost/stratometrics/internal/clickhousedb"
	"github.com/vexxhost/stratometrics/internal/keystoneauth"
)

func NewRouter(db *clickhousedb.Database) *gin.Engine {
	r := gin.Default()
	r.Use(keystoneauth.Middleware())

	v1alpha1Router := r.Group("/v1alpha1")
	v1alpha1.SetupRoutes(db, v1alpha1Router)

	return r
}
