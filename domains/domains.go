package domains

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
)

const (
	Endpoint = "domains.google.com/nic/update"
	Version  = "0.1"
)

var (
	NoAuth       = errors.New("nohost")
	BadAuth      = errors.New("badauth")
	NotFQDN      = errors.New("notfqdn")
	BadAgent     = errors.New("badagent")
	Abuse        = errors.New("abuse")
	ServerSide   = errors.New("911")
	ConflictA    = errors.New("conflict A")
	ConflictAAAA = errors.New("conflict AAAA")
)

type Options struct {
	Username string
	Password string
	Hostname string
	Address  net.IP
	Offline  bool
}

type DDNS struct {
	url     *url.URL
	options Options
}

func parseResposneErrors(s string) error {
	errs := []error{NoAuth, BadAuth, NotFQDN, BadAgent, Abuse, ServerSide, ConflictA, ConflictAAAA}
	for _, err := range errs {
		if err.Error() == s {
			return err
		}
	}
	return nil
}

func UserAgent() string {
	return fmt.Sprintf("gddns/%v (%v; %v) %v", Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func (g DDNS) BaseURL() string {
	return fmt.Sprintf("https://%v:%v@%v", g.options.Username, g.options.Password, Endpoint)
}

func (g DDNS) EncodeURL() (u string, err error) {
	if g.url, err = url.Parse(g.BaseURL()); err != nil {
		return
	}

	params := url.Values{}
	params.Add("hostname", g.options.Hostname)
	params.Add("myip", g.options.Address.String())
	if g.options.Offline {
		params.Add("offline", "yes")
	}
	g.url.RawQuery = params.Encode()

	return g.url.String(), nil
}

func (g DDNS) Update(opts Options) (response string, err error) {
	g.options = opts
	var (
		body []byte
		req  *http.Request
		resp *http.Response
		url  string
	)

	if req, err = http.NewRequest("GET", "http://httpbin.org/user-agent", nil); err != nil {
		return
	}
	req.Header.Set("User-Agent", UserAgent())

	if url, err = g.EncodeURL(); err != nil {
		return
	}
	if resp, err = http.Get(url); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return "", fmt.Errorf("unexpected HTTP status code %d, want 2xx", resp.StatusCode)
	}
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	if err = parseResposneErrors(string(body)); err != nil {
		return
	}

	return string(body), nil
}
