package api

import (
	"time"

	"github.com/code-xd/k8s-deployment-manager/internal/api/handlers"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/ports"
	portsapi "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/apiService"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// SetupRouter creates and configures a new Gin router with Zap logging middleware and routes.
// It accepts dependencies (services) that are injected into handlers.
func SetupRouter(
	log *zap.Logger,
	deploymentRequest portsapi.DeploymentRequest,
	deployment portsapi.Deployment,
	userRepo portsdb.User,
	deploymentRequestRepo portsdb.DeploymentRequest,
) *gin.Engine {
	router := gin.New()

	// Zap request logging (replaces gin.Logger())
	router.Use(ginzap.Ginzap(log, time.RFC3339, true))
	// Zap panic recovery with stack trace
	router.Use(ginzap.RecoveryWithZap(log, true))

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize handlers with injected dependencies and auto-populate routes
	initHandlers(
		router,
		deploymentRequest,
		deployment,
		userRepo,
		deploymentRequestRepo,
		log,
	)

	return router
}

// initHandlers initializes all handler instances and registers their routes
func initHandlers(
	router *gin.Engine,
	deploymentRequest portsapi.DeploymentRequest,
	deployment portsapi.Deployment,
	userRepo portsdb.User,
	deploymentRequestRepo portsdb.DeploymentRequest,
	log *zap.Logger,
) {
	// Get all handlers that implement the Handler interface with injected dependencies
	allHandlers := getHandlers(
		deploymentRequest,
		deployment,
		userRepo,
		deploymentRequestRepo,
		log,
	)

	// Collect all route definitions from handlers
	allRoutes := []dto.RouteDefinition{}
	for _, handler := range allHandlers {
		routes := handler.GetRoutes()
		allRoutes = append(allRoutes, routes...)
	}

	// Register routes on the router
	for _, routeDef := range allRoutes {
		// Build handler chain: middlewares -> handler
		var handler gin.HandlerFunc = routeDef.Handler

		// Apply middlewares in reverse order (last middleware wraps handler)
		// So middlewares are applied: mw1 -> mw2 -> ... -> handler
		for i := len(routeDef.Middlewares) - 1; i >= 0; i-- {
			mw := routeDef.Middlewares[i]
			currentHandler := handler
			handler = func(c *gin.Context) {
				mw(c)
				if !c.IsAborted() && currentHandler != nil {
					currentHandler(c)
				}
			}
		}

		registerRouteOnRouter(router, routeDef.Method, routeDef.Path, handler)
	}
}

// getHandlers returns all handler instances that implement the Handler interface
// Dependencies are injected via constructor functions
func getHandlers(
	deploymentRequest portsapi.DeploymentRequest,
	deployment portsapi.Deployment,
	userRepo portsdb.User,
	deploymentRequestRepo portsdb.DeploymentRequest,
	log *zap.Logger,
) []ports.Handler {
	return []ports.Handler{
		handlers.NewDeploymentRequestHandler(
			deploymentRequest,
			userRepo,
			deploymentRequestRepo,
			log,
		),
		handlers.NewDeploymentHandler(
			deployment,
			userRepo,
			log,
		),
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
