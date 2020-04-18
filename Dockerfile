FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ARG MYSQL_PASSWORD
ENV MYSQL_PASSWORD=$MYSQL_PASSWORD

ARG MYSQL_USER
ENV MYSQL_USER=$MYSQL_USER

ENV PORT 80
EXPOSE $PORT

RUN mkdir /app

COPY ./rds /app

WORKDIR /app

RUN chmod +x /app/rds
