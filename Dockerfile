FROM golang:alpine as builder
WORKDIR /src
ENV CGO_ENABLED=0

# Make sure to not redownload all dependencies each time
COPY go.mod .
COPY go.sum .
RUN go mod download
# Copy source files
COPY cmd cmd
COPY internal internal

RUN go build -o bin/learncoind cmd/learncoind.go

FROM alpine
COPY --from=builder /src/bin/learncoind /bin/learncoind
CMD ["/bin/learncoind"]
