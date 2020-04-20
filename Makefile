include .env

GOOS=linux
GOARCH=amd64

all: clean build

clean:
	@echo ">> cleaning..."
	@rm -f $(APP_NAME)

build: clean
	@echo ">> building..."
	@ CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
	    go build -o $(APP_NAME) ./cmd/main.go
	@chmod +x $(APP_NAME)

#TODO add scp release for
release: clean
	@echo ">> building..."
	@ CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
		go build -ldflags "-X main.Hostname=$(HOSTNAME)" -o $(APP_NAME) ./cmd/main.go
	@chmod +x $(APP_NAME)
	@echo ">> deploy..."
	@scp -P 22 ${APP_NAME} GeoIP2.mmdb index.html ${USERNAME}@${HOSTNAME}:${APP_DIR}
	@ssh -f ${USERNAME}@${HOSTNAME} '${APP_DIR}/${APP_NAME}'

