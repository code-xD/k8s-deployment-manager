package api

import (
	"time"

	"github.com/code-xd/k8s-deployment-manager/internal/api/handlers"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/ports"
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

	// Initialize handlers and auto-populate routes
	initHandlers(router)

	return router
}

// initHandlers initializes all handler instances and registers their routes
func initHandlers(router *gin.Engine) {
	// Get all handlers that implement the Handler interface
	allHandlers := getHandlers()

	// Collect all route definitions from handlers
	allRoutes := []dto.RouteDefinition{}
	for _, handler := range allHandlers {
		allRoutes = append(allRoutes, handler.GetRoutes()...)
	}

	// Register routes on the router
	for _, routeDef := range allRoutes {
		registerRouteOnRouter(router, routeDef.Method, routeDef.Path, routeDef.Handler)
	}
}

// getHandlers returns all handler instances that implement the Handler interface
func getHandlers() []ports.Handler {
	return []ports.Handler{
		handlers.NewDeploymentHandler(),
		handlers.NewHealthHandler(),
	}
}

// registerRouteOnRouter registers a route on the Gin router based on HTTP method
func registerRouteOnRouter(router *gin.Engine, method string, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		router.GET(path, handler)
	case "POST":
		router.POST(path, handler)
	case "PUT":
		router.PUT(path, handler)
	case "PATCH":
		router.PATCH(path, handler)
	case "DELETE":
		router.DELETE(path, handler)
	case "OPTIONS":
		router.OPTIONS(path, handler)
	case "HEAD":
		router.HEAD(path, handler)
	default:
		// Log unsupported method or handle error
		// For now, we'll skip it
	}
}
