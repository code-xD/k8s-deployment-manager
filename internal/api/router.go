package api

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// SetupRouter creates and configures a new Gin router with Zap logging middleware and routes.
func SetupRouter(log *zap.Logger) *gin.Engine {
	router := gin.New()

	// Zap request logging (replaces gin.Logger())
	router.Use(ginzap.Ginzap(log, time.RFC3339, true))
	// Zap panic recovery with stack trace
	router.Use(ginzap.RecoveryWithZap(log, true))

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", Ping)
	}

	return router
}

// Ping godoc
// @Summary      Ping endpoint
// @Description  Returns a pong message to verify the API is working
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
