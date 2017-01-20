package types

import (
	"net"
	"time"
)

type Instance struct {
	VcpusCurrent int      `json:"vcpus_current"`
	ID           string   `json:"instance_id"`
	VolumeIds    []string `json:"volume_ids"`
	Vxnets       []struct {
		Name      string `json:"vxnet_name"`
		Type      int    `json:"vxnet_type"`
		ID        string `json:"vxnet_id"`
		NicID     string `json:"nic_id"`
		PrivateIp net.IP `json:"private_ip"`
	} `json:"vxnets"`
	MemoryCurrent    int       `json:"memory_current"`
	SubCode          int       `json:"sub_code"`
	TransitionStatus string    `json:"transition_status"`
	Name             string    `json:"instance_name"`
	Type             string    `json:"instance_type"`
	CreateTime       time.Time `json:"create_time"`
	Status           string    `json:"status"`
	Description      string    `json:"description"`
	SecurityGroup    struct {
		IsDefault int    `json:"is_default"`
		ID        string `json:"security_group_id"`
	} `json:"security_group"`
	StatusTime time.Time `json:"status_time"`
	Image      struct {
		ProcessorType string `json:"processor_type"`
		Platform      string `json:"platform"`
		Size          int    `json:"image_size"`
		Name          string `json:"image_name"`
		ID            string `json:"image_id"`
		OsFamily      string `json:"os_family"`
		Provider      string `json:"provider"`
	} `json:"image"`
	KeypairIds []string `json:"keypair_ids"`
}

type DescribeInstancesResponse struct {
	ResponseStatus
	Total     int         `json:"total_count"`
	Instances []*Instance `json:"instance_set"`
}
