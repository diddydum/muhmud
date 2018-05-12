FROM golang:1.10 as builder
WORKDIR /go/src/github.com/diddydum/muhmud/muhmud
COPY . .
# Think I need the CGO stuff because of alpine
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o muhmud .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/diddydum/muhmud/muhmud/muhmud .
COPY --from=builder /go/src/github.com/diddydum/muhmud/muhmud/muhmud.conf.yaml .
CMD ["./muhmud"]
