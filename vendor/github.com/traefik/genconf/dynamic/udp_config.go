package dynamic

// UDPConfiguration contains all the UDP configuration parameters.
type UDPConfiguration struct {
	Routers  map[string]*UDPRouter  `json:"routers,omitempty"`
	Services map[string]*UDPService `json:"services,omitempty"`
}

// UDPRouter holds the UDP router configuration.
type UDPRouter struct {
	EntryPoints []string `json:"entryPoints,omitempty"`
	Service     string   `json:"service,omitempty"`
}

// UDPService holds a UDP service configuration.
type UDPService struct {
	LoadBalancer *UDPServersLoadBalancer `json:"loadBalancer,omitempty"`
	Weighted     *UDPWeightedRoundRobin  `json:"weighted,omitempty"`
}

// UDPServersLoadBalancer holds the UDP load balancer configuration.
type UDPServersLoadBalancer struct {
	Servers []UDPServer `json:"servers,omitempty"`
}

// UDPServer holds the UDP server configuration.
type UDPServer struct {
	Address string `json:"address,omitempty"`
	Port    string `json:"port,omitempty"`
}

// UDPWeightedRoundRobin holds the UDP WRR configuration.
type UDPWeightedRoundRobin struct {
	Services []UDPWRRService `json:"services,omitempty"`
}

// UDPWRRService holds one UDP WRR service entry.
type UDPWRRService struct {
	Name   string `json:"name,omitempty"`
	Weight *int   `json:"weight,omitempty"`
}
