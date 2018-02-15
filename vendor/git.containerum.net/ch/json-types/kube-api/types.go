package kube_api

type Protocol string
const (
	UDP Protocol = "UDP"
	TCP Protocol = "TCP"
)

type Endpoint struct {
	Name      string   `json:"name" binding:"required"`
	Owner     *string  `json:"owner,omitempty" binding:"required"`
	CreatedAt *int64   `json:"created_at,omitempty"`
	Addresses []string `json:"addresses" binding:"required"`
	Ports     []Port   `json:"ports" binding:"required"`
}

type Port struct {
	Name       string   `json:"name" binding:"required"`
	Port       int      `json:"port" binding:"required"`
	Protocol   Protocol `json:"protocol" binding:"required"`
}
