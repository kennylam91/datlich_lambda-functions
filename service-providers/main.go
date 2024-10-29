package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"golang.org/x/crypto/bcrypt"
)

var tableName = "service_providers"

func handleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	dynamoDbClient := dynamodb.NewFromConfig(cfg)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, err
	}

	switch event.Resource {
	case "/service-providers/register":
		switch event.HTTPMethod {
		case http.MethodPost:
			return handleProviderCreationRequest(ctx, event, dynamoDbClient)
		}

	case "/service-providers/auth":
		switch event.HTTPMethod {
		case http.MethodPost:
			return handleProviderAuthRequest(ctx, event, dynamoDbClient)

		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusMethodNotAllowed,
	}, nil
}

func handleProviderAuthRequest(ctx context.Context, event events.APIGatewayProxyRequest, client *dynamodb.Client) (events.APIGatewayProxyResponse, error) {
	credential, err := getBasicCredential(event.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       err.Error(),
		}, nil
	}

	username, err := attributevalue.Marshal(credential.Username)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, err
	}

	item, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tableName,
		Key:       map[string]types.AttributeValue{"username": username},
	})

	foundProvider := ServiceProvider{}
	err = attributevalue.UnmarshalMap(item.Item, &foundProvider)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundProvider.Password), []byte(credential.Password))
	if err == nil {
		cookie := buildSessionCookie(foundProvider)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Set-Cookie": cookie.String(),
			},
		}, nil
	} else {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

}

func handleProviderCreationRequest(ctx context.Context, event events.APIGatewayProxyRequest, dynamoDbClient *dynamodb.Client) (events.APIGatewayProxyResponse, error) {
	item, err := buildPutItem(event.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Error processing request: %v", err),
		}, err
	}
	_, err = dynamoDbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      item,
	})
	if err != nil {
		log.Printf("Couldn't add item to table %v.\n%v", tableName, err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       err.Error(),
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
	}, nil
}

func buildPutItem(eventBody string) (map[string]types.AttributeValue, error) {
	provider := ServiceProvider{}
	err := json.Unmarshal([]byte(eventBody), &provider)
	if err != nil {
		fmt.Printf("Failed to parse request body %v", eventBody)
		return nil, err
	}
	provider.CreatedAt = time.Now()
	rawPassword := provider.Password
	provider.Password, err = hashPassword(rawPassword)
	if err != nil {
		fmt.Printf(`Failed to hash password "%v"`, rawPassword)
		return nil, err
	}
	return attributevalue.MarshalMap(provider)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func main() {
	lambda.Start(handleRequest)
}

type ServiceProvider struct {
	Username  string    `json:"username" dynamodbav:"username"`
	Name      string    `json:"name" dynamodbav:"name"`
	Email     string    `json:"email" dynamodbav:"email"`
	Phone     string    `json:"phone" dynamodbav:"phone"`
	Password  string    `json:"password" dynamodbav:"password"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
}
