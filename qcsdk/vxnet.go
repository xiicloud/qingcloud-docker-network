package qcsdk

import (
	"github.com/nicescale/qingcloud-docker-network/qcsdk/types"
)

func (api *Api) DescribeVxnets(filters ...Params) ([]*types.Vxnet, error) {
	req := api.NewRequest("DescribeVxnets")
	mergeFilterParams(req, []string{"tags", "vxnets"}, filters)

	ret := types.DescribeVxnetsResponse{}
	err := api.SendRequest(req, &ret)
	if err != nil {
		return nil, err
	}

	return ret.Vxnets, nil
}
