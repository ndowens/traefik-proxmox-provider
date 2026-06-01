package dynamic

// TCPConfiguration contains all the TCP configuration parameters.
type TCPConfiguration struct {
	Routers  map[string]*TCPRouter  `json:"routers,omitempty"`
	Services map[string]*TCPService `json:"services,omitempty"`
}

// TCPRouter holds the TCP router configuration.
type TCPRouter struct {
	EntryPoints []string          `json:"entryPoints,omitempty"`
	Service     string            `json:"service,omitempty"`
	Rule        string            `json:"rule,omitempty"`
	TLS         *RouterTCPTLSConfig `json:"tls,omitempty"`
}

// RouterTCPTLSConfig holds the TLS configuration for a TCP router.
type RouterTCPTLSConfig struct {
	Passthrough  bool   `json:"passthrough,omitempty"`
	Options      string `json:"options,omitempty"`
	CertResolver string `json:"certResolver,omitempty"`
}

// TCPService holds a TCP service configuration.
type TCPService struct {
	LoadBalancer *TCPServersLoadBalancer `json:"loadBalancer,omitempty"`
	Weighted     *TCPWeightedRoundRobin  `json:"weighted,omitempty"`
}

// TCPServersLoadBalancer holds the TCP load balancer configuration.
type TCPServersLoadBalancer struct {
	TerminationDelay *int        `json:"terminationDelay,omitempty"`
	Servers          []TCPServer `json:"servers,omitempty"`
}

// TCPServer holds the TCP server configuration.
type TCPServer struct {
	Address string `json:"address,omitempty"`
	Port    string `json:"port,omitempty"`
}

// TCPWeightedRoundRobin holds the TCP WRR configuration.
type TCPWeightedRoundRobin struct {
	Services []TCPWRRService `json:"services,omitempty"`
}

// TCPWRRService holds one TCP WRR service entry.
type TCPWRRService struct {
	Name   string `json:"name,omitempty"`
	Weight *int   `json:"weight,omitempty"`
}
