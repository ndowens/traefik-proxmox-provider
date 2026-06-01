package dynamic

import "github.com/traefik/genconf/dynamic/types"

// HTTPConfiguration contains all the HTTP configuration parameters.
type HTTPConfiguration struct {
	Routers           map[string]*Router           `json:"routers,omitempty"`
	Services          map[string]*Service          `json:"services,omitempty"`
	Middlewares       map[string]*Middleware       `json:"middlewares,omitempty"`
	ServersTransports map[string]*ServersTransport `json:"serversTransports,omitempty"`
}

// Router holds the router configuration.
type Router struct {
	EntryPoints []string         `json:"entryPoints,omitempty"`
	Middlewares []string         `json:"middlewares,omitempty"`
	Service     string           `json:"service,omitempty"`
	Rule        string           `json:"rule,omitempty"`
	Priority    int              `json:"priority,omitempty"`
	TLS         *RouterTLSConfig `json:"tls,omitempty"`
}

// RouterTLSConfig holds the TLS configuration for a router.
type RouterTLSConfig struct {
	Options      string         `json:"options,omitempty"`
	CertResolver string         `json:"certResolver,omitempty"`
	Domains      []types.Domain `json:"domains,omitempty"`
}

// Service holds a service configuration (LOAD_BALANCER, WEIGHTED, MIRRORING).
type Service struct {
	LoadBalancer *ServersLoadBalancer `json:"loadBalancer,omitempty"`
	Weighted     *WeightedRoundRobin  `json:"weighted,omitempty"`
	Mirroring    *Mirroring           `json:"mirroring,omitempty"`
}

// ServersLoadBalancer holds the load balancing configuration.
type ServersLoadBalancer struct {
	Sticky             *Sticky              `json:"sticky,omitempty"`
	Servers            []Server             `json:"servers,omitempty"`
	HealthCheck        *ServerHealthCheck   `json:"healthCheck,omitempty"`
	PassHostHeader     *bool                `json:"passHostHeader,omitempty"`
	ResponseForwarding *ResponseForwarding  `json:"responseForwarding,omitempty"`
	ServersTransport   string               `json:"serversTransport,omitempty"`
}

// Server holds the server configuration.
type Server struct {
	URL    string `json:"url,omitempty"`
	Scheme string `json:"scheme,omitempty"`
	Port   string `json:"port,omitempty"`
}

// ServerHealthCheck holds the health-check configuration for a service.
type ServerHealthCheck struct {
	Scheme          string            `json:"scheme,omitempty"`
	Path            string            `json:"path,omitempty"`
	Method          string            `json:"method,omitempty"`
	Port            int               `json:"port,omitempty"`
	Interval        string            `json:"interval,omitempty"`
	Timeout         string            `json:"timeout,omitempty"`
	Hostname        string            `json:"hostname,omitempty"`
	FollowRedirects *bool             `json:"followRedirects,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
}

// Sticky holds the sticky-session configuration.
type Sticky struct {
	Cookie *Cookie `json:"cookie,omitempty"`
}

// Cookie holds the cookie configuration for sticky sessions.
type Cookie struct {
	Name     string `json:"name,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
}

// ResponseForwarding holds the response-forwarding configuration.
type ResponseForwarding struct {
	FlushInterval string `json:"flushInterval,omitempty"`
}

// ServersTransport holds the ServersTransport configuration.
type ServersTransport struct {
	ServerName          string              `json:"serverName,omitempty"`
	InsecureSkipVerify  bool                `json:"insecureSkipVerify,omitempty"`
	RootCAs             []string            `json:"rootCAs,omitempty"`
	MaxIdleConnsPerHost int                 `json:"maxIdleConnsPerHost,omitempty"`
	ForwardingTimeouts  *ForwardingTimeouts `json:"forwardingTimeouts,omitempty"`
}

// ForwardingTimeouts holds the connection timeouts for requests forwarded to a backend server.
type ForwardingTimeouts struct {
	DialTimeout           string `json:"dialTimeout,omitempty"`
	ResponseHeaderTimeout string `json:"responseHeaderTimeout,omitempty"`
	IdleConnTimeout       string `json:"idleConnTimeout,omitempty"`
	ReadIdleTimeout       string `json:"readIdleTimeout,omitempty"`
	PingTimeout           string `json:"pingTimeout,omitempty"`
}

// WeightedRoundRobin holds the weighted round-robin configuration.
type WeightedRoundRobin struct {
	Services    []WRRService `json:"services,omitempty"`
	Sticky      *Sticky      `json:"sticky,omitempty"`
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
}

// WRRService holds the service name and weight for WRR.
type WRRService struct {
	Name   string `json:"name,omitempty"`
	Weight *int   `json:"weight,omitempty"`
}

// HealthCheck holds the health-check configuration for WRR/Mirroring.
type HealthCheck struct{}

// Mirroring holds the mirroring service configuration.
type Mirroring struct {
	Service     string          `json:"service,omitempty"`
	MaxBodySize *int64          `json:"maxBodySize,omitempty"`
	Mirrors     []MirrorService `json:"mirrors,omitempty"`
	HealthCheck *HealthCheck    `json:"healthCheck,omitempty"`
}

// MirrorService holds the mirror target configuration.
type MirrorService struct {
	Name    string `json:"name,omitempty"`
	Percent int    `json:"percent,omitempty"`
}

// Middleware holds the middleware configuration.
type Middleware struct {
	AddPrefix         *AddPrefix         `json:"addPrefix,omitempty"`
	StripPrefix       *StripPrefix       `json:"stripPrefix,omitempty"`
	StripPrefixRegex  *StripPrefixRegex  `json:"stripPrefixRegex,omitempty"`
	ReplacePath       *ReplacePath       `json:"replacePath,omitempty"`
	ReplacePathRegex  *ReplacePathRegex  `json:"replacePathRegex,omitempty"`
	Chain             *Chain             `json:"chain,omitempty"`
	IPWhiteList       *IPWhiteList       `json:"ipWhiteList,omitempty"`
	Headers           *Headers           `json:"headers,omitempty"`
	Errors            *ErrorPage         `json:"errors,omitempty"`
	RateLimit         *RateLimit         `json:"rateLimit,omitempty"`
	RedirectRegex     *RedirectRegex     `json:"redirectRegex,omitempty"`
	RedirectScheme    *RedirectScheme    `json:"redirectScheme,omitempty"`
	BasicAuth         *BasicAuth         `json:"basicAuth,omitempty"`
	DigestAuth        *DigestAuth        `json:"digestAuth,omitempty"`
	ForwardAuth       *ForwardAuth       `json:"forwardAuth,omitempty"`
	InFlightReq       *InFlightReq       `json:"inFlightReq,omitempty"`
	Buffering         *Buffering         `json:"buffering,omitempty"`
	CircuitBreaker    *CircuitBreaker    `json:"circuitBreaker,omitempty"`
	Compress          *Compress          `json:"compress,omitempty"`
	PassTLSClientCert *PassTLSClientCert `json:"passTLSClientCert,omitempty"`
	Retry             *Retry             `json:"retry,omitempty"`
	ContentType       *ContentType       `json:"contentType,omitempty"`
	Plugin            PluginConf         `json:"plugin,omitempty"`
}

// AddPrefix holds the add-prefix middleware configuration.
type AddPrefix struct{ Prefix string `json:"prefix,omitempty"` }

// StripPrefix holds the strip-prefix middleware configuration.
type StripPrefix struct {
	Prefixes   []string `json:"prefixes,omitempty"`
	ForceSlash bool     `json:"forceSlash,omitempty"`
}

// StripPrefixRegex holds the strip-prefix-regex middleware configuration.
type StripPrefixRegex struct{ Regex []string `json:"regex,omitempty"` }

// ReplacePath holds the replace-path middleware configuration.
type ReplacePath struct{ Path string `json:"path,omitempty"` }

// ReplacePathRegex holds the replace-path-regex middleware configuration.
type ReplacePathRegex struct {
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

// Chain holds the chain middleware configuration.
type Chain struct{ Middlewares []string `json:"middlewares,omitempty"` }

// IPWhiteList holds the IP-whitelist middleware configuration.
type IPWhiteList struct {
	SourceRange []string    `json:"sourceRange,omitempty"`
	IPStrategy  *IPStrategy `json:"ipStrategy,omitempty"`
}

// IPStrategy holds the IP strategy configuration.
type IPStrategy struct {
	Depth       int      `json:"depth,omitempty"`
	ExcludedIPs []string `json:"excludedIPs,omitempty"`
}

// Headers holds the headers middleware configuration.
type Headers struct {
	CustomRequestHeaders  map[string]string `json:"customRequestHeaders,omitempty"`
	CustomResponseHeaders map[string]string `json:"customResponseHeaders,omitempty"`
	AccessControlAllowCredentials bool     `json:"accessControlAllowCredentials,omitempty"`
	AccessControlAllowHeaders     []string `json:"accessControlAllowHeaders,omitempty"`
	AccessControlAllowMethods     []string `json:"accessControlAllowMethods,omitempty"`
	AccessControlAllowOriginList  []string `json:"accessControlAllowOriginList,omitempty"`
	AccessControlExposeHeaders    []string `json:"accessControlExposeHeaders,omitempty"`
	AccessControlMaxAge           int64    `json:"accessControlMaxAge,omitempty"`
	AddVaryHeader                 bool     `json:"addVaryHeader,omitempty"`
	AllowedHosts                  []string `json:"allowedHosts,omitempty"`
	HostsProxyHeaders             []string `json:"hostsProxyHeaders,omitempty"`
	SSLRedirect                   bool     `json:"sslRedirect,omitempty"`
	SSLTemporaryRedirect          bool     `json:"sslTemporaryRedirect,omitempty"`
	SSLHost                       string   `json:"sslHost,omitempty"`
	SSLProxyHeaders               map[string]string `json:"sslProxyHeaders,omitempty"`
	SSLForceHost                  bool     `json:"sslForceHost,omitempty"`
	STSSeconds                    int64    `json:"stsSeconds,omitempty"`
	STSIncludeSubdomains          bool     `json:"stsIncludeSubdomains,omitempty"`
	STSPreload                    bool     `json:"stsPreload,omitempty"`
	ForceSTSHeader                bool     `json:"forceSTSHeader,omitempty"`
	FrameDeny                     bool     `json:"frameDeny,omitempty"`
	CustomFrameOptionsValue       string   `json:"customFrameOptionsValue,omitempty"`
	ContentTypeNosniff            bool     `json:"contentTypeNosniff,omitempty"`
	BrowserXSSFilter              bool     `json:"browserXssFilter,omitempty"`
	CustomBrowserXSSValue         string   `json:"customBrowserXSSValue,omitempty"`
	ContentSecurityPolicy         string   `json:"contentSecurityPolicy,omitempty"`
	PublicKey                     string   `json:"publicKey,omitempty"`
	ReferrerPolicy                string   `json:"referrerPolicy,omitempty"`
	FeaturePolicy                 string   `json:"featurePolicy,omitempty"`
	IsDevelopment                 bool     `json:"isDevelopment,omitempty"`
}

// ErrorPage holds the error-page middleware configuration.
type ErrorPage struct {
	Status  []string `json:"status,omitempty"`
	Service string   `json:"service,omitempty"`
	Query   string   `json:"query,omitempty"`
}

// RateLimit holds the rate-limit middleware configuration.
type RateLimit struct {
	Average int64      `json:"average,omitempty"`
	Period  string     `json:"period,omitempty"`
	Burst   int64      `json:"burst,omitempty"`
	SourceCriterion *SourceCriterion `json:"sourceCriterion,omitempty"`
}

// SourceCriterion holds the source criterion configuration.
type SourceCriterion struct {
	IPStrategy        *IPStrategy `json:"ipStrategy,omitempty"`
	RequestHeaderName string      `json:"requestHeaderName,omitempty"`
	RequestHost       bool        `json:"requestHost,omitempty"`
}

// RedirectRegex holds the redirect-regex middleware configuration.
type RedirectRegex struct {
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
	Permanent   bool   `json:"permanent,omitempty"`
}

// RedirectScheme holds the redirect-scheme middleware configuration.
type RedirectScheme struct {
	Scheme    string `json:"scheme,omitempty"`
	Port      string `json:"port,omitempty"`
	Permanent bool   `json:"permanent,omitempty"`
}

// BasicAuth holds the basic-auth middleware configuration.
type BasicAuth struct {
	Users        Users  `json:"users,omitempty"`
	UsersFile    string `json:"usersFile,omitempty"`
	Realm        string `json:"realm,omitempty"`
	RemoveHeader bool   `json:"removeHeader,omitempty"`
	HeaderField  string `json:"headerField,omitempty"`
}

// Users is a list of users.
type Users []string

// DigestAuth holds the digest-auth middleware configuration.
type DigestAuth struct {
	Users        Users  `json:"users,omitempty"`
	UsersFile    string `json:"usersFile,omitempty"`
	RemoveHeader bool   `json:"removeHeader,omitempty"`
	Realm        string `json:"realm,omitempty"`
	HeaderField  string `json:"headerField,omitempty"`
}

// ForwardAuth holds the forward-auth middleware configuration.
type ForwardAuth struct {
	Address             string     `json:"address,omitempty"`
	TLS                 *ClientTLS `json:"tls,omitempty"`
	TrustForwardHeader  bool       `json:"trustForwardHeader,omitempty"`
	AuthResponseHeaders []string   `json:"authResponseHeaders,omitempty"`
}

// ClientTLS holds the client TLS configuration.
type ClientTLS struct {
	CA                 string `json:"ca,omitempty"`
	CAOptional         bool   `json:"caOptional,omitempty"`
	Cert               string `json:"cert,omitempty"`
	Key                string `json:"key,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"`
}

// InFlightReq holds the in-flight-req middleware configuration.
type InFlightReq struct {
	Amount          int64            `json:"amount,omitempty"`
	SourceCriterion *SourceCriterion `json:"sourceCriterion,omitempty"`
}

// Buffering holds the buffering middleware configuration.
type Buffering struct {
	MaxRequestBodyBytes  int64  `json:"maxRequestBodyBytes,omitempty"`
	MemRequestBodyBytes  int64  `json:"memRequestBodyBytes,omitempty"`
	MaxResponseBodyBytes int64  `json:"maxResponseBodyBytes,omitempty"`
	MemResponseBodyBytes int64  `json:"memResponseBodyBytes,omitempty"`
	RetryExpression      string `json:"retryExpression,omitempty"`
}

// CircuitBreaker holds the circuit-breaker middleware configuration.
type CircuitBreaker struct{ Expression string `json:"expression,omitempty"` }

// Compress holds the compress middleware configuration.
type Compress struct{ ExcludedContentTypes []string `json:"excludedContentTypes,omitempty"` }

// PassTLSClientCert holds the pass-TLS-client-cert middleware configuration.
type PassTLSClientCert struct {
	PEM  bool                      `json:"pem,omitempty"`
	Info *TLSClientCertificateInfo `json:"info,omitempty"`
}

// TLSClientCertificateInfo holds the client certificate info fields.
type TLSClientCertificateInfo struct {
	NotAfter  bool                          `json:"notAfter,omitempty"`
	NotBefore bool                          `json:"notBefore,omitempty"`
	Sans      bool                          `json:"sans,omitempty"`
	Subject   *TLSCLientCertificateDNInfo   `json:"subject,omitempty"`
	Issuer    *TLSCLientCertificateDNInfo   `json:"issuer,omitempty"`
	Serial    bool                          `json:"serialNumber,omitempty"`
}

// TLSCLientCertificateDNInfo holds the DN info fields.
type TLSCLientCertificateDNInfo struct {
	Country         bool `json:"country,omitempty"`
	Province        bool `json:"province,omitempty"`
	Locality        bool `json:"locality,omitempty"`
	Organization    bool `json:"organization,omitempty"`
	CommonName      bool `json:"commonName,omitempty"`
	SerialNumber    bool `json:"serialNumber,omitempty"`
	DomainComponent bool `json:"domainComponent,omitempty"`
}

// Retry holds the retry middleware configuration.
type Retry struct {
	Attempts        int    `json:"attempts,omitempty"`
	InitialInterval string `json:"initialInterval,omitempty"`
}

// ContentType holds the content-type middleware configuration.
type ContentType struct{ AutoDetect bool `json:"autoDetect,omitempty"` }

// PluginConf is a map of plugin configurations.
type PluginConf map[string]interface{}

// Model holds the model configuration.
type Model struct {
	Middlewares []string         `json:"middlewares,omitempty"`
	TLS         *RouterTLSConfig `json:"tls,omitempty"`
}
