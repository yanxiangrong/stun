package utils

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

func GeneratePort() int {
	return rand.Intn(48128) + 1024
}

func TimeStr() string {
	currentTime := time.Now()
	return fmt.Sprintf("%02d:%02d:%02d.%03d",
		currentTime.Hour(),
		currentTime.Hour(),
		currentTime.Second(),
		currentTime.Nanosecond()/1000_000)
}

// GenerateTLSConfig Set up a bare-bones TLS config for the server
func GenerateTLSConfig(alpn string) *tls.Config {
	key, err := rsa.GenerateKey(crand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(crand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{alpn},
	}
}
