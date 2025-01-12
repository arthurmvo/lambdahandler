package lambdahandler

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Params represents the extracted URL parameters
type Params map[string]string

type LambdaError interface {
	Code() int       // Returns the HTTP status code
	Message() string // Returns the error message
}

// ErrorResponse is the struct that implements the LambdaError interface
type LambdaErrorResponse struct {
	CodeValue    int    `json:"code"`    // The error code (e.g., 500, 404)
	MessageValue string `json:"message"` // A human-readable error message
}

// Code returns the error code for ErrorResponse
func (e *LambdaErrorResponse) Code() int {
	return e.CodeValue
}

// Message returns the error message for ErrorResponse
func (e *LambdaErrorResponse) Message() string {
	return e.MessageValue
}

// NewLambdaError is a constructor for creating a new LambdaError with code and message
func NewLambdaError(code int, message string) LambdaError {
	return &LambdaErrorResponse{
		CodeValue:    code,
		MessageValue: message,
	}
}

// HandlerFunc defines the type for route handlers
type HandlerFunc func(ctx context.Context, req events.LambdaFunctionURLRequest, params Params) (interface{}, LambdaError)

// Route holds information about a single route
type Route struct {
	Method   string
	Pattern  *regexp.Regexp
	Handler  HandlerFunc
	Template string
}

// Router manages routes and CORS configuration
type Router struct {
	routes  []*Route
	Origins []string // Allowed origins, default is ["*"]
	Methods []string // Allowed methods, default is all methods
	Headers []string // Allowed headers, default is all headers
}

// NewRouter creates a new Router instance with default CORS settings
func NewRouter() *Router {
	return &Router{
		routes:  []*Route{},
		Origins: []string{"*"}, // Allow all origins by default
		Methods: []string{"*"}, // Allow all methods by default
		Headers: []string{"*"}, // Allow all headers by default
	}
}

// AddRoute adds a new route to the router
func (r *Router) AddRoute(method, path string, handler HandlerFunc) {
	pattern := buildPathPattern(path)
	r.routes = append(r.routes, &Route{
		Method:   method,
		Pattern:  pattern,
		Handler:  handler,
		Template: path,
	})
}

// Shortcut methods
func (r *Router) Get(path string, handler HandlerFunc)    { r.AddRoute("GET", path, handler) }
func (r *Router) Post(path string, handler HandlerFunc)   { r.AddRoute("POST", path, handler) }
func (r *Router) Put(path string, handler HandlerFunc)    { r.AddRoute("PUT", path, handler) }
func (r *Router) Delete(path string, handler HandlerFunc) { r.AddRoute("DELETE", path, handler) }

// HandleRequest is the main entry point for the Lambda function
func (r *Router) HandleRequest(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	var response events.LambdaFunctionURLResponse
	path, method := req.RequestContext.HTTP.Path, req.RequestContext.HTTP.Method

	// Handle CORS preflight requests
	if method == "OPTIONS" {
		return r.corsPreflightResponse(req), nil
	}

	// Match route
	for _, route := range r.routes {
		if route.Method == method && route.Pattern.MatchString(path) {
			params := extractParams(path, route.Pattern)
			data, err := route.Handler(ctx, req, params)
			if err != nil {
				return ErrorResponse(err), nil
			}
			response = SuccessResponse(data)
			break
		}
	}

	// Route not found
	if response.StatusCode == 0 {
		response = events.LambdaFunctionURLResponse{
			StatusCode: 404,
			Body:       "Route not found",
		}
	}

	// Attach CORS headers
	r.attachCORSHeaders(&response, req)

	return response, nil
}

// corsPreflightResponse handles preflight CORS requests
func (r *Router) corsPreflightResponse(req events.LambdaFunctionURLRequest) events.LambdaFunctionURLResponse {
	origin := req.Headers["origin"]
	if !r.isOriginAllowed(origin) {
		return events.LambdaFunctionURLResponse{
			StatusCode: 403,
			Body:       "Origin not allowed",
		}
	}
	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  origin,
			"Access-Control-Allow-Methods": strings.Join(r.Methods, ", "),
			"Access-Control-Allow-Headers": strings.Join(r.Headers, ", "),
		},
	}
}

// attachCORSHeaders adds CORS headers to responses
func (r *Router) attachCORSHeaders(response *events.LambdaFunctionURLResponse, req events.LambdaFunctionURLRequest) {
	origin := req.Headers["origin"]
	if origin == "" || !r.isOriginAllowed(origin) {
		origin = "*"
	}

	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["Access-Control-Allow-Origin"] = origin
	response.Headers["Access-Control-Allow-Methods"] = strings.Join(r.Methods, ", ")
	response.Headers["Access-Control-Allow-Headers"] = strings.Join(r.Headers, ", ")
}

// isOriginAllowed checks if the request origin is allowed
func (r *Router) isOriginAllowed(origin string) bool {
	if len(r.Origins) == 1 && r.Origins[0] == "*" {
		return true
	}
	for _, allowedOrigin := range r.Origins {
		if allowedOrigin == origin {
			return true
		}
	}
	return false
}

// Utility to generate error response
func ErrorResponse(err LambdaError) events.LambdaFunctionURLResponse {
	return events.LambdaFunctionURLResponse{
		StatusCode: err.Code(),
		Body:       fmt.Sprintf("Error: %s", err.Message()),
	}
}

// Utility to generate success response
func SuccessResponse(data interface{}) events.LambdaFunctionURLResponse {
	body, _ := json.Marshal(data) // Ignoring errors for simplicity
	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}
