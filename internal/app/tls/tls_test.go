package tls

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTLSCert(t *testing.T) {
	const (
		testCertPath = "test-cert.pem"
		testKeyPath  = "test-key.pem"
	)
	err := CreateTLSCert(testCertPath, testKeyPath)
	require.NoError(t, err)
	_, err = os.Stat(testCertPath)
	require.NoError(t, err)
	_, err = os.Stat(testKeyPath)
	require.NoError(t, err)
	_ = os.Remove(testCertPath)
	_ = os.Remove(testKeyPath)
}
