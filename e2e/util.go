package e2e

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"text/template"
	"time"
)

type values struct {
	PrivSk   string
	SignCert string
	MtlsCrt  string
	MtlsKey  string
	Mtls     bool
	Addr     string
}

func generateCertAndKeys(key, cert *os.File) error {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	privDer, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	err = pem.Encode(key, &pem.Block{Type: "PRIVATE KEY", Bytes: privDer})
	if err != nil {
		return err
	}

	template := &x509.Certificate{
		SerialNumber: new(big.Int),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDer, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		return err
	}
	err = pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: certDer})
	if err != nil {
		return err
	}
	return nil
}

func generateConfigFile(fileName string, values values) {
	var Text = `# Definition of nodes
node: &node
  addr: {{ .Addr }}
  {{ if .Mtls }}
  tls_ca_cert: {{.MtlsCrt}}
  tls_ca_key: {{.MtlsKey}}
  tls_ca_root: {{.MtlsCrt}}
  {{ end }}
# Nodes to interact with
endorsers:
  - *node
committer: *node
orderer: *node
channel: test-channel
chaincode: test-chaincode
mspid: Org1MSP
private_key: {{.PrivSk}}
sign_cert: {{.SignCert}}
num_of_conn: 10
client_per_conn: 10
`
	tmpl, err := template.New("test").Parse(Text)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = tmpl.Execute(file, values)
	if err != nil {
		panic(err)
	}
}
