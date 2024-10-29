package main

import "time"

type ServiceProvider struct {
	Username  string    `json:"username" dynamodbav:"username" validate:"required"`
	Name      string    `json:"name" dynamodbav:"name" validate:"required"`
	Email     string    `json:"email" dynamodbav:"email" validate:"required,email"`
	Phone     string    `json:"phone" dynamodbav:"phone"`
	Password  string    `json:"password" dynamodbav:"password" validate:"required"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
}
