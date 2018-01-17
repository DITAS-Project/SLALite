###
FROM golang:1.8.3 as builder
WORKDIR /go/src/SLALite

RUN go get -d -v github.com/gorilla/mux
RUN go get -d -v github.com/coreos/bbolt
RUN go get -d -v github.com/spf13/viper
RUN go get -d -v gopkg.in/mgo.v2
RUN go get -d -v github.com/labstack/gommon/log
RUN go get -d -v github.com/oleksandr/conditions

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o SLALite .

###
FROM mvertes/alpine-mongo:latest
WORKDIR /opt/slalite
COPY --from=builder /go/src/SLALite/SLALite .
COPY docker/slalite_mongo.yml /etc/slalite/slalite.yml
COPY docker/run_slalite_mongo.sh run_slalite.sh

RUN addgroup -S slalite && adduser -D -G slalite slalite
RUN chown -R slalite:slalite /etc/slalite && chmod 700 /etc/slalite

EXPOSE 8090
ENTRYPOINT ["./run_slalite.sh"]
#USER slalite
#ENTRYPOINT ["/opt/slalite/SLALite"]

