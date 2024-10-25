package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

var emailIndex = "email-index"
var tableName = "service_providers"

func handleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	credential, err := getBasicCredential(event.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       err.Error(),
		}, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	dynamoDbClient := dynamodb.NewFromConfig(cfg)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	expr, err := expression.NewBuilder().WithKeyCondition(expression.Key("email").Equal(expression.Value(credential.Email))).Build()
	if err != nil {
		log.Printf("Couldn't build expression for query. Here's why: %v\n", err)
	}

	query, err := dynamoDbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:                 &tableName,
		IndexName:                 &emailIndex,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition()})

	if err != nil {
		log.Printf("Couldn't query for service provider %v. Here's why: %v\n", credential.Email, err.Error())
		return events.APIGatewayProxyResponse{}, err
	}

	for _, item := range query.Items {
		provider := ServiceProvider{}
		err := attributevalue.UnmarshalMap(item, &provider)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		err = bcrypt.CompareHashAndPassword([]byte(provider.Password), []byte(credential.Password))
		if err != nil {
			cookie := buildSessionCookie(provider)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"Set-Cookie": cookie.String(),
				},
			}, nil
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
	}, nil

}

func buildSessionCookie(provider ServiceProvider) *http.Cookie {
	return &http.Cookie{
		Name:     "session_cookie",
		Value:    base64.StdEncoding.EncodeToString([]byte(provider.Email)),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(8 * time.Hour),
	}
}

func getBasicCredential(body string) (BasicCredential, error) {
	credential := BasicCredential{}
	err := json.Unmarshal([]byte(body), &credential)
	if err != nil {
		fmt.Println("Couldn't convert request body due to: ", err.Error())
	}
	return credential, err
}

func main() {
	lambda.Start(handleRequest)
}

type BasicCredential struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ServiceProvider struct {
	Id       string `json:"id" dynamodbav:"id"`
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
}
