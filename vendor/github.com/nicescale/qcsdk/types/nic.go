package types

import (
	"net"
	"time"
)

type Nic struct {
	VxnetID       string    `json:"vxnet_id"`
	NicName       string    `json:"nic_name"`
	Status        string    `json:"status"`
	Tags          []string  `json:"tags"`
	Role          int       `json:"role"`
	Sequence      int       `json:"sequence"`
	InstanceID    string    `json:"instance_id"`
	PrivateIP     net.IP    `json:"private_ip"`
	SecurityGroup string    `json:"security_group"`
	ID            string    `json:"nic_id"`
	StatusTime    time.Time `json:"status_time"`
	CreateTime    time.Time `json:"create_time"`
}

type DescribeNicsResponse struct {
	ResponseStatus
	Total int    `json:"total_count"`
	Nics  []*Nic `json:"nic_set"`
}

type CreateNicResponse struct {
	ResponseStatus
	Nics []*Nic `json:"nics"`
}

type NicActionResponse struct {
	ResponseStatus
	JobID string `json:"job_id"`
}
