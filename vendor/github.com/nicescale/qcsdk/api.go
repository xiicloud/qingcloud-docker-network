package qcsdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	EndPoint              = "https://api.qingcloud.com/iaas/"
	Zones                 = "sh1a"
	DefaultJobWaitTimeout = 30
)

type Params map[string]string

func (params Params) Keys() (ret []string) {
	for k, _ := range params {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return
}

func (params Params) String() string {
	vals := []string{}
	for _, k := range params.Keys() {
		vals = append(vals, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	return strings.Join(vals, "&")
}

func (params Params) AddIndexedParams(prefix string, arr []string) {
	for i, el := range arr {
		params[fmt.Sprintf("%s.%d", prefix, i)] = el
	}
}

// add CSV values to params
func (params Params) AddIndexedFilter(prefix string, vals string) {
	params.AddIndexedParams(prefix, strings.Split(vals, ","))
}

func (params Params) AddParam(key string, val interface{}) {
	if v := fmt.Sprint(val); v != "" {
		params[key] = v
	}
}

type Request struct {
	Params
	Action string
}

type Api struct {
	Ak     string
	Sk     string
	Zone   string
	Debug  bool
	client *http.Client
}

func NewApi(ak, sk, zone string) *Api {
	return &Api{
		Ak:   ak,
		Sk:   sk,
		Zone: zone,
		client: &http.Client{
			Timeout: time.Second * 50,
		},
	}
}

func (api *Api) SetDebug(dbg bool) {
	api.Debug = dbg
}

func (api *Api) debug(fmt string, args ...interface{}) {
	if api.Debug {
		log.Printf(fmt, args...)
	}
}

func (api *Api) NewRequest(action string) *Request {
	req := &Request{
		Params: api.commonParams(action),
		Action: action,
	}
	return req
}

func (api *Api) commonParams(action string) Params {
	rand.Seed(time.Now().UnixNano())
	params := make(Params)
	params["time_stamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	params["action"] = action
	params["access_key_id"] = api.Ak
	params["zone"] = api.Zone
	params["version"] = "1"
	params["signature_version"] = "1"
	params["signature_method"] = "HmacSHA256"
	return params
}

func (api *Api) sign(params Params) {
	signParams := make([]string, len(params))
	for i, k := range params.Keys() {
		signParams[i] = url.QueryEscape(k) + "=" + url.QueryEscape(params[k])
	}
	strToSign := "GET\n/iaas/\n" + strings.Join(signParams, "&")
	mac := hmac.New(sha256.New, []byte(api.Sk))
	mac.Write([]byte(strToSign))
	params["signature"] = base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// SendRequest sends the http request to qingcloud API endpoint and parses the response.
// out must be a pointer to a struct that embeds a types.ResponseStatus struct.
func (api *Api) SendRequest(req *Request, out interface{}) error {
	api.sign(req.Params)
	url := EndPoint + "?" + req.String()
	resp, err := api.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		api.debug("failed to parse json. url: %s, error: %q", url, err.Error())
		return err
	}

	status := statusFromResponse(out)
	if status.Code != 0 {
		status.Action = req.Action
		api.debug("req=%s. Invalid ret_code %d. message: %q", url, status.Code, status.Message)
		return status
	}

	return nil
}
