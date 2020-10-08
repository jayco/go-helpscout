package helpscout

import (
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// ErrorInterrupted ..
var ErrorInterrupted = errors.New("")

// helpscoutAPIEndpoint ..
const helpscoutAPIEndpoint = "https://api.helpscout.net/v2"

// Page ..
type Page struct {
	Size          int `json:"size"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
	Number        int `json:"number"`
}

type generalListAPICallReq struct {
	Embedded interface{} `json:"_embedded"`
	Page     Page        `json:"page"`
}

// Client ..
type Client struct {
	httpClient *httpClient
	auth       *auth
}

// NewClient ..
func NewClient(appID string, appKey string) *Client {
	httpClient := newHTTPClient()

	return &Client{
		httpClient: httpClient,
		auth:       newAuth(httpClient, appID, appKey),
	}
}

// AuthKey ..
func (c *Client) AuthKey(forceUpdate bool) (string, error) {
	token, err := c.auth.getToken(forceUpdate)
	if err != nil {
		return "", errors.Wrap(err, "Unable to update Auth Token")
	}

	return token, nil
}

// SetAuthKey ..
func (c *Client) SetAuthKey(key string, expTime time.Time) {
	c.auth.token = key
	c.auth.tokenExpireTime = expTime
}

// doAPICall ..
func (c *Client) doAPICall(method string, resource string, query *url.Values,
	reqData interface{}, respData interface{}) error {

	repeatAllCnt := 0
	forceTokenUpdate := false
	for {
		token, err := c.auth.getToken(forceTokenUpdate)
		if err != nil {
			return errors.Wrap(err, "Unable to update Auth Token")
		}

		url := helpscoutAPIEndpoint + resource

		authHeader := make(map[string]string)
		authHeader["Authorization"] = fmt.Sprintf("Bearer %s", token)

		repeatCnt := 0
		for {
			err := c.httpClient.doRequest(url, method, authHeader, query, reqData, respData)
			if err == ErrorRateLimit {
				time.Sleep(time.Second)
				repeatCnt++
				if repeatCnt > 10 {
					return errors.New("Unable to submit a request (rate-limit)")
				}

				continue
			}

			if err == ErrorUnauthorized {
				break
			}

			return err
		}

		forceTokenUpdate = true
		repeatAllCnt++
		if repeatAllCnt > 3 {
			return errors.New("Unable to submit a request (authorization failed)")
		}
	}
}
