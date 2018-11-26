# SLALite #

## Description ##

The SLALite is a lightweight implementation of an SLA system, inspired by the
WS-Agreement standard. Its features are:

* REST interface to manage creation and update of agreements
* Agreements evaluation on background; any breach in the agreement terms
  generates an SLA violation.
* Configurable monitoring: a monitoring has to be provided externally.
* Configurable repository: a memory repository (for developing purposes)
  and a mongodb repository are provided, but more can be added.

An agreement is represented by a simple JSON structure 
(see examples in resources/samples):

```
{
    "id": "2018-000234",
    "name": "an-agreement-name",
    "details":{
        "id": "2018-000234",
        "type": "agreement",
        "name": "an-agreement-name",
        "provider": { "id": "a-provider", "name": "A provider" },
        "client": { "id": "a-client", "name": "A client" },
        "creation": "2018-01-16T17:09:45Z",
        "expiration": "2019-01-17T17:09:45Z",
        "guarantees": [
            {
                "name": "TestGuarantee",
                "constraint": "[execution_time] < 100"
            }
        ]
    }
}
```

## Quick usage guide ##

### Installation ###

Build the Docker image:

    make docker

Run the container:

    docker run -ti -p 8090:8090 slalite:<version>

Stop execution pressing CTRL-C

To run the service under HTTPs, you must change supply a different configuration file and the certificate files. You will find these files in docker/https for debugging purposes. DO NOT USE THE CERT.PEM and KEY.PEM in production!!

    docker run -ti -p 8090:8090 -v $PWD/docker/https:/etc/slalite slalite

### Configuration ###

The SLALite can be configured with a configuration file and with environment 
variables. The configuration file is read by default from /etc/slalite and the current 
working directory. The `-f` parameter can be used to set the config file location.

```
$ ./SLALite -h
Usage of SLALite:
  -b string
        Filename (w/o extension) of config file (default "slalite")
  -d string
        Directories where to search config files (default "/etc/slalite:.")
  -f string
        Path of configuration file. Overrides -b and -d
```

#### File settings ####

*General settings*

* `singlefile` (default: `false`). Sets if all file settings are read 
  from a single file or from several files. For example, when `singlefile=false`,
  the MongoDB settings are read from the file `mongodb.yml`.
* `repository` (default: `memory`). Sets the repository type to use. Set this
  value to `mongodb` to use a MongoDB database.
* `checkPeriod` (default: `60`). Sets the period in seconds of assessments 
  executions.
* `CAPath`. Sets the value of a file path containing certificates of trusted
  CAs; to be used to connect as client to SSL servers whose certificate is
  not trusted by default (e.g. self-signed certificates)

*REST interface settings*

* `port` (default: `8090`). Port of REST interface.
* `enableSsl` (default: `false`). Enables the use of SSL on the REST
  interface. The two following variables should be set.
* `sslCertPath` (default: `cert.pem`). Sets the certificate path.
* `sslKeyPath` (default: `key.pem`). Sets the private key path to access the
  certificate.

*MongoDB settings (default file: /etc/slalite/mongodb.yml)*

* `connection` (default: `localhost`). Sets the MongoDB host.
* `database` (default: `slalite`). Sets the MongoDB database name to use.
* `clear_on_boot` (default: `false`). Sets if the database is cleared on
  startup (useful for tests).

#### Env vars  ####

Every file setting can be overriden with the use of environment variables.
The name of the var is the uppercase setting name prefixed with `SLA_`. For
example, to override the check period, set the env var `SLA_CHECKPERIOD`.

### Usage ###

SLALite offers a usual REST API, with an endpoint on /agreements

Add an agreement:

    curl -k -X POST -d @agreement.json https://localhost:8090/agreements

Get agreements:

    curl -k https://localhost:8090/agreements
    curl -k https://localhost:8090/agreements/a02

