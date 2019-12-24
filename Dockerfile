FROM golang:1.13 as build
ARG RELEASE=TRUE
WORKDIR /var/myrunes
ADD . .
RUN go mod tidy
RUN mkdir ./bin
RUN go build \
        -v -o /app/myrunes -ldflags "\
            -X github.com/myrunes/backend/internal/static.Release=${RELEASE} \
            -X github.com/myrunes/backend/internal/static.AppVersion=$(git describe --tags)" \
        ./cmd/server/*.go

FROM debian:stretch-slim AS final
LABEL maintainer="zekro <contact@zekro.de>"
RUN apt-get update &&\
    apt-get install -y ca-certificates &&\
    update-ca-certificates
WORKDIR /app
COPY --from=build /app .

EXPOSE 8080
RUN mkdir -p /etc/myrunes
CMD ["/app/myrunes", "-c", "/etc/myrunes/config.yml"]
