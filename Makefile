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
		"-X main.Hostname=${HOSTNAME} -X main.WriteUser=${MYSQL_WRITE_USER}:${MYSQL_WRITE_PASS} -X main.ReadUser=${MYSQL_READ_USER}:${MYSQL_READ_PASS}" \
		-o ${APP_NAME} ./cmd/main.go
	@chmod +x ${APP_NAME}
	@echo ">> deploy..."
	@scp -P 22 ${APP_NAME} GeoIP2.mmdb index.html ${USERNAME}@${HOSTNAME}:${APP_DIR}
	@ssh ${USERNAME}@${HOSTNAME} 'rm -f ${APP_DIR}/.env; \
		echo MYSQL_HOST=${MYSQL_HOST} > ${APP_DIR}/.env; \
		echo MYSQL_DATABASE=${MYSQL_DATABASE} >> ${APP_DIR}/.env; \
		source ${APP_DIR}/.env'
	@ssh ${USERNAME}@${HOSTNAME} 'supervisorctl restart'
