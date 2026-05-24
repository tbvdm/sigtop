FROM golang:1.23 as builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libsecret-1-dev pkg-config && rm -rf /var/lib/apt/lists/*

WORKDIR /src
RUN go install github.com/tbvdm/sigtop@master

FROM scratch AS export
COPY --from=builder /go/bin/sigtop /

