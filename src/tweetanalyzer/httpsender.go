package tweetanalyzer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strings"
)

var debug = false

type HttpSender struct {
	Client     *http.Client
	csrf_token string
}

func NewHttpSender() *HttpSender {
	jar, _ := cookiejar.New(nil)
	t := HttpSender{Client: &http.Client{Jar: jar}}
	return &t
}

func (t *HttpSender) Send(urlstr string, send_post_data bool, post_data url.Values) (string, string, error) {
	var req *http.Request
	var err error

	if send_post_data == false {
		req, err = http.NewRequest("GET", urlstr, nil)
		if err != nil {
			return "", "", fmt.Errorf("Get request failed: %s", err)
		}
	} else {
		req, err = http.NewRequest("POST", urlstr, strings.NewReader(post_data.Encode()))
		if err != nil {
			return "", "", fmt.Errorf("Post request failed: %s", err)
		}
	}

	if len(t.csrf_token) > 0 {
		req.Header.Set("X-CSRFToken", t.csrf_token)
		req.Header.Set("X-Instagram-AJAX", "1")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Accept", "*/*")
	} else {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	}

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	//req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9, * /*;q=0.8")
	//req.Header.Set("Accept-Encoding", "gzip,deflate,sdch")
	//req.Header.Set("Accept-Language", "en-US,en;q=0.8")

	req.Header.Set("Host", "instagram.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Host", "twitter.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:16.0) Gecko/20100101 Firefox/16.0")
	//req.Header.Set("Origin", "https://twitter.com")
	req.Header.Set("Referer", "https://instagram.com")
	//req.Header.Set("Cache-Control", "max-age=0")

	if debug {
		dump, err := httputil.DumpRequest(req, true)
		if err == nil {
			fmt.Println("request header: " + string(dump) + "\n")
		}
	}

	resp, err := t.Client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("Http request failed: %s", err)
	}

	defer resp.Body.Close()

	// should be: redirect_url := resp.Request.URL.String()
	redirect_url, _ := url.QueryUnescape(resp.Request.URL.String())

	// Read HTML body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("Read HTML body failed: %s", err)
	}
	str := string(b)

	if debug {
		fmt.Println("StatusCode:", resp.StatusCode)
		// print cookies
		fmt.Println("cookies:")
		for _, c := range resp.Cookies() {
			fmt.Println(c)
		}
	}

	return str, redirect_url, nil
}

func (t *HttpSender) GetData(s string, start_str string, end_str string) (string, error) {
	var data string

	i_start := strings.Index(s, start_str)
	if i_start == -1 {
		return "", fmt.Errorf("start string not found")
	}

	s_new := s[i_start+len(start_str):]

	i_end := strings.Index(s_new, end_str)
	if i_end == -1 {
		return "", fmt.Errorf("end string not found")
	}

	data = s[i_start+len(start_str) : i_start+len(start_str)+i_end]

	return data, nil
}
