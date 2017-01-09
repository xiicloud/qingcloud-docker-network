package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

const qingcloudMetadataURL = "http://metadata.ks.qingcloud.com/"

var (
	InstanceID string
	inspectURL string
	httpClient = &http.Client{Timeout: time.Second * 5}
	NlHandle   *netlink.Handle
)

func Init() {
	id, err := ioutil.ReadFile("/etc/qingcloud/instance-id")
	if err != nil {
		id, err = ioutil.ReadFile("/proc/sys/kernel/hostname")
	}
	if err != nil {
		msg := fmt.Sprintf("can't find instance ID from /etc/qingcloud/instance-id: %v", err)
		panic(msg)
	}

	if len(id) < 10 || id[0] != 'i' || id[1] != '-' {
		msg := fmt.Sprintf("invalid instance id %s. Are you running in a qingcloud VM?", string(id))
		panic(msg)
	}
	InstanceID = strings.TrimSpace(string(id))
	inspectURL = qingcloudMetadataURL + InstanceID

	NlHandle, err = netlink.NewHandle()
	if err != nil {
		msg := fmt.Sprintf("failed to create netlink handle: %v", err)
		panic(msg)
	}
}

func ReadJSON(path string, data interface{}) error {
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	return json.NewDecoder(fp).Decode(data)
}

type VxnetNic struct {
	VxnetType int    `json:"vxnet_type"`
	VxnetID   string `json:"vxnet_id"`
	VxnetName string `json:"vxnet_name"`
	Role      int    `json:"role"`
	PrivateIP net.IP `json:"private_ip"`
	NicID     string `json:"nic_id"`
}

type InstanceMetadata struct {
	VcpusCurrent  int           `json:"vcpus_current"`
	InstanceName  string        `json:"instance_name"`
	VxnetsCount   int           `json:"vxnets_count"`
	VolumeIds     []interface{} `json:"volume_ids"`
	Vxnets        []*VxnetNic   `json:"vxnets"`
	MemoryCurrent int           `json:"memory_current"`
	Eip           struct {
		EipID     string `json:"eip_id"`
		Bandwidth int    `json:"bandwidth"`
		EipAddr   string `json:"eip_addr"`
	} `json:"eip"`
	ImageID      string `json:"image_id"`
	InstanceID   string `json:"instance_id"`
	InstanceType string `json:"instance_type"`
	OsFamily     string `json:"os_family"`
	Platform     string `json:"platform"`
}

func GetInstanceMetedata() (*InstanceMetadata, error) {
	resp, err := httpClient.Get(inspectURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	md := &InstanceMetadata{}
	err = json.NewDecoder(resp.Body).Decode(md)
	return md, err
}

func LinkList() (map[string]netlink.Link, error) {
	links, err := NlHandle.LinkList()
	if err != nil {
		return nil, err
	}
	m := make(map[string]netlink.Link)
	for _, n := range links {
		m[n.Attrs().HardwareAddr.String()] = n
	}
	return m, nil
}

func RenameLink(link netlink.Link, name string) error {
	if err := NlHandle.LinkSetDown(link); err != nil {
		return err
	}
	return NlHandle.LinkSetName(link, name)
}
