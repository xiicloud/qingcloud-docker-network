package types

import (
	"net"
	"time"
)

type Vxnet struct {
	Type        int       `json:"vxnet_type"`
	Description string    `json:"description"`
	Name        string    `json:"vxnet_name"`
	CreateTime  time.Time `json:"create_time"`
	InstanceIDs []string  `json:"instance_ids"`
	Router      struct {
		ID         string    `json:"router_id"`
		Name       string    `json:"router_name"`
		ManagerIP  net.IP    `json:"manager_ip"`
		IPNetwork  net.IPNet `json:"ip_network"`
		DynIPEnd   net.IP    `json:"dyn_ip_end"`
		DynIPStart net.IP    `json:"dyn_ip_start"`
		Mode       int       `json:"mode"`
	} `json:"router"`
	ID string `json:"vxnet_id"`
}

type DescribeVxnetsResponse struct {
	ResponseStatus
	Total  int      `json:"total_count"`
	Vxnets []*Vxnet `json:"vxnet_set"`
}
