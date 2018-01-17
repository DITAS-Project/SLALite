# SLALite #

##

The SLALite is an implementation of an SLA system.

## Quick usage guide ##

### Installation ###

Build the Docker image:

    docker build -t slalite .

Run the container:

    docker run -ti -p 8090:8090 slalite

Stop execution pressing CTRL-C

To run the service under HTTPs, you must change supply a different configuration file and the certificate files. You will find these files in docker/https for debugging purposes. DO NOT USE THE CERT.PEM and KEY.PEM in production!!

    docker run -ti -p 8090:8090 -v $PWD/docker/https:/etc/slalite slalite

### Usage ###

SLALite offers a usual REST API, with an endpoint on /agreements

Add an agreement:

    curl -k -X POST -d @agreement.json https://localhost:8090/agreements

Get agreements:

    curl -k https://localhost:8090/agreements
    curl -k https://localhost:8090/agreements/a02

