###
FROM golang:1.13-alpine as builder

RUN apk update
RUN apk add git

RUN mkdir $GOPATH/src/SLALite
COPY . $GOPATH/src/SLALite
WORKDIR $GOPATH/src/SLALite

ENV GO111MODULE=on

ARG VERSION
ARG DATE 

RUN CGO_ENABLED=0 GOOS=linux go build -a -o SLALite -ldflags="-X main.version=${VERSION} -X main.date=${DATE}" .

###
FROM alpine:latest
WORKDIR /opt/slalite
COPY --from=builder /go/src/SLALite/SLALite .

RUN mkdir /etc/slalite
RUN addgroup -S slalite && adduser -D -G slalite slalite
RUN chown -R slalite:slalite /etc/slalite && chmod 700 /etc/slalite

EXPOSE 8090
#ENTRYPOINT ["./run_slalite.sh"]
USER slalite
ENTRYPOINT ["/opt/slalite/SLALite", "-f", "/etc/ditas/slalite.yml"]

