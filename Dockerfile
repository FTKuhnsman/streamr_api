FROM golang:latest as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
ENV GOCACHE=/root/.cache/go-build
# ADD . /go/src/myapp
# WORKDIR /go/src/myapp
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o streamr_api main.go
#RUN go install

FROM ubuntu:22.04
ENV SSL_CERT_DIR=/etc/ssl/certs
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/streamr_api .
RUN chmod +x streamr_api
ENTRYPOINT ["./streamr_api"]