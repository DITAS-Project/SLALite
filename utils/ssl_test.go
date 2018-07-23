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
This test makes use of the certificate stored in testdata/. In case the certificate
expires or there is other problem, generate a new self-signed cert with:

	openssl req -nodes -new -x509 -keyout key.pem -out cert.pem -days 3650 -subj "/C=/ST=/L=/O=/CN=localhost"

You can check the certificate info with:

	openssl x509 -in cert.pem  -text -noout
*/
import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	result := m.Run()

	os.Exit(result)
}

func TestAddTrustedCA(t *testing.T) {
	cfg := viper.New()

	AddTrustedCAs(cfg)

	cfg.Set(CAPathPropertyName, "testdata/cert.pem")
	AddTrustedCAs(cfg)

	cfg.Set(CAPathPropertyName, "testdata/notexists.pem")
	AddTrustedCAs(cfg)

	cfg.Set(CAPathPropertyName, "testdata/wrongcert.pem")
	AddTrustedCAs(cfg)
}

func TestConnectToServer(t *testing.T) {
	/*
	 * Prepare client configuration
	 */
	cfg := viper.New()
	cfg.Set(CAPathPropertyName, "testdata/cert.pem")
	AddTrustedCAs(cfg)
	var cl *http.Client
	cl = GetClient(false)

	/*
	 * Prepare server configuration
	 */
	certPath := "testdata/cert.pem"
	keyPath := "testdata/key.pem"
	x509Cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatalf("Error reading x509 key pair. cert: %s  key: %s", certPath, keyPath)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(f))
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{x509Cert},
	}
	server.StartTLS()
	defer server.Close()

	/*
	 * Make request to test server (certificate issued to localhost)
	 * Documented here: https://ericchiang.github.io/post/go-tls/
	 */
	req, err := http.NewRequest("GET", strings.Replace(server.URL, "127.0.0.1", "localhost", 1), nil)
	resp, err := cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

func f(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
