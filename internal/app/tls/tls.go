package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

// CreateTLSCert - generate TLS certificate and key for run server HTTPS
func CreateTLSCert(certPath string, keyPath string) error {
	cert := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"tiny-url"},
			Country:      []string{"RU"},
			Province:     []string{"Moscow"},
			Locality:     []string{"Moscow"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	certBytes, _ := x509.CreateCertificate(rand.Reader, &cert, &cert, &privateKey.PublicKey, privateKey)
	err := saveToFile(certPath, "CERTIFICATE", certBytes)
	if err != nil {
		return err
	}

	err = saveToFile(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privateKey))
	if err != nil {
		return err
	}

	return nil
}

func saveToFile(filePath string, cypherType string, cypher []byte) error {
	var (
		buf  bytes.Buffer
		file *os.File
	)

	_ = pem.Encode(&buf, &pem.Block{
		Type:  cypherType,
		Bytes: cypher,
	})

	file, _ = os.Create(filePath)
	defer file.Close()

	_, err := buf.WriteTo(file)
	if err != nil {
		return err
	}

	return nil
}
