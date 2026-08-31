ARG GOOS=darwin
ARG GOARCH=arm64

FROM golang:1.26 AS builder
ARG GOOS
ARG GOARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="-s -w" -o /out/printmark ./cmd/printmark

FROM scratch
COPY --from=builder /out/printmark /printmark
