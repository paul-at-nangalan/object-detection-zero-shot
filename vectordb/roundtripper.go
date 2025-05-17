package vectordb

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ProxyRoundtripper struct {
	trueRoundtripper http.RoundTripper
}

func NewRoundTripper() http.RoundTripper {
	return &ProxyRoundtripper{trueRoundtripper: http.DefaultTransport}
}

var ErrTooManyRequests = errors.New("Too many requests")

type ErrFromHost struct {
	msg  string
	code int
}

func (e ErrFromHost) Error() string {
	return fmt.Sprintf("ERROR: code: %d %s", e.code, e.msg)
}

func (p *ProxyRoundtripper) RoundTrip(request *http.Request) (*http.Response, error) {
	resp, err := p.trueRoundtripper.RoundTrip(request)
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrTooManyRequests
	}

	if resp.StatusCode > 299 {
		errStatuscode := ErrFromHost{
			code: resp.StatusCode,
		}
		if resp.Body != nil {
			msg, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Unable to read response body ", err)
			}
			errStatuscode.msg = string(msg)
		}
		return resp, errStatuscode
	}
	return resp, err
}
