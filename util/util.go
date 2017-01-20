package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/vishvananda/netlink"
)

var (
	InstanceID string
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
