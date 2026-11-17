FROM golang:1.27.4@sha256:205c7636d74e0ffbe3d007162df7d74f859f5a8c8322dfb9f907977686c126cf as build-base
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
