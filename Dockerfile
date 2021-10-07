FROM golang:alpine as builder
WORKDIR /learncoin
COPY ./ .
RUN go get && CGO_ENABLED=0 go build -o learncoind *.go

FROM scratch
COPY --from=builder /learncoin/learncoind /learncoind
CMD ["/learncoind", "-i", "-d"]