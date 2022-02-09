package ihttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//post请求
func Post(urlStr string, data map[string]interface{}, contentType map[string]string) ([]byte, error) {
	d := ""
	if contentType["Content-Type"] == "application/json" {
		bt, _ := jsoniter.Marshal(data)
		d = string(bt)
	} else {
		dataVal := url.Values{}
		for v1, v2 := range data {
			dataVal.Add(v1, fmt.Sprintf("%v", v2))
		}
		d = dataVal.Encode()
	}

	req, err := http.NewRequest(`POST`, urlStr, strings.NewReader(d))
	for k, v := range contentType {
		req.Header.Add(k, v)
	}
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

//get
func Get(url string, contentType map[string]string) (string, error) {
	jsonStr, _ := json.Marshal("")
	req, err := http.NewRequest(`GET`, url, bytes.NewBuffer(jsonStr))
	for k, v := range contentType {
		req.Header.Add(k, v)
	}
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	return string(result), nil
}

func GetContentTypeJson() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}
