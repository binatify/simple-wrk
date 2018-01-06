package loader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/binatify/simple-wrk/util"
)

type Client struct {
	requstUrl string

	*http.Client
}

func NewClient(requstUrl string) *Client {
	return &Client{
		requstUrl: requstUrl,

		Client: &http.Client{
			Transport: &http.Transport{
				DisableCompression:    false,
				DisableKeepAlives:     false,
				ResponseHeaderTimeout: time.Second * time.Duration(5),
			},
		},
	}
}

func (c *Client) DoRequest() (respSize int, duration time.Duration) {
	req, err := http.NewRequest(http.MethodGet, c.requstUrl, nil)
	if err != nil {
		fmt.Println("An error occured doing request", err)
		return
	}

	start := time.Now()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println("An error occured doing request", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("An error occured reading body", err)
		return
	}

	duration = time.Since(start)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		respSize = len(body) + int(util.EstimateHttpHeadersSize(resp.Header))
	} else if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusTemporaryRedirect {
		respSize = int(resp.ContentLength) + int(util.EstimateHttpHeadersSize(resp.Header))
	} else {
		fmt.Println("received status code", resp.StatusCode, "from", resp.Header, "content", string(body), req)
	}

	return
}
