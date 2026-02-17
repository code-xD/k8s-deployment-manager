package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

const (
	// MaxIdentifierLength is the maximum length for K8s resource names (63 chars)
	// We use 55 to be safe: 25 + 1 + 25 + 1 + 3 = 55
	maxNameLength      = 25
	maxNamespaceLength = 25
	randomDigits       = 3
)

// sanitizeForK8s sanitizes a string to be K8s-compliant
// Rules: lowercase, alphanumeric and hyphens only, must start/end with alphanumeric
func sanitizeForK8s(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	
	// Replace invalid characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	s = reg.ReplaceAllString(s, "-")
	
	// Remove consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")
	
	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")
	
	// Ensure it starts and ends with alphanumeric
	if len(s) > 0 {
		// If starts with non-alphanumeric (not a-z or 0-9), prepend 'a'
		firstChar := s[0]
		if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= '0' && firstChar <= '9')) {
			s = "a" + s
		}
		// If ends with non-alphanumeric (not a-z or 0-9), append '0'
		lastChar := s[len(s)-1]
		if !((lastChar >= 'a' && lastChar <= 'z') || (lastChar >= '0' && lastChar <= '9')) {
			s = s + "0"
		}
	} else {
		s = "deployment"
	}
	
	return s
}

// truncate truncates a string to max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// generateRandomDigits generates a random 3-digit number
func generateRandomDigits() (string, error) {
	// Generate random number between 100 and 999
	n, err := rand.Int(rand.Reader, big.NewInt(900))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%03d", n.Int64()+100), nil
}

// GenerateDeploymentIdentifier generates a K8s-compliant identifier
// Format: sanitize(truncate(name, 25))-sanitize(truncate(namespace, 25))-XXX
// where XXX is a random 3-digit number
func GenerateDeploymentIdentifier(name, namespace string) (string, error) {
	// Truncate and sanitize name
	sanitizedName := sanitizeForK8s(truncate(name, maxNameLength))
	
	// Truncate and sanitize namespace
	sanitizedNamespace := sanitizeForK8s(truncate(namespace, maxNamespaceLength))
	
	// Generate random 3-digit number
	randomNum, err := generateRandomDigits()
	if err != nil {
		return "", fmt.Errorf("failed to generate random digits: %w", err)
	}
	
	// Combine: name-namespace-XXX
	identifier := fmt.Sprintf("%s-%s-%s", sanitizedName, sanitizedNamespace, randomNum)
	
	// Final safety check - ensure it's within K8s limits
	if len(identifier) > 63 {
		// If somehow exceeds, truncate further
		identifier = identifier[:63]
		// Ensure it ends with alphanumeric
		if identifier[len(identifier)-1] == '-' {
			identifier = identifier[:len(identifier)-1] + "0"
		}
	}
	
	return identifier, nil
}
