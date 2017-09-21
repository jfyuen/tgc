package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
)

// TLSCert stores paths for a CA, a Cert and key file pair
type TLSCert struct {
	ca, crt, key string
}

func (c TLSCert) createCerts() (*x509.CertPool, *tls.Certificate, error) {
	caCertPEM, err := ioutil.ReadFile(c.ca)
	if err != nil {
		return nil, nil, err
	}

	cas := x509.NewCertPool()
	ok := cas.AppendCertsFromPEM(caCertPEM)
	if !ok {
		return nil, nil, errors.New("failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair(c.crt, c.key)
	if err != nil {
		return nil, nil, err
	}
	return cas, &cert, nil
}

// CreateClientConfig creates a tls.Config for a client connection
func (c TLSCert) CreateClientConfig() (*tls.Config, error) {
	cas, cert, err := c.createCerts()
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
		RootCAs:      cas,
	}, nil
}

// CreateServerConfig creates a tls.Config for a server listening, and will authenticate clients
func (c TLSCert) CreateServerConfig() (*tls.Config, error) {
	cas, cert, err := c.createCerts()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ClientCAs:    cas,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}
