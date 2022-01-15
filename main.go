package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mblydenburgh/go-rest/domain"
	"mblydenburgh/go-rest/repository"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

func main() {
	lambda.Start(router)
}

func router(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	switch request.RequestContext.HTTP.Method {
	case "GET":
		return getHandler(request)
	case "POST":
		return postHandler(request)
	default:
		return clientError(http.StatusMethodNotAllowed)
	}
}

func getHandler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Handling request")
	dynamoClient := dynamo.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

	carId := request.QueryStringParameters["id"]
	car, err := repository.GetItem(dynamoClient, carId)
	if err != nil {
		return serverError(err)
	}
	if car == nil {
		return clientError(http.StatusNotFound)
	}

	carJson, err := json.Marshal(car)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(carJson),
	}, nil
}

func postHandler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	dynamoClient := dynamo.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})
	if request.Headers["content-type"] != "application/json" && request.Headers["Content-Type"] != "application/json" {
		return clientError(http.StatusNotAcceptable)
	}

	car := new(domain.SaveCarPayload)
	err := json.Unmarshal([]byte(request.Body), car)
	if err != nil {
		return clientError(http.StatusUnprocessableEntity)
	}

	if car.Manufacturer == "" || car.Model == "" || car.Year <= 1900 {
		return clientError(http.StatusNotAcceptable)
	}

	id, err := repository.PutItem(dynamoClient, car)
	if err != nil {
		return serverError(err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers:    map[string]string{"Location": fmt.Sprintf("/cars?id=%s", id)},
	}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	fmt.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}
