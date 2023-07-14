# golangイメージとdistrolessイメージの間でdebianバージョンおよびCPUアーキテクチャを一致させること

FROM amd64/golang:1.20-bullseye as builder

WORKDIR /workspace
RUN wget https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.3.1.tar.gz \
    && tar xvzf libwebp-1.3.1.tar.gz \
    && cd libwebp-1.3.1 \
    && ./configure \
    && make \
    && make install

WORKDIR /go/src
COPY . .
RUN go mod download
RUN GOOS=linux GOARC=amd64 CGO_ENABLED=1 go build -ldflags '-s -w' -o /go/bin/mono --tags=prod

FROM gcr.io/distroless/base-debian11:latest-amd64
COPY --from=builder /go/bin/mono /
COPY --from=builder /usr/local/lib/ /usr/local/lib/
ENV GOOGLE_APPLICATION_CREDENTIALS=/etc/mono/gcpkey.json
ENV LD_LIBRARY_PATH=/usr/local/lib
CMD ["/mono", "-conf", "/etc/mono/config.json"]
