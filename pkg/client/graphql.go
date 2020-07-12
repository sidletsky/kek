package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type GraphQL struct {
	Client  *http.Client
	Headers map[string]string
}

func NewGraphQL() *GraphQL {
	client := &http.Client{Timeout: 5 * time.Second}
	return &GraphQL{Client: client}
}

type GraphQLQuery struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
}

func (g *GraphQL) Query(endpoint string, q GraphQLQuery) (response []byte, err error) {
	req, err := prepareRequest(endpoint, q, g.Headers)
	if err != nil {
		return []byte{}, err
	}
	res, err := g.Client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func prepareRequest(endpoint string, q GraphQLQuery, headers map[string]string) (*http.Request, error) {
	b, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}
