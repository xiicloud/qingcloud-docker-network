package qcsdk

import (
	"fmt"
	"testing"
)

func TestDescribeInstances(t *testing.T) {
	instances, err := api.DescribeInstances()
	assertNoError(t, err)

	for _, i := range instances {
		fmt.Printf("%-12s\t%-32s\t%s\n", i.ID, i.Image.Name, i.Status)
	}
}
