package dynamic

import (
	"github.com/traefik/genconf/dynamic/tls"
)

// Configuration is the root of the dynamic configuration.
type Configuration struct {
	HTTP *HTTPConfiguration `json:"http,omitempty"`
	TCP  *TCPConfiguration  `json:"tcp,omitempty"`
	UDP  *UDPConfiguration  `json:"udp,omitempty"`
	TLS  *TLSConfiguration  `json:"tls,omitempty"`
}

// TLSConfiguration contains all the TLS configuration parameters.
type TLSConfiguration struct {
	Certificates []*tls.CertAndStores       `json:"certificates,omitempty"`
	Options      map[string]tls.Options     `json:"options,omitempty"`
	Stores       map[string]tls.Store       `json:"stores,omitempty"`
}
