gowebshareproxy
==

[![GoDoc](https://godoc.org/github.com/j7mbo/gowebshareproxy?status.svg)](https://godoc.org/github.com/j7mbo/gowebshareproxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE.md)

Simple SDK to make requests through [proxy.webshare.io](https://proxy.webshare.io).

Make proxied requests by specifying the proxy you will go through or choose a random proxy if given a list.

### Basic Usage

1. Sign up for free [here](https://proxy.webshare.io/register/?referral_code=z3bbc9g2zzei).
2. Go to [proxy > list](https://proxy.webshare.io/proxy/list?type=username) and make a note of a proxy address, username 
and password. You can also download the proxy list as a colon delimited text file if you want to do some stuff 
programatically.
3. Do the following (add your own error checking):

```go
package main

import (
    "net/http"
    "net/url"
    
    "github.com/j7mbo/gowebshareproxy"
)

func main() {
    // Input your proxy configuration here.
    proxyUser := "<YOUR-PROXY-USERNAME>"
    proxyPass := "<YOUR-PROXY-PASSWORD>"
    proxyUri, _ := url.Parse("<YOUR-PROXY-HOST>:<YOUR-PROXY-PORT>")
    
    // Url to make proxied request to.
    uri, _ := url.Parse("http://httpbin.org/get")
    
    // Create your request as normal.
    request, _ := http.NewRequest("GET", uri.String(), nil)
    
    // Initialise gowebshareproxy.
    proxy := gowebshareproxy.New(&http.Client{})

    // Make the proxied request.
    res, err := proxy.Request(request, proxyUri, proxyUser, proxyPass)
}
```

### Advanced Usage

You can use `Proxy.RequestWithRandomProxy` to choose a random proxy to perform a request through. From this call you
get back the response, any error as well as the chosen proxy url so you know which proxy you went through. 

On the [proxy > list](https://proxy.webshare.io/proxy/list?type=username) page you will find a button where you can
download your proxy list. It will have the following format:

```text
x.x.x.x:80:username-123:password1
x.x.x.x:80:username-234:password2
x.x.x.x:80:username-345:password3
```

Read in and parse this file into an `[]string`, and then use `Proxy.RequestWithRandomProxy`:

```go
file, _ := os.Open(path)
defer file.Close()

var lines []string

scanner := bufio.NewScanner(file)

for scanner.Scan() {
    lines = append(lines, scanner.Text())
}

proxy := proxy.NewWithList(&http.Client{}, lines)

request, _ := http.NewRequest("GET", "http://httpbin.org/get", nil)

chosenProxy, res, err := proxy.RequestWithRandomProxy(request)
```

Note that if you try and call `Proxy.RequestWithRandomProxy` without having initialised via `NewWithList` you will get
an error.