env GOOS=linux GOARCH=arm go build -o app cmd/app/main.go 
docker cp ./app resourcemanager:/