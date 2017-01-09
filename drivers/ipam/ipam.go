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

	nic, err := d.findOrCreateNic(req.PoolID)
	if err != nil {
		return nil, err
	}

	return &ipam.RequestAddressResponse{
		Address: nic.PrivateIP.String() + "/24",
	}, nil
}

func (d *driver) ReleaseAddress(req *ipam.ReleaseAddressRequest) error {
	logrus.WithField("req", req).Debug("ipam.ReleaseAddress called")
	return nil
}

func (d *driver) findOrCreateNic(vxnet string) (*sdktypes.Nic, error) {
	if nic, err := d.findAttachedIdleNic(vxnet); err == nil {
		return nic, nil
	}

	nics, err := d.api.DescribeNics(qcsdk.Params{"status": "available", "vxnets": vxnet})
	if err != nil {
		return nil, err
	}

	if len(nics) > 0 {
		rand.Seed(time.Now().UnixNano())
		nic := nics[rand.Intn(len(nics))]
		if _, err := d.api.AttachNics([]string{nic.ID}, util.InstanceID, true); err != nil {
			return nil, err
		}
		return nic, nil
	}

	// Create a network interface and attach it to the instance.
	nics, err = d.api.CreateNics(vxnet, "", 1, nil)
	if err != nil {
		return nil, err
	}
	if _, err := d.api.AttachNics([]string{nics[0].ID}, util.InstanceID, true); err != nil {
		return nil, err
	}
	return nics[0], nil
}

// TODO: if this failed, call DescribeNics API to pick an idle NIC
func (d *driver) findAttachedIdleNic(vxnet string) (*sdktypes.Nic, error) {
	md, err := util.GetInstanceMetedata()
	if err != nil {
		return nil, err
	}

	m, err := util.LinkList()
	if err != nil {
		return nil, err
	}

	var preferedNic *util.VxnetNic
	d.nicLock.Lock()
	defer d.nicLock.Unlock()
	for _, nic := range md.Vxnets {
		if nic.Role == 1 {
			// Role == 1 means the interface is used by the VM
			continue
		}
		if nic.VxnetID != vxnet || m[nic.NicID] == nil {
			continue
		}

		preferedNic = nic
		break
	}

	if preferedNic == nil {
		return nil, errNoAvailableNic
	}
	return &sdktypes.Nic{
		ID:        preferedNic.NicID,
		PrivateIP: preferedNic.PrivateIP,
	}, nil
}
