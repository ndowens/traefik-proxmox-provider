package tls

// Store holds the TLS store configuration.
type Store struct {
	DefaultCertificate *Certificate `json:"defaultCertificate,omitempty"`
}

// Options holds the TLS options configuration.
type Options struct {
	MinVersion               string   `json:"minVersion,omitempty"`
	MaxVersion               string   `json:"maxVersion,omitempty"`
	CipherSuites             []string `json:"cipherSuites,omitempty"`
	CurvePreferences         []string `json:"curvePreferences,omitempty"`
	ClientAuth               ClientAuth `json:"clientAuth,omitempty"`
	SniStrict                bool     `json:"sniStrict,omitempty"`
	PreferServerCipherSuites bool     `json:"preferServerCipherSuites,omitempty"`
}

// ClientAuth holds the TLS client-auth configuration.
type ClientAuth struct {
	CAFiles        []string `json:"caFiles,omitempty"`
	ClientAuthType string   `json:"clientAuthType,omitempty"`
}
