run:
	go run cmd/main.go

swag-generate:
	cd cmd && swag init -g ../cmd/main.go -d ../config,../models,../controllers,../database,../repository -o ../docs

.PHONY: run swag-generate