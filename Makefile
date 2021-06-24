run-dev:
	docker-compose up -d
	sleep 10
	HTTP_PORT=8080 DB_DSN="root:password@tcp(127.0.0.1:3306)/db" DATAGOUV_URL="https://www.data.gouv.fr/fr/datasets/r/406c6a23-e283-4300-9484-54e78c8ae675" go run ./cmd/api

.PHONY: run-dev
