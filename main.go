package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mblydenburgh/go-rest/domain"
	"mblydenburgh/go-rest/repository"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/golang-jwt/jwt"
	"github.com/guregu/dynamo"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

func main() {
	lambda.Start(router)
}

func router(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	switch request.RequestContext.HTTP.Method {
	case http.MethodGet:
		return getHandler(request)
	case http.MethodPost:
		return postHandler(request)
	case http.MethodDelete:
		return deleteHandler(request)
	default:
		return clientError(http.StatusMethodNotAllowed)
	}
}

func getHandler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Println("Handling request")
	dynamoClient := dynamo.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

	isValidJwt := validateJWT(request.Headers["Authorization"])
	if !isValidJwt {
		return clientError(403)
	}

	carId := request.QueryStringParameters["id"]
	car, err := repository.GetCar(dynamoClient, carId)
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

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       string(carJson),
	}, nil
}

func postHandler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	dynamoClient := dynamo.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})
	if request.Headers["content-type"] != "application/json" && request.Headers["Content-Type"] != "application/json" {
		return clientError(http.StatusNotAcceptable)
	}

	fmt.Println("Handling request headers: ", request.Headers)
	isValidJwt := validateJWT(request.Headers["Authorization"])
	if !isValidJwt {
		return clientError(403)
	}

	car := new(domain.SaveCarPayload)
	err := json.Unmarshal([]byte(request.Body), car)
	if err != nil {
		return clientError(http.StatusUnprocessableEntity)
	}

	if car.Manufacturer == "" || car.Model == "" || car.Year <= 1900 {
		return clientError(http.StatusNotAcceptable)
	}

	id, err := repository.PutCar(dynamoClient, car)
	if err != nil {
		return serverError(err)
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 201,
		Headers:    map[string]string{"Location": fmt.Sprintf("/cars?id=%s", id)},
	}, nil
}

func deleteHandler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Println("Handling delete request")
	dynamoClient := dynamo.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})
	id := request.QueryStringParameters["id"]

	err := repository.DeleteCar(dynamoClient, id)
	if err != nil {
		log.Printf("Error deleting car due to %v", err)
		return serverError(err)
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
	}, nil

}

func validateJWT(authHeader string) bool {
	var hmacSampleSecret []byte
	log.Println("Validating JWT with authHeader: ", authHeader)
	tokenStr := strings.Split(authHeader, " ")[1]
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return hmacSampleSecret, nil
	})
	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return false
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["signer"] == "go-auth" {
			log.Println("Valid jwt claim")
			return true
		} else {
			log.Println("Invalid jwt claim")
			return false
		}
	} else {
		log.Println("Error validating JWT")
		fmt.Println(err)
		return false
	}
}

func serverError(err error) (events.APIGatewayV2HTTPResponse, error) {
	fmt.Println(err.Error())

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func clientError(status int) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}
