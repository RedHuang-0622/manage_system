//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	proxyBase := "http://127.0.0.1:5173/api/v1"

	// Login
	loginReq, _ := http.NewRequest("POST", proxyBase+"/auth/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(loginReq)
	if err != nil || resp == nil {
		fmt.Printf("FAIL: Login error: %v\n", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	token := extractToken(string(body))
	if token == "" {
		fmt.Println("FAIL: Login failed")
		return
	}
	fmt.Println("Login OK")

	transport := &http.Transport{
		DisableKeepAlives:  false,
		MaxConnsPerHost:    6,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	endpoints := []string{
		"/borrows?page=1&page_size=10",
		"/users?page=1&page_size=10&status=-1",
		"/equipments?page=1&page_size=12",
		"/borrows/pending?page=1&page_size=10",
		"/borrows/my?page=1&page_size=10",
		"/roles",
		"/users/3",
	}

	pass, fail := 0, 0
	for i := 0; i < 20; i++ {
		ep := endpoints[i%len(endpoints)]
		start := time.Now()
		resp, reqErr := doReqAuth(client, "GET", proxyBase+ep, token)
		elapsed := time.Since(start)
		if reqErr == nil && resp != nil && resp.StatusCode == 200 {
			pass++
			fmt.Printf("  [%2d] PASS %-35s (%3dms)\n", i+1, ep, elapsed.Milliseconds())
		} else {
			fail++
			sc := 0
			errMsg := ""
			if resp != nil {
				sc = resp.StatusCode
			}
			if reqErr != nil {
				errMsg = reqErr.Error()
			}
			fmt.Printf("  [%2d] FAIL %-35s status=%d err=%s\n", i+1, ep, sc, errMsg)
		}
		time.Sleep(time.Duration(2000+i*300) * time.Millisecond)
	}

	fmt.Printf("\n=== %d/%d passed ===\n", pass, pass+fail)
	if fail == 0 {
		fmt.Println("ALL GOOD!")
	}
}

func doReq(client *http.Client, method, url, body, auth string) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp, nil
}

func doReqAuth(client *http.Client, method, url, auth string) (*http.Response, error) {
	return doReq(client, method, url, "", auth)
}

func extractToken(body string) string {
	idx := strings.Index(body, `"token":"`)
	if idx < 0 {
		return ""
	}
	start := idx + 9
	end := strings.Index(body[start:], `"`)
	if end < 0 {
		return ""
	}
	return body[start : start+end]
}
