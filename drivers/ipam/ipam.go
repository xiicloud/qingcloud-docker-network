package ipam

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/nicescale/qingcloud-docker-network/qcsdk"
	sdktypes "github.com/nicescale/qingcloud-docker-network/qcsdk/types"
	"github.com/nicescale/qingcloud-docker-network/util"
)

var errNoAvailableNic = fmt.Errorf("no available nic")

type driver struct {
	api     *qcsdk.Api
	nicLock sync.Mutex
}

func New(api *qcsdk.Api) ipam.Ipam {
	return &driver{
		api: api,
	}
}

func (d *driver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	logrus.Debug("ipam.GetCapabilities called")
	return &ipam.CapabilitiesResponse{}, nil
}

func (d *driver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	logrus.Debug("ipam.GetDefaultAddressSpaces called")
	return &ipam.AddressSpacesResponse{
		LocalDefaultAddressSpace:  "qingcloud-local",
		GlobalDefaultAddressSpace: "none",
	}, nil
}

func (d *driver) RequestPool(req *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	logrus.WithField("req", req).Debug("ipam.RequestPool called")
	if req.Options == nil || req.Options["vxnet"] == "" {
		return nil, fmt.Errorf("--ipam-opt vxnet=xxx must be provided")
	}
	return &ipam.RequestPoolResponse{
		PoolID: req.Options["vxnet"],
		Pool:   req.Pool,
	}, nil
}

func (d *driver) ReleasePool(req *ipam.ReleasePoolRequest) error {
	logrus.WithField("req", req).Debug("ipam.ReleasePool called")
	return nil
}

func (d *driver) RequestAddress(req *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	logrus.WithField("req", req).Debug("ipam.RequestAddress called")
	if req.Options != nil && req.Options["RequestAddressType"] == "com.docker.network.gateway" {
		return &ipam.RequestAddressResponse{
			Address: req.Address + "/24", // TODO: call qingcloud API to get the mask
		}, nil
	}

	nic, err := d.findOrCreateNic(req.PoolID, req.Address)
	if err != nil {
		return nil, err
	}

	links, err := util.LinkList()
	if err != nil {
		return nil, err
	}
	util.NicStore.Add(nic.PrivateIP.String(), links[nic.ID])

	return &ipam.RequestAddressResponse{
		Address: nic.PrivateIP.String() + "/24",
	}, nil
}

func (d *driver) ReleaseAddress(req *ipam.ReleaseAddressRequest) error {
	logrus.WithField("req", req).Debug("ipam.ReleaseAddress called")
	return nil
}

func (d *driver) findOrCreateNic(vxnet, ip string) (*sdktypes.Nic, error) {
	nic, err := d.findAttachedIdleNic(vxnet, ip)
	if err == nil {
		return nic, nil
	}

	var ips []string
	if ip == "" {
		nic, err = d.findRandomAvailableNic(vxnet)
	} else {
		nic, err = d.findAvailableNicByIP(vxnet, ip)
		ips = append(ips, ip)
	}
	if err != nil && err != errNoAvailableNic {
		return nil, err
	}

	// Create a network interface and attach it to the instance.
	if err == errNoAvailableNic {
		nics, err := d.api.CreateNics(vxnet, "", 1, ips)
		if err != nil {
			return nil, err
		}
		nic = nics[0]
	}

	if _, err := d.api.AttachNics([]string{nic.ID}, util.InstanceID, true); err != nil {
		return nil, err
	}
	return nic, nil
}

func (d *driver) findAttachedIdleNic(vxnet, ip string) (*sdktypes.Nic, error) {
	nics, err := d.api.DescribeNics(qcsdk.Params{"vxnets": vxnet, "instances": util.InstanceID})
	if err != nil {
		return nil, err
	}

	m, err := util.LinkList()
	if err != nil {
		return nil, err
	}

	var preferedNic *sdktypes.Nic
	d.nicLock.Lock()
	defer d.nicLock.Unlock()
	for _, nic := range nics {
		if nic.Role == 1 {
			// Role == 1 means the interface is used by the VM
			continue
		}
		if m[nic.ID] == nil {
			continue
		}
		if ip != "" && nic.PrivateIP.String() != ip {
			continue
		}

		preferedNic = nic
		break
	}

	if preferedNic == nil {
		return nil, errNoAvailableNic
	}
	return preferedNic, nil
}

func (d *driver) findRandomAvailableNic(vxnet string) (*sdktypes.Nic, error) {
	nics, err := d.api.DescribeNics(qcsdk.Params{"status": "available", "vxnets": vxnet})
	if err != nil {
		return nil, err
	}

	if len(nics) == 0 {
		return nil, errNoAvailableNic
	}

	rand.Seed(time.Now().UnixNano())
	return nics[rand.Intn(len(nics))], nil
}

func (d *driver) findAvailableNicByIP(vxnet, ip string) (*sdktypes.Nic, error) {
	nics, err := d.api.DescribeNics(qcsdk.Params{
		"status":      "available",
		"search_word": ip,
		"vxnets":      vxnet})
	if err != nil {
		return nil, err
	}
	if len(nics) == 0 {
		return nil, errNoAvailableNic
	}

	for _, nic := range nics {
		if nic.PrivateIP.String() == ip {
			return nic, nil
		}
	}

	return nil, errNoAvailableNic
}
