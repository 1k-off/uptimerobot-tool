FROM golang:1.15.6-alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build

FROM alpine:latest
LABEL maintainer="Bogdan Kosarevskyi <bogdan.kosarevskyi@gmail.com>"
LABEL vendor="1node"
LABEL lastUpdate="26-12-2020"
LABEL description="Sonarqube stats bitbucket application."
COPY --from=build /app/uptimerobot-tool /app/uptimerobot-tool
COPY --from=build /app/data /app/data
WORKDIR /app
VOLUME /app/data
CMD ["/app/uptimerobot-tool"]