package utils

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"
)

// TemplateRenderer loads and renders deployment templates.
// Methods are intended to be invoked in sequence: ExtractTemplateName -> Load -> Execute.
type TemplateRenderer[T any] struct {
	basePath     string
	templateName string
	content      string
}

// NewTemplateRenderer creates a new renderer with the given base path and image.
// ExtractTemplateName is called to derive the template folder from the image (e.g. "nginx:latest" -> "nginx").
func NewTemplateRenderer[T any](basePath, image string) *TemplateRenderer[T] {
	return &TemplateRenderer[T]{
		basePath:     basePath,
		templateName: extractTemplateName(image),
	}
}

// Load reads the template file from basePath/templates/<templateName>/deployment.yaml.
func (t *TemplateRenderer[T]) Load() error {
	tmplPath := path.Join(t.basePath, "templates", t.templateName, "deployment.yaml")
	content, err := os.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to read template %q: %w", tmplPath, err)
	}
	t.content = string(content)
	return nil
}

// Execute renders the template with the given data and returns the manifest string.
func (t *TemplateRenderer[T]) Execute(data T) (string, error) {
	tmpl, err := template.New("deployment").Parse(t.content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

// TemplateName returns the extracted template name (e.g. "nginx").
func (t *TemplateRenderer[T]) TemplateName() string {
	return t.templateName
}

// extractTemplateName derives the template folder name from the image.
// e.g. "nginx" -> "nginx", "nginx:latest" -> "nginx", "docker.io/library/nginx:latest" -> "nginx"
func extractTemplateName(image string) string {
	if idx := strings.LastIndex(image, ":"); idx > 0 {
		image = image[:idx]
	}
	if idx := strings.LastIndex(image, "/"); idx >= 0 {
		image = image[idx+1:]
	}
	return image
}
