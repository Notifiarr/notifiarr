package docs

//go:generate go run github.com/swaggo/swag/cmd/swag@v1.8.12 i --parseDependency --instanceName api --outputTypes go  --parseInternal --dir ../../../ -g main.go --output .
