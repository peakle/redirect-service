GOOS=linux
GOARCH=amd64
HOSTNAME="very.grozny.ru"
NAME=redirect-service

all: clean build

clean:
	@echo ">> cleaning..."
	@rm -f $(NAME)

build: clean
	@echo ">> building..."
	@ CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
	    go build -ldflags "-X main.Hostname=$(HOSTNAME)" -o $(NAME) ./cmd/main.go
	@chmod +x $(NAME)
