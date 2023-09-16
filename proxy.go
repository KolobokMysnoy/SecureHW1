package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type SaveFunc func(Response, Request) error

type Proxy interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	SaveReqAndResp(SaveFunc)
}

type ProxyHTTP struct {
	save SaveFunc
}

func (p *ProxyHTTP) SaveReqAndResp(addFunc SaveFunc) {
	p.save = addFunc
}

func (p *ProxyHTTP) getClientReply(req *http.Request, wr http.ResponseWriter) (*http.Response, error) {
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Print("ServeHTTP:", err)
		return nil, err
	}

	delHopHeaders(resp.Header)

	return resp, nil
}

func (p ProxyHTTP) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	postParams := make(map[string][]string)

	contentTypeValues := req.Header.Values("CONTENT-TYPE")
	for _, typeOfHead := range contentTypeValues {
		if typeOfHead == "application/x-www-form-urlencoded" {
			err := req.ParseForm()
			if err != nil {
				log.Println("Error parsing form data:", err)
				http.Error(wr, "Internal Server Error",
					http.StatusInternalServerError)
				return
			}

			for key, values := range req.Form {
				for _, value := range values {
					postParams[key] = append(postParams[key], value)
				}
			}
		}
	}

	getParams := make(map[string][]string)

	query := req.URL.Query()
	for key, values := range query {
		for _, value := range values {
			getParams[key] = append(getParams[key], value)
		}
	}

	var cookies []http.Cookie
	for _, cookie := range req.Cookies() {
		cookies = append(cookies, *cookie)
	}

	requestToWork := Request{
		Method:     req.Method,
		Path:       req.URL.Path,
		GetParams:  getParams,
		Headers:    req.Header,
		Cookies:    cookies,
		PostParams: postParams,
	}

	resp, err := p.getClientReply(req, wr)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	response := Response{
		Code:    resp.StatusCode,
		Message: resp.Status,
		Headers: resp.Header,
		Body:    string(htmlBytes),
	}

	err = p.save(response, requestToWork)
	if err != nil {
		log.Print(err)
		return
	}

	log.Println(req.RemoteAddr, " ", resp.Status)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func appendHostToXForwardHeader(header http.Header, host string) {
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func copyHeader(dest, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dest.Add(k, v)
		}
	}
}
