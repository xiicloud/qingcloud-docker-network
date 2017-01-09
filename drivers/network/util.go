package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/nicescale/qingcloud-docker-network/util"
)

var (
	httpClient        = &http.Client{Timeout: time.Second * 5}
	errNoAvailableNic = fmt.Errorf("no available nic")
)

func (d *driver) writeFile(name string, data interface{}) error {
	tmpPath := filepath.Join(d.root, "tmp")
	if err := os.MkdirAll(tmpPath, 0700); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}

	tmp, err := ioutil.TempFile(tmpPath, "")
	if err != nil {
		return err
	}
	err = json.NewEncoder(tmp).Encode(data)
	if err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	tmp.Sync()
	tmp.Close()
	return os.Rename(tmp.Name(), name)
}

func (d *driver) configDir() string {
	return filepath.Join(d.root, "networks")
}

func (d *driver) netConfigDir(networkID string) string {
	return filepath.Join(d.configDir(), networkID)
}

func (d *driver) netConfigPath(networkID string) string {
	return filepath.Join(d.netConfigDir(networkID), networkID+".json")
}

func (d *driver) epConfigDir(networkID string) string {
	return filepath.Join(d.netConfigDir(networkID), "endpoints")
}

func (d *driver) epConfigPath(networkID, endpointID string) string {
	return filepath.Join(d.epConfigDir(networkID), endpointID+".json")
}

func (d *driver) loadNetworks() error {
	cfgDir := d.configDir()
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		return err
	}

	networks, err := filepath.Glob(filepath.Join(d.configDir(), "*", "*.json"))
	if err != nil {
		return err
	}
	for _, f := range networks {
		n := &netConfig{}
		if err := util.ReadJSON(f, n); err != nil {
			return err
		}

		if n.endpoints == nil {
			n.endpoints = make(map[string]*endpoint)
		}

		eps, err := filepath.Glob(filepath.Join(d.epConfigDir(n.ID), "*", "*.json"))
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		for _, e := range eps {
			ep := &endpoint{}
			if err := util.ReadJSON(e, ep); err != nil {
				return err
			}
			n.endpoints[ep.ID] = ep
		}

		d.networks[n.ID] = n
	}

	return nil
}

func (d *driver) getNetwork(nid string) *netConfig {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.networks[nid]
}

// TODO: if this failed, call DescribeNics API to pick an idle NIC
func (d *driver) findAvailableNic(epid, vxnet, ip string) (*endpoint, error) {
	md, err := util.GetInstanceMetedata()
	if err != nil {
		return nil, err
	}

	m, err := util.LinkList()
	if err != nil {
		return nil, err
	}

	var preferedNic *util.VxnetNic
	d.mu.Lock()
	for _, nic := range md.Vxnets {
		if nic.Role == 1 {
			// Role == 1 means the interface is used by the VM
			continue
		}
		if nic.VxnetID != vxnet || m[nic.NicID] == nil {
			continue
		}
		if d.lockedNics[nic.NicID] {
			continue
		}

		if util.IPEquals(ip, nic.PrivateIP.String()) {
			preferedNic = nic
			break
		}
	}

	if preferedNic == nil {
		d.mu.Unlock()
		return nil, errNoAvailableNic
	}
	d.lockedNics[preferedNic.NicID] = true
	d.mu.Unlock()
	defer d.unlockNic(preferedNic.NicID)

	nicName := genNicName(epid)
	if err = util.RenameLink(m[preferedNic.NicID], nicName); err != nil {
		return nil, err
	}

	ep := &endpoint{
		ID:    epid,
		NicID: preferedNic.NicID,
		IP:    ip,
	}
	return ep, nil
}

func (d *driver) lockNic(id string) {
	d.mu.Lock()
	d.lockedNics[id] = true
	d.mu.Unlock()
}

func (d *driver) unlockNic(id string) {
	d.mu.Lock()
	delete(d.lockedNics, id)
	d.mu.Unlock()
}

func (d *driver) saveEndpoint(nid string, ep *endpoint) error {
	if err := os.MkdirAll(d.epConfigDir(nid), 0700); err != nil {
		return err
	}

	return d.writeFile(d.epConfigPath(nid, ep.ID), ep)
}

func genNicName(epid string) string {
	return epid[:12]
}
