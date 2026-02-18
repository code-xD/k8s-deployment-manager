package dto

import "errors"

// HTTP Header constants
const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// UserIDHeader is the header name for user ID
	UserIDHeader = "X-User-ID"
)

// API path constants
const (
	PathPing                    = "/api/v1/ping"
	PathDeploymentRequestsList  = "/api/v1/deployments/requests"
	PathDeploymentRequestByID  = "/api/v1/deployment/requests/:id"
	PathDeploymentsCreate       = "/api/v1/deployments/create"
)

// API response message constants (user-facing)
const (
	MessagePong = "pong"

	MsgDeploymentRequestsRetrieved = "Deployment requests retrieved successfully"
	MsgDeploymentRequestRetrieved   = "Deployment request retrieved successfully"
	MsgDeploymentRequestCreated     = "Deployment request created successfully"

	ErrMsgUserIDNotFound              = "User ID not found"
	ErrMsgRequestIDNotFound           = "Request ID not found"
	ErrMsgRequestIDRequired           = "Request ID is required"
	ErrMsgFailedToListDeploymentRequests = "Failed to list deployment requests"
	ErrMsgDeploymentRequestNotFound   = "Deployment request not found"
	ErrMsgFailedToGetDeploymentRequest = "Failed to get deployment request"
	ErrMsgDeploymentAlreadyExists      = "Deployment already exists"
	ErrMsgFailedToCreateDeploymentRequest = "Failed to create deployment request"
	ErrMsgRequestIDHeaderRequired     = "X-Request-ID header is required"
	ErrMsgFailedToCheckDeploymentRequest = "Failed to check existing deployment request"
	ErrMsgDeploymentRequestSameRequestIDExists = "Deployment request with same request ID already exists"
	ErrMsgUserIDHeaderRequired = "X-User-ID header is required"
	ErrMsgUserNotFoundResponse = "User not found"
	ErrMsgFailedToCreateUser  = "Failed to create user"
)

// API response body keys
const (
	ResponseKeyMessage = "message"
	ResponseKeyError   = "error"
	ResponseKeyDetails = "details"
	ResponseKeyParam   = "param"
)

// Path param names
const (
	ParamID = "id"
)

// Context key constants
const (
	// RequestIDKey is the context key for storing request ID
	RequestIDKey = "request_id"
	// UserIDKey is the context key for storing user ID
	UserIDKey = "user_id"
	// UserKey is the context key for storing user object
	UserKey = "user"
)

// K8s / template constants
const (
	TemplateNginx       = "nginx"
	ConfigMapIndexHTML  = "index.html"
	ConfigMapHTMLSuffix = "-html"
	MapstructureTagJSON = "json"
	// LabelKeyManagedBy is the label key for filtering deployments by manager (value from config manager_tag).
	LabelKeyManagedBy = "managed-by"
)

// Conflict detection: substring used to detect "already exists" errors
const StrAlreadyExists = "already exists"

// K8s config validation
const ErrMsgK8sManagerTagRequired = "k8s config manager-tag is required"

// Worker queue group default
const QueueGroupDeploymentWorkers = "deployment-workers"

// Error variables
var (
	// ErrUserNotFound is returned when user is not found in database
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidUserID is returned when user ID header is invalid
	ErrInvalidUserID = errors.New("invalid user ID")
	// ErrDeploymentRequestNotFound is returned when deployment request is not found or not owned by user
	ErrDeploymentRequestNotFound = errors.New("deployment request not found")
	// ErrRequestIDNotFoundInContext is returned when request ID is missing from context
	ErrRequestIDNotFoundInContext = errors.New("request ID not found in context")
	// ErrInvalidRequestIDTypeInContext is returned when request ID in context has wrong type
	ErrInvalidRequestIDTypeInContext = errors.New("invalid request ID type in context")
)
