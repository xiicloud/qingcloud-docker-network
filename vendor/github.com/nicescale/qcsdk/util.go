package qcsdk

import (
	"fmt"
	"reflect"

	"github.com/nicescale/qcsdk/types"
)

var (
	errUnknown = types.ResponseStatus{Code: -1, Message: "unknown status code"}
	statusType = reflect.TypeOf(types.ResponseStatus{})
)

func statusFromResponse(data interface{}) types.ResponseStatus {
	d := reflect.Indirect(reflect.ValueOf(data))
	if d.Kind() != reflect.Struct {
		return errUnknown
	}

	status := reflect.Indirect(d.FieldByName("ResponseStatus"))
	if status.Type() != statusType {
		// panic here to ensure all API response types embed the ResponseStatus struct
		err := fmt.Sprintf("response type %s must embed a ResponseStatus struct", d.Kind())
		panic(err)
	}

	code := status.FieldByName("Code").Int()
	message := status.FieldByName("Message").String()
	return types.ResponseStatus{
		Code:    int(code),
		Message: message,
	}
}

func flattenFilters(filters []Params) Params {
	filter := make(Params)
	for _, f := range filters {
		for k, v := range f {
			filter[k] = v
		}
	}
	return filter
}

func mergeFilterParams(req *Request, idxParams []string, filters []Params) {
	if len(filters) == 0 {
		return
	}

	filter := flattenFilters(filters)

	m := make(map[string]bool, len(idxParams))
	for _, p := range idxParams {
		m[p] = true
	}

	for k, v := range filter {
		if m[k] {
			req.AddIndexedFilter(k, v)
		} else {
			req.AddParam(k, v)
		}
	}
}
