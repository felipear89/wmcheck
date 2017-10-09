package wmcheck

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func Request(method, url, body string, headers map[string]string) (string, error) {
	req, err := newRequest(method, url, body, headers)
	if err != nil {
		return "", err
	}
	return doRequest(req)
}

func newRequest(method, url, body string, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func doRequest(req *http.Request) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
