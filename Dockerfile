###
FROM golang:alpine as builder

ARG VERSION
ARG DATE 

RUN apk add --no-cache git curl

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/SLALite

COPY . .
RUN rm -rf vendor
RUN dep ensure

RUN CGO_ENABLED=0 GOOS=linux go build -a -o SLALite -ldflags="-X main.version=${VERSION} -X main.date=${DATE}" .

###
FROM golang:alpine
WORKDIR /opt/slalite
COPY --from=builder /go/src/SLALite/SLALite .

RUN mkdir /etc/slalite
RUN addgroup -S slalite && adduser -D -G slalite slalite
RUN chown -R slalite:slalite /etc/slalite && chmod 700 /etc/slalite

EXPOSE 8090
#ENTRYPOINT ["./run_slalite.sh"]
USER slalite
ENTRYPOINT ["/opt/slalite/SLALite"]

