FROM golang:1.26.2
ENV GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local CGO_ENABLED=0
WORKDIR /workspace
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd
COPY internal ./internal
COPY web ./web
RUN go build -mod=vendor ./...
CMD ["go", "run", "-mod=vendor", "./cmd/cellforge"]
