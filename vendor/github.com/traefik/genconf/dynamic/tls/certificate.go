package tls

// CertAndStores holds a certificate and its associated stores.
type CertAndStores struct {
	Certificate
	Stores []string `json:"stores,omitempty"`
}

// Certificate holds a SSL cert/key pair.
type Certificate struct {
	CertFile string `json:"certFile,omitempty"`
	KeyFile  string `json:"keyFile,omitempty"`
}
