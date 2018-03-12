package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const HttpUrl = "http://localhost:3000/"

type HttpClient struct {
	Token    string
	Response *http.Response
}

func NewHttpClient(token string) *HttpClient {
	return &HttpClient{Token: token}
}

func (h *HttpClient) GetRequest(path string) []byte {
	return h.PerformRequest(path, "GET", []byte(""))
}

func (h *HttpClient) PostRequest(path string, body []byte) []byte {
	return h.PerformRequest(path, "POST", body)
}

func (h *HttpClient) PutRequest(path string, body []byte) []byte {
	return h.PerformRequest(path, "PUT", body)
}

func (h *HttpClient) PerformRequest(path string, verb string, body []byte) []byte {

	url := HttpUrl + path
	client := &http.Client{}
	req, _ := http.NewRequest(verb, url, bytes.NewBuffer(body))
	fmt.Println("Verb is ", verb)
	fmt.Println("url is ", url)
	authHeader := fmt.Sprintf("Bearer %s", h.Token)

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var requestError error
	if h.Response, requestError = client.Do(req); requestError != nil {
		fmt.Println("Error contacting server: ", requestError)
		os.Exit(0)
	}

	bodyBytes, err := ioutil.ReadAll(h.Response.Body)
	if err != nil {
		panic(err)
	}
	defer h.Response.Body.Close()

	return bodyBytes
}
