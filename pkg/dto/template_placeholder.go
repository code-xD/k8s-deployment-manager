package dto

// CreateTemplateData holds the values for deployment template substitution
type CreateTemplateData struct {
	Name                string
	Namespace           string
	Identifier          string
	Image               string
	UserID              string
	RequestID           string
	DeploymentRequestID string
	// HasCustomHTML is true when metadata contains inline HTML (html/content/body); enables ConfigMap volume mount
	HasCustomHTML bool
}
