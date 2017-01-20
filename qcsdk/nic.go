package qcsdk

import (
	"github.com/nicescale/qingcloud-docker-network/qcsdk/types"
)

func (api *Api) DescribeNics(filters ...Params) ([]*types.Nic, error) {
	req := api.NewRequest("DescribeNics")
	mergeFilterParams(req, []string{"nics", "vxnets", "status", "instances"}, filters)

	ret := types.DescribeNicsResponse{}
	err := api.SendRequest(req, &ret)
	if err != nil {
		return nil, err
	}

	return ret.Nics, nil
}

func (api *Api) CreateNics(vxnet, name string, count int, ips []string) ([]*types.Nic, error) {
	req := api.NewRequest("CreateNics")
	req.AddParam("vxnet", vxnet)
	req.AddParam("nic_name", name)
	req.AddParam("count", count)
	req.AddIndexedParams("private_ips", ips)

	ret := types.CreateNicResponse{}
	err := api.SendRequest(req, &ret)
	if err != nil {
		return nil, err
	}

	return ret.Nics, nil
}

// AttachNics attaches the specified nics to the instance.
// Returns the job ID on success.
func (api *Api) AttachNics(nics []string, instanceId string, wait bool) (string, error) {
	req := api.NewRequest("AttachNics")
	req.AddIndexedParams("nics", nics)
	req.AddParam("instance", instanceId)

	ret := types.NicActionResponse{}
	err := api.SendRequest(req, &ret)
	if err != nil {
		return "", err
	}

	if !wait {
		return ret.JobID, nil
	}

	return api.WaitForJobSuccess(ret.JobID)
}

// DetachNics detaches the specified nics from the instance.
// Returns the job ID on success.
func (api *Api) DetachNics(nics []string, wait bool) (string, error) {
	req := api.NewRequest("DetachNics")
	req.AddIndexedParams("nics", nics)

	ret := types.NicActionResponse{}
	err := api.SendRequest(req, &ret)
	if err != nil {
		return "", err
	}

	if !wait {
		return ret.JobID, nil
	}

	return api.WaitForJobSuccess(ret.JobID)
}

func (api *Api) ModifyNicAttributes(id, name, vxnet, ip string) error {
	req := api.NewRequest("ModifyNicAttributes")
	req.AddParam("nic", id)
	req.AddParam("nic_name", name)
	req.AddParam("vxnet", vxnet)
	req.AddParam("private_ip", ip)

	ret := types.EmptyResponse{}
	return api.SendRequest(req, &ret)
}

func (api *Api) DeleteNics(nics []string) error {
	req := api.NewRequest("DeleteNics")
	req.AddIndexedParams("nics", nics)

	ret := types.EmptyResponse{}
	return api.SendRequest(req, &ret)
}
