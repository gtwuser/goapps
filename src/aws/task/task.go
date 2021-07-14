package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"os"
)

// MyEvent is a thing
type MyEvent struct {
	Name string `json:"name"`
}

// HandleRequest for an event
func HandleRequest(name MyEvent) (string, error) {
	host := os.Getenv("HOST")
	return fmt.Sprintf("hi %s from %s", name.Name, host), nil
}

func main() {
	lambda.Start(HandleRequest)
}