FROM golang:1.20 as builder

WORKDIR /app
COPY . .
RUN go get && CGO_ENABLED=0 go build --tags=prod


FROM gcr.io/distroless/static-debian11:latest

WORKDIR /var/lib/mono

COPY --from=builder /app/mono .

ENV GOOGLE_APPLICATION_CREDENTIALS=/etc/mono/gcpkey.json

CMD ["./mono", "-conf", "/etc/mono/config.json"]
