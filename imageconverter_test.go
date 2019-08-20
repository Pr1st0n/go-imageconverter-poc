package main

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var client = &http.Client{Timeout: time.Second}

func getBodyHash(body []byte) string {
	hash := md5.New()
	hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil))
}

func TestCropResizeRequest(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(imageHandler))
	defer testServer.Close()
	urlParams := url.Values{}

	urlParams.Add("crop", "0x0x500x500")
	urlParams.Add("resize", "250x250")
	urlParams.Add("source", "./files/cat.jpg")

	req, _ := http.NewRequest("GET", testServer.URL+"?"+urlParams.Encode(), nil)
	res, err := client.Do(req)

	if err != nil {
		t.Errorf("Unexpected server error: %v", err)
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	expected := "207c9ea863469333d227b2b740150b1f"
	hash := getBodyHash(body)

	if expected != hash {
		t.Errorf("\nInvalid response MD5 hash\nExpected: %v\nActual  : %v", expected, hash)
	}
}
