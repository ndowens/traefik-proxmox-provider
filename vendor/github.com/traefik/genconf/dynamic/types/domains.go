package types

// Domain holds a domain name with optional SANs.
type Domain struct {
	Main string   `json:"main,omitempty"`
	SANs []string `json:"sans,omitempty"`
}
