package request

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const agent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/603.3.8 (KHTML, like Gecko) Version/10.1.2 Safari/603.3.8"

// Request is a base function for sending HTTP requests.
func Request(client *http.Client, method string, url string, header http.Header) (*http.Client, *http.Response, http.Header, error) {
	if client == nil {
		// Make a new http client with cookie jar if no existing client is provided
		jar, _ := cookiejar.New(nil)
		client = &http.Client{Timeout: 30 * time.Second, Jar: jar}
	}
	// Build a new request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		// logrus.Errorln(err)
		return nil, nil, nil, err
	}

	// Set headers
	if header != nil {
		if _, ok := header["User-Agent"]; !ok {
			header.Set("User-Agent", agent)
		}
		req.Header = header
	}

	// Do request
	// logrus.Debugf("Sending %v request to %v with headers %v\n", req.Method, req.URL, req.Header)
	res, err := client.Do(req)
	if err != nil {
		// logrus.Errorln(err)
		return nil, nil, nil, err
	}
	return client, res, req.Header, nil
}

// Get wraps the Request function, sends a HTTP GET request, returns the smae client and the html of the content body.
func Get(client *http.Client, url string, headers map[string]string) (*http.Client, string, error) {
	header := http.Header{}
	for k, v := range headers {
		header.Set(k, v)
	}
	client, res, _, err := Request(client, "GET", url, header)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, "", errors.New(http.StatusText(res.StatusCode))
	}
	content, err := ioutil.ReadAll(res.Body)
	return client, string(content), err
}
