FROM golang:latest as builder
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
    go build -o bin/chainsim cmd/chain_simulation/main.go

FROM alpine
COPY --from=builder /src/bin/learncoind /bin/learncoind
COPY --from=builder /src/bin/chainsim /bin/chainsim
CMD ["/bin/chainsim"]
