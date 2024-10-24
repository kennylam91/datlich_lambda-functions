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
)

var tableName = "service_providers"

func handleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	credential := BasicCredential{}
	err := json.Unmarshal([]byte(event.Body), &credential)
	if err != nil {
		fmt.Println(err.Error())
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	dynamoDbClient := dynamodb.NewFromConfig(cfg)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, err
	}

	expr, err := expression.NewBuilder().WithKeyCondition(expression.Key("email").Equal(expression.Value(credential.Email))).Build()
	if err != nil {
		log.Printf("Couldn't build expression for query. Here's why: %v\n", err)
	}
	query, err := dynamoDbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:                 &tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition()})
	if err != nil {
		log.Printf("Couldn't query for service provider %v", credential.Email)
	}

	for _, item := range query.Items {
		provider := ServiceProvider{}
		err := attributevalue.UnmarshalMap(item, &provider)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
			}, err
		}

		err = bcrypt.CompareHashAndPassword([]byte(provider.Password), []byte(credential.Password))
		if err != nil {
			cookie := &http.Cookie{
				Name:     "session_cookie",
				Value:    base64.StdEncoding.EncodeToString([]byte(provider.Email)),
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"Set-Cookie": cookie.String(),
				},
			}, nil
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil

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
