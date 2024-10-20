package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"net/http"
	"strconv"
	"time"
)

//var config = aws.Config{
//	Region: "ap-southeast-1",
//}

var tableName = "service_providers"

func handleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Body size = %d.\n", len(event.Body))
	cfg, err := config.LoadDefaultConfig(ctx)
	dynamoDbClient := dynamodb.NewFromConfig(cfg)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, err
	}
	switch event.HTTPMethod {
	case "POST":
		item, err := buildPutItem(event.Body)
		_, err = dynamoDbClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &tableName,
			Item:      item,
		})
		if err != nil {
			log.Printf("Couldn't add item to table %v.\n%v", tableName, err)
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body:       err.Error(),
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 201,
		}, nil

	}
	fmt.Println("Hello lambda function from Go")
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusMethodNotAllowed,
	}, nil
}

func buildPutItem(eventBody string) (map[string]types.AttributeValue, error) {
	provider := ServiceProvider{}
	err := json.Unmarshal([]byte(eventBody), &provider)
	if err != nil {
		fmt.Println(err.Error())
	}
	now := time.Now()
	provider.CreatedAt = now
	provider.Id = strconv.FormatInt(now.Unix(), 10)
	fmt.Printf("provider: %v\n", provider)
	item, err := attributevalue.MarshalMap(provider)
	//item["id"] = &types.AttributeValueMemberS{
	//	Value: strconv.FormatInt(now.Unix(), 10),
	//}
	if err != nil {
		panic(err)
	}
	fmt.Printf("item: %v\n", item)

	return item, err
}

func main() {
	lambda.Start(handleRequest)
}

type ServiceProvider struct {
	Id        string    `json:"id" dynamodbav:"id"`
	Name      string    `json:"name" dynamodbav:"name"`
	Email     string    `json:"email" dynamodbav:"email"`
	Phone     string    `json:"phone" dynamodbav:"phone"`
	Password  string    `json:"password" dynamodbav:"password"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
}
