package main

import (
	"example.com/m/common"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(common.HandleCheckStravaActivityUpdate)
}
