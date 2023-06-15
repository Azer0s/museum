package util

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func PrintRequest(req *http.Request) {
	fmt.Println("Method: " + req.Method)
	fmt.Println("URL: " + req.URL.String())
	fmt.Println("Proto: " + req.Proto)
	fmt.Println("Host: " + req.Host)
	fmt.Println("RemoteAddr: " + req.RemoteAddr)
	fmt.Println("RequestURI: " + req.RequestURI)
	fmt.Println("Header: ")
	for k, v := range req.Header {
		fmt.Println("\t" + k + ": " + strings.Join(v, ", "))
	}
}

func PrintResponse(res *http.Response) {
	fmt.Println("Status: " + res.Status)
	fmt.Println("StatusCode: " + strconv.Itoa(res.StatusCode))
	fmt.Println("Host: " + res.Request.Host)
	fmt.Println("Proto: " + res.Proto)
	fmt.Println("RequestURI: " + res.Request.RequestURI)
	fmt.Println("Header: ")
	for k, v := range res.Header {
		fmt.Println("\t" + k + ": " + strings.Join(v, ", "))
	}
}
