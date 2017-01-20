package types

import "fmt"

type ResponseStatus struct {
	Code    int    `json:"ret_code"`
	Message string `json:"message"`
	Action  string `json:"action"`
}

func (r ResponseStatus) Error() string {
	return fmt.Sprintf("%s failed: code=%d, message=%q", r.Action, r.Code, r.Message)
}

var ErrJobTimeout = fmt.Errorf("job timed out")

type EmptyResponse struct {
	ResponseStatus
}
