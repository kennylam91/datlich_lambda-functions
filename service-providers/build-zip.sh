GOOS=linux GOARCH=amd64 go build -o bootstrap .
zip lambda-handler.zip bootstrap
