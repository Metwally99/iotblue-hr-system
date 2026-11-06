FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/hr-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bootstrap-admin ./cmd/bootstrap-admin

FROM alpine:3.22
RUN apk add --no-cache tzdata && addgroup -S app && adduser -S app -G app
WORKDIR /app
USER app
COPY --from=build /out/hr-api /usr/local/bin/hr-api
COPY --from=build /out/bootstrap-admin /usr/local/bin/bootstrap-admin
COPY api/openapi.yaml /app/api/openapi.yaml
EXPOSE 8080
ENTRYPOINT ["hr-api"]
