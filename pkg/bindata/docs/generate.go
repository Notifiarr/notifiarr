package docs

//go:generate go run github.com/swaggo/swag/cmd/swag@latest i --parseDependency --instanceName api --outputTypes go  --parseInternal --dir ../../../ -g main.go --output .
