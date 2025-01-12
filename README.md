# Lambdahandler

**lambdahandler** is a custom lightweight router for creating serverless APIs using AWS Lambda Function URLs, bypassing the need for API Gateway. This package simplifies routing and CORS handling while providing flexibility similar to traditional API frameworks.

## Features

- Simplified routing with support for path parameters.
- Built-in CORS configuration.
- Centralized error handling.
- JSON serialization for responses.
- Lightweight and easy to configure.

## Installation

Install the package using Go modules:

```bash
$ go get github.com/arthurmvo/lambdahandler
```

## Usage

### Basic Example

```go
package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/arthurmvo/lambdahandler"
	"github.com/aws/aws-lambda-go/events"
)

func main() {
	router := lambdahandler.NewRouter()

	// Configure CORS (Optional)
	router.Origins = []string{"https://example.com"} // Allow specific origins
	router.Methods = []string{"GET", "POST"}         // Allow specific methods
	router.Headers = []string{"Content-Type"}       // Allow specific headers

	// Define routes
	router.Get("/users", func(ctx context.Context, req events.LambdaFunctionURLRequest, params map[string]string) (interface{}, error) {
		return map[string]string{"message": "List of users"}, nil
	})

	router.Get("/users/:id", func(ctx context.Context, req events.LambdaFunctionURLRequest, params map[string]string) (interface{}, error) {
		id := params["id"]
		return map[string]string{"message": "User details", "id": id}, nil
	})

	router.Post("/users", func(ctx context.Context, req events.LambdaFunctionURLRequest, params map[string]string) (interface{}, error) {
		return map[string]string{"message": "User created"}, nil
	})

	lambda.Start(router.HandleRequest)
}
```

## CORS Configuration

You can customize CORS settings in your `Router` instance:

- **Origins**: A list of allowed origins or `*` for all origins.
- **Methods**: A list of allowed HTTP methods.
- **Headers**: A list of allowed headers.

If no values are set, the default is `*` for all three.

### Example

```go
router.Origins = []string{"https://example1.com", "https://example2.com"}
router.Methods = []string{"GET", "POST", "PUT"}
router.Headers = []string{"Content-Type", "Authorization"}
```

### Route Matching and Parameters

- Static paths like `/users` are matched exactly.
- Dynamic paths like `/users/:id` allow extracting parameters.

### Dynamic Route Example

```go
router.Get("/users/:id", func(ctx context.Context, req events.LambdaFunctionURLRequest, params map[string]string) (interface{}, error) {
	id := params["id"]
	return map[string]string{"message": "User ID received", "id": id}, nil
})
```

The `params` map will include all extracted parameters (e.g., `{ "id": "123" }`).

## Utility Functions

### JSON Responses

- Automatically serializes data structures to JSON.
- Adds `Content-Type: application/json` header.

### Error Handling

To return an error from a route handler:

```go
return nil, fmt.Errorf("Something went wrong")
```

This results in a `500` response with the error message in the body.

## Folder Structure

A typical usage scenario:

```
.
├── main.go            # Main application file
└── go.mod            # Go module file
```

## Contribution

Feel free to open issues or submit pull requests to enhance the library.

## License

This project is licensed under the [MIT License](LICENSE).

