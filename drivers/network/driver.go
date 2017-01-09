package network

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/nicescale/qingcloud-docker-network/qcsdk"
	"github.com/nicescale/qingcloud-docker-network/util"
)

var errNotImplemented = fmt.Errorf("not implemented")

const (
	maxIdleNicsPerInstance = 2
)

type endpoint struct {
	ID         string
	NicID      string
	IP         string
	SandboxKey string
}

type netConfig struct {
	ID        string
	Vxnet     string
	Router    string
	IPAMData  *network.IPAMData
	endpoints map[string]*endpoint
	gateway   string
	mu        sync.Mutex
}

func (n *netConfig) getEndpoint(id string) *endpoint {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.endpoints[id]
}

type driver struct {
	api        *qcsdk.Api
	root       string // The root path to store network files.
	mu         sync.Mutex
	lockedNics map[string]bool
	networks   map[string]*netConfig
}

func New(api *qcsdk.Api, root string) (network.Driver, error) {
	driver := &driver{
		api:        api,
		root:       root,
		lockedNics: make(map[string]bool),
		networks:   make(map[string]*netConfig),
	}
	if err := driver.loadNetworks(); err != nil {
		return nil, err
	}
	return driver, nil
}

func (d *driver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	logrus.Debug("network.GetCapabilities called")
	return &network.CapabilitiesResponse{Scope: network.LocalScope}, nil
}

func (d *driver) CreateNetwork(req *network.CreateNetworkRequest) error {
	logrus.WithField("req", req).Debug("network.CreateNetwork")

	opts, ok := req.Options["com.docker.network.generic"].(map[string]interface{})
	if !ok {
		return fmt.Errorf(`must provide "-o vxnet=xxx" option`)
	}
	vxnet, ok := opts["vxnet"].(string)
	if !ok || vxnet == "" {
		return fmt.Errorf(`must provide "-o vxnet=xxx" option`)
	}

	n := &netConfig{
		ID:        req.NetworkID,
		Router:    "",
		Vxnet:     vxnet,
		IPAMData:  req.IPv4Data[0],
		endpoints: make(map[string]*endpoint),
	}
	if err := d.writeFile(d.netConfigPath(n.ID), n); err != nil {
		return err
	}

	d.mu.Lock()
	d.networks[n.ID] = n
	d.mu.Unlock()
	return nil
}

func (d *driver) AllocateNetwork(req *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	logrus.WithField("req", req).Debug("network.AllocateNetwork")
	return nil, errNotImplemented
}

func (d *driver) DeleteNetwork(req *network.DeleteNetworkRequest) error {
	logrus.WithField("req", req).Debug("network.DeleteNetwork")
	n := d.getNetwork(req.NetworkID)
	if n == nil {
		return nil
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.endpoints) > 1 {
		return fmt.Errorf("can't delete the network because there are %d active endpoints within the network.", len(n.endpoints))
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	return os.RemoveAll(d.netConfigDir(n.ID))
}

func (d *driver) FreeNetwork(req *network.FreeNetworkRequest) error {
	logrus.WithField("req", req).Debug("network.FreeNetwork")
	return errNotImplemented
}

func (d *driver) CreateEndpoint(req *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	logrus.WithField("req", req).Debug("network.CreateEndpoint")
	ip := ""
	if req.Interface != nil {
		ip = req.Interface.Address
	}
	n := d.getNetwork(req.NetworkID)
	if n == nil {
		return nil, fmt.Errorf("network %s not found", req.NetworkID)
	}

	ep, err := d.findAvailableNic(req.EndpointID, n.Vxnet, ip)
	if err != nil {
		return nil, err
	}

	if err := d.saveEndpoint(n.ID, ep); err != nil {
		return nil, err
	}

	n.mu.Lock()
	n.endpoints[req.EndpointID] = ep
	n.mu.Unlock()

	iface := &network.EndpointInterface{MacAddress: ep.NicID}
	if req.Interface == nil || req.Interface.Address == "" {
		iface.Address = ep.IP
	}

	resp := &network.CreateEndpointResponse{
		Interface: iface,
	}
	return resp, nil
}

func (d *driver) DeleteEndpoint(req *network.DeleteEndpointRequest) error {
	logrus.WithField("req", req).Debug("network.DeleteEndpoint")
	n := d.getNetwork(req.NetworkID)
	if n == nil {
		return fmt.Errorf("network %s not found", req.NetworkID)
	}
	ep := n.getEndpoint(req.EndpointID)
	if ep == nil {
		return nil
	}

	if ep.SandboxKey != "" {
		return fmt.Errorf("endpoint %s is used by another container", ep.ID)
	}

	links, err := util.LinkList()
	if err != nil {
		logrus.Errorf("Failed to get link list in DeleteEndpoint: %v", err)
	} else {
		for mac, l := range links {
			if l.Type() != "device" || l.Attrs().Name == "lo" {
				delete(links, mac)
				continue
			}
		}
		// -1 means exclude the main nic of the VM
		if len(links)-1 > maxIdleNicsPerInstance {
			if jobID, err := d.api.DetachNics([]string{ep.NicID}, false); err != nil {
				logrus.Errorf("failed to detach nic %s. job_id: %s, err: %v", ep.NicID, jobID, err)
			}
		}
	}

	n.mu.Lock()
	delete(n.endpoints, ep.ID)
	if err := os.Remove(d.epConfigPath(n.ID, ep.ID)); err != nil {
		logrus.Errorf("remove endpoint file failed: %v", err)
	}
	n.mu.Unlock()

	return nil
}

func (d *driver) EndpointInfo(req *network.InfoRequest) (*network.InfoResponse, error) {
	logrus.WithField("req", req).Debug("network.EndpointInfo")
	return nil, nil
}

func (d *driver) Join(req *network.JoinRequest) (*network.JoinResponse, error) {
	logrus.WithField("req", req).Debug("network.Join")
	n := d.getNetwork(req.NetworkID)
	if n == nil {
		return nil, fmt.Errorf("network %s not found", req.NetworkID)
	}
	ep := n.getEndpoint(req.EndpointID)
	if ep == nil {
		return nil, fmt.Errorf("no such endpoint")
	}

	if ep.SandboxKey != "" {
		return nil, fmt.Errorf("endpoint %s is used by another container", ep.ID)
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	ep.SandboxKey = req.SandboxKey
	if err := d.saveEndpoint(n.ID, ep); err != nil {
		return nil, err
	}

	resp := &network.JoinResponse{
		InterfaceName: network.InterfaceName{
			SrcName:   genNicName(req.EndpointID),
			DstPrefix: "eth",
		},
		Gateway: strings.Split(n.IPAMData.Gateway, "/")[0],
	}
	return resp, nil
}

func (d *driver) Leave(req *network.LeaveRequest) error {
	logrus.WithField("req", req).Debug("network.Leave")
	n := d.getNetwork(req.NetworkID)
	if n == nil {
		return fmt.Errorf("network %s not found", req.NetworkID)
	}
	ep := n.getEndpoint(req.EndpointID)
	if ep == nil {
		return nil
	}

	if ep.SandboxKey == "" {
		return nil
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	ep.SandboxKey = ""
	if err := d.saveEndpoint(n.ID, ep); err != nil {
		return err
	}
	return nil
}

func (d *driver) DiscoverNew(req *network.DiscoveryNotification) error {
	logrus.WithField("req", req).Debug("network.DiscoverNew")
	return errNotImplemented
}

func (d *driver) DiscoverDelete(req *network.DiscoveryNotification) error {
	logrus.WithField("req", req).Debug("network.DiscoverDelete")
	return errNotImplemented
}

func (d *driver) ProgramExternalConnectivity(req *network.ProgramExternalConnectivityRequest) error {
	logrus.WithField("req", req).Debug("network.ProgramExternalConnectivity")
	return nil
}

func (d *driver) RevokeExternalConnectivity(req *network.RevokeExternalConnectivityRequest) error {
	logrus.WithField("req", req).Debug("network.RevokeExternalConnectivity")
	return nil
}
