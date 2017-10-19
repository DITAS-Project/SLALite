FROM golang:1.8.3 as builder
WORKDIR /go/src/SLALite

RUN go get -d -v github.com/gorilla/mux
RUN go get -d -v github.com/coreos/bbolt
RUN go get -d -v github.com/spf13/viper
RUN go get -d -v gopkg.in/mgo.v2

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o SLALite .

FROM mvertes/alpine-mongo:latest
WORKDIR /root/
COPY --from=builder /go/src/SLALite/SLALite .
COPY docker/slalite_mongo.yml /etc/slalite/slalite.yml
COPY docker/run_slalite_mongo.sh run_slalite.sh
RUN chmod 0777 ./run_slalite.sh
CMD ["./run_slalite.sh"]