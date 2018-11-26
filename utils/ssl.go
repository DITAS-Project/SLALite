/*
Copyright 2018 Atos

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

/*
This file contains code to add trusted CAs to a x509.CertPool and obtain
http.Clients that can connect to servers with certificates issued by that CAs.

Documented here: https://forfuncsake.github.io/post/2017/08/trust-extra-ca-cert-in-go-app/
*/

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// CAPathPropertyName is the name of the viper.config key that contains the path of
	// a file containing a list of certificates to be trusted
	CAPathPropertyName = "CAPath"
)

// x509.CertPool that contains the trusted CAs
var cas *x509.CertPool

/*
AddTrustedCAs add a list of certificates in a PEM file to a
copy of the System Certificate Pool.

Parameters:
- config. A Viper config that contains in the key sslCAPathPropertyName the path
  of the PEM file with the certificates to add.

NOTE: in Windows, the System Certificate Pool cannot be obtained, and new
empty Cert Pool will be created.
*/
func AddTrustedCAs(config *viper.Viper) {

	caPath := config.GetString(CAPathPropertyName)
	if caPath == "" {
		return
	}
	AddTrustedCAsFromFile(caPath)
}

/*
AddTrustedCAsFromFile add a list of certificates in a PEM file to a
copy of the System Certificate Pool.

Parameters:
- path: path of the PEM file containing the certificates
*/
func AddTrustedCAsFromFile(path string) {

	var err error

	if cas == nil {
		cas, err = x509.SystemCertPool()
		if err != nil {
			cas = x509.NewCertPool()
		}
	}

	certs, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warnf("Cannot read CA certicates in %s", path)
		return
	}
	if ok := cas.AppendCertsFromPEM(certs); !ok {
		log.Warnf("Error appending certificates")
	}
	log.Printf("Added trusted CAs to certificates in file %s", path)
}

/*
GetClient returns an *http.Client configured to use the (previously added with
AddTrustedCAs) trusted CAs.

An additional flag insecure is used to bypass the SSL verification (do not use
in production!!)
*/
func GetClient(insecure bool) *http.Client {
	var client = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
				RootCAs:            cas,
			},
		},
	}

	return client
}
