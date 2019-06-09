package gowebshareproxy

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// authorizationHeader The HTTP Proxy-Authorization request header contains the credentials to authenticate a user agent
// to a proxy server, usually after the server has responded with a 407 Proxy Authentication Required status and the
// Proxy-Authenticate header. See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Proxy-Authorization.
const authorizationHeader = "Proxy-Authorization"

// proxyStringFormat is the format of the file provided by gowebshareproxy.
const proxyStringFormat = "<HOST>:<PORT>:<USER>:<PASS>"

// expectedProxyStringSplitCount is for splitting the proxy string from the given file: this should be the result count.
const expectedProxyStringSplitCount = 4

// Proxy is used to perform a proxied request via gowebshareproxy.
type Proxy interface {
	// Request decorates http.Client.Do() by adding the given proxy configuration to the request.
	Request(request http.Request, proxyURL url.URL, proxyUser string, proxyPassword string) (*http.Response, error)
	// RequestWithRandomProxy chooses a random proxy to make the given request through and returns both the response /
	// error combination and the chosen proxy.
	RequestWithRandomProxy(request http.Request) (proxyURL *url.URL, response *http.Response, err error)
}

// proxy is used to perform a proxied request.
type proxy struct {
	client    http.Client
	proxyList []string
}

// NewProxy returns a new Proxy with an (optionally pre-configured) http client.
func New(client http.Client) Proxy {
	return &proxy{client: client, proxyList: make([]string, 0)}
}

// NewWithList returns a new Proxy with an (optionally pre-configured) http client and a list of proxies, which you can
// download directly from the [webshare proxy list page](https://proxy.webshare.io/proxy/list). You can then call
// Proxy.RequestWithRandomProxy() to have a proxy randomly chosen from this list for the given request.
func NewWithList(client http.Client, proxyList []string) Proxy {
	return &proxy{client: client, proxyList: proxyList}
}

// Request decorates http.Client.Do() by adding the given proxy configuration to the request.
func (p *proxy) Request(
	request http.Request, proxyURL url.URL, proxyUser string, proxyPassword string,
) (*http.Response, error) {
	header := http.Header{}

	auth := fmt.Sprintf("%s:%s", proxyUser, proxyPassword)
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	header.Add(authorizationHeader, basicAuth)

	p.client.Transport = &http.Transport{
		Proxy:              http.ProxyURL(&proxyURL),
		ProxyConnectHeader: header,
	}

	return p.client.Do(&request)
}

// RequestWithRandomProxy chooses a random proxy to make the given request through and returns both the response /
// error combination and the chosen proxy.
func (p *proxy) RequestWithRandomProxy(request http.Request) (proxyURL *url.URL, response *http.Response, err error) {
	if len(p.proxyList) == 0 {
		return nil, nil, errors.New(
			"cannot call RequestWithRandomProxy with no proxies provided on initialisation - ensure that you have " +
				"initialised with gowebshareproxy.NewWithList() and provided a proxy list as the README shows",
		)
	}

	index := p.chooseRandomIndex()
	proxyString := p.proxyList[index]

	host, port, username, password, err := p.parseProxyString(proxyString)
	if err != nil {
		return nil, nil, err
	}

	// The // are needed because of this joke: https://github.com/golang/go/issues/19297.
	proxyURL, err = url.Parse(fmt.Sprintf("//%s:%s", host, port))
	if err != nil {
		return nil, nil, errors.New(
			fmt.Sprintf("unable to parse host / port into a valid uri, error: '%s'", err.Error()),
		)
	}

	response, err = p.Request(request, *proxyURL, username, password)

	return proxyURL, response, err
}

// chooseRandomIndex chooses a random index given p.proxyList.
func (p *proxy) chooseRandomIndex() int {
	source := rand.NewSource(time.Now().Unix())
	random := rand.New(source)

	return random.Intn(len(p.proxyList))
}

// parseProxyString parses a single proxy string line expected to look like expectedProxyStringSplitCount.
func (p *proxy) parseProxyString(proxyString string) (host, port, username, password string, err error) {
	split := strings.Split(proxyString, ":")

	if len(split) != expectedProxyStringSplitCount {
		return "", "", "", "", errors.New(
			fmt.Sprintf(
				"expected a file containing a list of strings in the following format: '%s', but got a different "+
					"format - it looks like the format of the file may have changed or you provided an incorrect file",
				proxyStringFormat,
			),
		)
	}

	return split[0], split[1], split[2], split[3], nil
}
