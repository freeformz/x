package mustcert

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
)

// This example uses mustcert to create certificates, start a TLS server, and a client to
// talk to it.
func Example() {
	ca := CA("root", nil)
	serverCert := Leaf("localhost", ca)
	clientCert := Leaf("client", ca)

	// Create the TLS Test Server
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello, world!"))
	}))

	rootCAs := Pool(ca.TLS())
	server.TLS = &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{*serverCert.TLS()},
		RootCAs: rootCAs,
		ClientCAs: rootCAs,
	}
	server.StartTLS()
	defer server.Close()

	// Create the Client configuration
	cert, err := tls.X509KeyPair([]byte(clientCert.CertPEM()), []byte(clientCert.KeyPEM()))
	if err != nil {
		fmt.Println(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(ca.CertPEM()))
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs: caCertPool,
		InsecureSkipVerify: true,
	}
	config.BuildNameToCertificate()

	// Create the HTTP Client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		},
	}

	// Make a client request to the HTTP Server
	resp, err := client.Get(server.URL)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)

	// Output:
	// hello, world!
}
