package http

import (
	"net/http"
	"net/url"
)

func GetHttpClient(proxy string) *http.Client {
	if proxy == "" {
		return http.DefaultClient
	}

	// Parse the proxy URL
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return http.DefaultClient
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
	}

	return client
}
