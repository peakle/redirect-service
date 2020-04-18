GOOS=linux
GOARCH=amd64
HOSTNAME="very.grozny.ru"
NAME=rds

all: clean build

clean:
	@echo ">> cleaning..."
	@rm -f $(NAME)

build: clean
	@echo ">> building..."
	@ CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
	    go build -o $(NAME) ./cmd/main.go
	@chmod +x $(NAME)

#TODO add scp release to server
release: clean
	@echo "">> building..."
	@ CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
		go build -ldflags "-X main.Hostname=$(HOSTNAME)" -o $(NAME) ./cmd/main.go
	@chmod +x $(NAME)
