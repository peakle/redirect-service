include .env

GOOS=linux
GOARCH=amd64

clean:
	@echo ">> cleaning..."
	@rm -f ${APP_NAME}

build: clean
	@echo ">> building..."
	@ CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
	    go build -o ${APP_NAME} ./cmd/main.go
	@chmod +x ${APP_NAME}

release: clean
	@echo ">> building..."
	@ CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
		go build -ldflags \
		"-X main.Hostname=https://${HOSTNAME} \
		-X main.WriteUser=${MYSQL_WRITE_USER}:${MYSQL_WRITE_PASS} \
		-X main.ReadUser=${MYSQL_READ_USER}:${MYSQL_READ_PASS} \
		-X main.ProjectDir=${APP_DIR}" \
		-o ${APP_NAME} ./cmd/main.go
	@chmod +x ${APP_NAME}
	@echo ">> deploy..."
	@rsync -ve ssh --progress ${APP_NAME} GeoIP2.mmdb index.html favicon.ico ${USERNAME}@${HOSTNAME}:${APP_DIR}
	@ssh ${USERNAME}@${HOSTNAME} 'export MYSQL_HOST=${MYSQL_HOST} && export MYSQL_DATABASE=${MYSQL_DATABASE} \
	&& supervisorctl restart rds-server:'
