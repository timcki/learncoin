FROM golang:rc-alpine as builder
WORKDIR /src
ENV CGO_ENABLED=0

# Make sure to not redownload all dependencies each time
COPY go.mod .
COPY go.sum .
RUN go mod download
# Copy source files
COPY cmd cmd
COPY internal internal

RUN go build -o bin/learncoind cmd/learncoind/main.go && \
    go build -o bin/cryptotesting cmd/cryptotesting/main.go

FROM alpine
COPY --from=builder /src/bin/learncoind /bin/learncoind
COPY --from=builder /src/bin/cryptotesting /bin/cryptotesting
CMD ["/bin/cryptotesting"]
