###
FROM golang:alpine as builder

ARG VERSION
ARG DATE 

RUN apk add --no-cache git

WORKDIR /go/src/SLALite

COPY . .
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o SLALite -ldflags="-X main.version=${VERSION} -X main.date=${DATE}" .

###
FROM mvertes/alpine-mongo:latest
WORKDIR /opt/slalite
COPY --from=builder /go/src/SLALite/SLALite .
COPY docker/slalite_mongo.yml /etc/slalite/slalite.yml
COPY docker/run_slalite_mongo.sh run_slalite.sh

RUN addgroup -S slalite && adduser -D -G slalite slalite
RUN chown -R slalite:slalite /etc/slalite && chmod 700 /etc/slalite

EXPOSE 8090
#ENTRYPOINT ["./run_slalite.sh"]
USER slalite
ENTRYPOINT ["/opt/slalite/SLALite"]

