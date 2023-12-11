FROM golang:1.21.4 as build-base
COPY go.mod go.sum /go/src/app/
WORKDIR /go/src/app
RUN go mod download
COPY . /go/src/app

FROM build-base AS build-api
RUN CGO_ENABLED=0 go build -o /go/bin/stratometrics-api cmd/api/main.go

FROM build-base AS build-listener
RUN CGO_ENABLED=0 go build -o /go/bin/stratometrics-listener cmd/listener/main.go

FROM gcr.io/distroless/static-debian12 AS api
COPY --from=build-api /go/bin/stratometrics-api /
CMD ["/stratometrics-api"]

FROM gcr.io/distroless/static-debian12 AS listener
COPY --from=build-listener /go/bin/stratometrics-listener /
CMD ["/stratometrics-listener"]
