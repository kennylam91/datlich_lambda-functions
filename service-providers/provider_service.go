package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func getProviderByUsername(ctx context.Context, username string, client *dynamodb.Client) (*dynamodb.GetItemOutput, error) {
	usernameAttrValue, err := attributevalue.Marshal(username)
	if err != nil {
		return nil, err
	}

	item, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tableName,
		Key:       map[string]types.AttributeValue{"username": usernameAttrValue},
	})
	return item, nil
}
