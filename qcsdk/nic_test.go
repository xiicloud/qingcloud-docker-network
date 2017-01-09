package qcsdk

import (
	"fmt"
	"testing"
)

func TestDescribeNics(t *testing.T) {
	nics, err := api.DescribeNics(Params{"status": "available"})
	assertNoError(t, err)

	for _, i := range nics {
		fmt.Printf("%-20s\t%-32s\t%s\n", i.ID, i.Status, i.PrivateIP)
	}
}

func TestCreateDeleteNics(t *testing.T) {
	ip := "172.25.1.33"
	nics, err := api.CreateNics("vxnet-qpxj8ci", "xxx", 1, []string{ip})
	assertNoError(t, err)
	assertStringEqual(t, ip, nics[0].PrivateIP.String())
	for _, i := range nics {
		fmt.Printf("%-20s\t%-32s\t%s\n", i.ID, i.Status, i.PrivateIP)
	}

	err = api.DeleteNics([]string{nics[0].ID})
	assertNoError(t, err)
}
