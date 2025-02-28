FROM golang:1.23.5-bookworm

RUN apt update && apt install -y postgresql

RUN groupadd -g 1001 appuser && \
    useradd -u 1001 -g 1001 -m appuser
USER appuser

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY registration.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./registrationapi registration.go
RUN chmod +x ./registrationapi

CMD ["./registrationapi"]
