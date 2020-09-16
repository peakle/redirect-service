include .env

GOOS=linux
GOARCH=amd64

clean:
	@echo ">> cleaning..."
	@rm -f ${APP_NAME}

build: clean
	@echo ">> building..."
	@go generate ./cmd/main.go
	@ CGO_ENABLED=0 go build -o ${APP_NAME} ./cmd/*.go
	@chmod +x ${APP_NAME}

release: clean
	@echo ">> building..."
	@go generate ./cmd/main.go
	@ CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${APP_NAME} ./cmd/*.go
	@chmod +x ${APP_NAME}
	@echo ">> deploy..."
	@rsync -ve ssh --progress ${APP_NAME} GeoIP2.mmdb index.html favicon.ico ${USERNAME}@${HOSTNAME}:${APP_DIR}
	@ssh ${USERNAME}@${HOSTNAME} 'supervisorctl restart rds-server:'
