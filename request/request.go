package request

import (
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	netURL "net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	cookiemonster "github.com/MercuryEngineering/CookieMonster"
	"github.com/fatih/color"
	"github.com/kr/pretty"
)

// ReqParams 存放参数设置
type ReqParams struct {
	Debug       bool
	HTTPProxy   string
	Socks5Proxy string
	HTTPRefer   string
	HTTPCookie  string
	FakeHeaders map[string]string
	RetryTimes  int
}

var (
	// Params 参数设置
	Params = &ReqParams{
		RetryTimes: 1,
		Debug:      false,
	}
)

// Request base request
func Request(
	method, url string, body io.Reader, headers map[string]string,
) (*http.Response, error) {
	transport := &http.Transport{
		DisableCompression:  true,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	if Params.HTTPProxy != "" {
		var HTTPProxy, err = netURL.Parse(Params.HTTPProxy)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(HTTPProxy)
	}
	if Params.Socks5Proxy != "" {
		dialer, err := proxy.SOCKS5(
			"tcp",
			Params.Socks5Proxy,
			nil,
			&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			},
		)
		if err != nil {
			return nil, err
		}
		transport.Dial = dialer.Dial
	}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range Params.FakeHeaders {
		req.Header.Set(k, v)
		//fmt.Printf("%s:%s\n", k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if _, ok := headers["Referer"]; !ok {
		req.Header.Set("Referer", url)
	}
	if Params.HTTPCookie != "" {
		var cookie string
		var cookies []*http.Cookie
		cookies, err = cookiemonster.ParseString(Params.HTTPCookie)
		if err != nil || len(cookies) == 0 {
			cookie = Params.HTTPCookie
		}
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		if cookies != nil {
			for _, c := range cookies {
				req.AddCookie(c)
			}
		}
	}
	if Params.HTTPRefer != "" {
		req.Header.Set("Referer", Params.HTTPRefer)
	}
	var (
		res          *http.Response
		requestError error
	)
	for i := 0; ; i++ {
		res, requestError = client.Do(req)
		if requestError == nil && res.StatusCode < 400 {
			break
		} else if i+1 >= Params.RetryTimes {
			var err error
			if requestError != nil {
				err = fmt.Errorf("request error: %v", requestError)
			} else {
				err = fmt.Errorf("%s request error: HTTP %d", url, res.StatusCode)
			}
			return nil, err
		}
		time.Sleep(1 * time.Second)
	}
	if Params.Debug {
		blue := color.New(color.FgBlue)
		fmt.Println()
		blue.Printf("URL:         ")
		fmt.Printf("%s\n", url)
		blue.Printf("Method:      ")
		fmt.Printf("%s\n", method)
		blue.Printf("Headers:     ")
		pretty.Printf("%# v\n", req.Header)
		blue.Printf("Status Code: ")
		if res.StatusCode >= 400 {
			color.Red("%d", res.StatusCode)
		} else {
			color.Green("%d", res.StatusCode)
		}
	}
	return res, nil
}

// Get get request
func Get(url, refer string, headers map[string]string) (string, error) {
	if headers == nil {
		headers = map[string]string{}
	}
	if refer != "" {
		headers["Referer"] = refer
	}
	res, err := Request("GET", url, nil, headers)
	if err != nil {
		return "", err
	}
	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
	case "deflate":
		reader = flate.NewReader(res.Body)
	default:
		reader = res.Body
	}
	defer res.Body.Close()
	defer reader.Close()
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Headers return the HTTP Headers of the url
func Headers(url, refer string) (http.Header, error) {
	headers := map[string]string{
		"Referer": refer,
	}
	res, err := Request("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}
	return res.Header, nil
}

// Size get size of the url
func Size(url, refer string) (int64, error) {
	h, err := Headers(url, refer)
	if err != nil {
		return 0, err
	}
	s := h.Get("Content-Length")
	size, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

// ContentType get Content-Type of the url
func ContentType(url, refer string) (string, error) {
	h, err := Headers(url, refer)
	if err != nil {
		return "", err
	}
	s := h.Get("Content-Type")
	// handle Content-Type like this: "text/html; charset=utf-8"
	return strings.Split(s, ";")[0], nil
}