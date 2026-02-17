package dto

import "github.com/gin-gonic/gin"

// RouteDefinition represents a route definition that handlers can return
type RouteDefinition struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}
