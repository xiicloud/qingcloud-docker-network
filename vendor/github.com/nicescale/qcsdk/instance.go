package qcsdk

import (
	"github.com/nicescale/qcsdk/types"
)

func (api *Api) DescribeInstances(filters ...Params) ([]*types.Instance, error) {
	req := api.NewRequest("DescribeInstances")
	mergeFilterParams(req, []string{"status", "instances"}, filters)
	if _, ok := req.Params["status.0"]; !ok {
		req.AddIndexedParams("status", []string{"pending", "running", "stopped", "suspended"})
	}

	ret := types.DescribeInstancesResponse{}
	err := api.SendRequest(req, &ret)
	if err != nil {
		return nil, err
	}

	return ret.Instances, nil
}
