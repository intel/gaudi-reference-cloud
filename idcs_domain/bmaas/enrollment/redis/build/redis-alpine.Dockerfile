FROM redis:7.4.0-alpine3.20

RUN apk update && apk upgrade && apk add bash
