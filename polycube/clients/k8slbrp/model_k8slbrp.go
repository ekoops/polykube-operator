/*
 * k8slbrp API
 *
 * k8slbrp API generated from k8slbrp.yang
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

type K8sLbrp struct {
	// Name of the k8slbrp service
	Name string `json:"name,omitempty"`
	// UUID of the Cube
	Uuid string `json:"uuid,omitempty"`
	// Type of the Cube (TC, XDP_SKB, XDP_DRV)
	Type_       string `json:"type,omitempty"`
	ServiceName string `json:"service-name,omitempty"`
	// Defines the logging level of a service instance, from none (OFF) to the most verbose (TRACE)
	Loglevel string `json:"loglevel,omitempty"`
	// Entry of the ports table
	Ports []Ports `json:"ports,omitempty"`
	// Defines if the service is visible in Linux
	Shadow bool `json:"shadow,omitempty"`
	// Defines if all traffic is sent to Linux
	Span bool `json:"span,omitempty"`
	// If configured, when a client request arrives to the LB, the source IP addrress is replaced with another IP address from the 'new' range
	SrcIpRewrite *SrcIpRewrite `json:"src-ip-rewrite,omitempty"`
	// Services (i.e., virtual ip:protocol:port) exported to the client
	Service []Service `json:"service,omitempty"`
	// K8s lbrp mode of operation. 'MULTI' allows to manage multiple FRONTEND port. 'SINGLE' is optimized for working with a single FRONTEND port
	PortMode_ string `json:"port_mode,omitempty"`
}
