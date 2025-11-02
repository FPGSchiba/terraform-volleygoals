package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/router"
	"github.com/fpgschiba/volleygoals/utils"
)

func init() {
	err := utils.SetupTracing()
	if err != nil {
		log.Fatalf("failed to setup tracing: %v", err)
	}
	db.GetClient()
}

func main() {
	lambda.Start(router.HandleRequest)
}
