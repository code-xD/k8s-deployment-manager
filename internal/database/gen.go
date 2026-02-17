//go:build ignore
// +build ignore

package main

import (
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	"gorm.io/gen"
)

func main() {
	// Initialize the generator
	g := gen.NewGenerator(gen.Config{
		OutPath:       "./internal/database/query",
		Mode:          gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true,
	})

	// Generate the code from model structs (no database connection needed)
	g.ApplyBasic(
		g.GenerateModel(&models.User{}),
		g.GenerateModel(&models.DeploymentRequest{}),
		g.GenerateModel(&models.Deployment{}),
	)

	// Execute the generator
	g.Execute()
}
