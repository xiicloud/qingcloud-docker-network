package qcsdk

import (
	"fmt"
	"strings"
	"time"

	"github.com/nicescale/qcsdk/types"
)

func (api *Api) DescribeJobs(filters ...Params) ([]*types.Job, error) {
	req := api.NewRequest("DescribeJobs")
	mergeFilterParams(req, []string{"jobs", "status"}, filters)

	ret := types.DescribeJobsResponse{}
	err := api.SendRequest(req, &ret)
	if err != nil {
		return nil, err
	}

	return ret.Jobs, nil
}

func (api *Api) WaitForJob(id, status string, timeout int) (job *types.Job, err error) {
	m := make(map[string]bool)
	for _, s := range strings.Split(status, ",") {
		m[s] = true
	}

	for i := 0; i < timeout*2; i++ {
		jobs, err := api.DescribeJobs(Params{"jobs": id})
		if err != nil {
			return nil, err
		}
		if len(jobs) != 1 {
			continue
		}

		job = jobs[0]
		if m[job.Status] {
			return job, nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return job, types.ErrJobTimeout
}

func (api *Api) WaitForJobSuccess(id string) (string, error) {
	job, err := api.WaitForJob(id, "failed,successful", DefaultJobWaitTimeout)
	if err != nil || job.Status == "successful" {
		return id, err
	}

	return id, fmt.Errorf("job %s failed: %s", id, job.ErrorCodes)
}
