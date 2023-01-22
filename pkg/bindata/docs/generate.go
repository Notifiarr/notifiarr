package docs

//nolint:lll
//go:generate go run github.com/swaggo/swag/cmd/swag@master i --parseDependency --instanceName api --outputTypes go  --parseInternal --dir ../../../ -g main.go --output .
