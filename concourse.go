package main

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"time"

	"github.com/concourse/atc"
	"golang.org/x/oauth2"
)

func defaultHttpClient(token *atc.AuthToken, insecure bool, caCertPool *x509.CertPool) *http.Client {
	var oAuthToken *oauth2.Token
	if token != nil {
		oAuthToken = &oauth2.Token{
			TokenType:   token.Type,
			AccessToken: token.Value,
		}
	}

	transport := transport(insecure, caCertPool)

	if token != nil {
		transport = &oauth2.Transport{
			Source: oauth2.StaticTokenSource(oAuthToken),
			Base:   transport,
		}
	}

	return &http.Client{Transport: transport}
}

func basicAuthHttpClient(
	username string,
	password string,
	insecure bool,
	caCertPool *x509.CertPool,
) *http.Client {
	return &http.Client{
		Transport: basicAuthTransport{
			username: username,
			password: password,
			base:     transport(insecure, caCertPool),
		},
	}
}

func transport(insecure bool, caCertPool *x509.CertPool) http.RoundTripper {
	var transport http.RoundTripper

	transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
			RootCAs:            caCertPool,
		},
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}

	return transport
}

type basicAuthTransport struct {
	username string
	password string

	base http.RoundTripper
}

func (t basicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.username, t.password)
	return t.base.RoundTrip(r)
}
