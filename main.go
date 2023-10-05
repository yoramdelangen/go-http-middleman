package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const oneMB = 1024 * 1024
const oneGB = 1024 * oneMB
const responseSize = 2 * oneGB

func main() {
	http.HandleFunc("/", HandleProxyRequest)

	fmt.Println("Starting server on :3333")
	log.Fatalln(http.ListenAndServe(":3333", nil))
}

type RequestHeader = map[string]interface{}

// Handles the incoming request
func HandleProxyRequest(res http.ResponseWriter, req *http.Request) {

	host := req.Header.Get("referrer")
	// get current request query string and pass it on
	// to the client request.
	q := req.URL.Query()
	for k, v := range q {
		key := strings.ToLower(k)
		if !strings.HasPrefix(key, "x-"){
			continue
		}

		if key == "x-host" {
			host = v[0]
		}


		q.Del(k)
	}

	iptvurl, _ := url.Parse(host)
	iptvurl.Path = req.URL.Path
	// iptvurl.RawPath = req.URL.RawPath
	iptvurl.RawQuery = q.Encode()

	// create a header of request header
	headers := RequestHeader{}
	for header, value := range req.Header {
		if strings.HasPrefix(strings.ToLower(header), "x-iptv-") {
			continue
		}
		headers[header] = value
	}

	iptvurl.RawPath = req.URL.RawPath

	r, _ := http.NewRequest(http.MethodGet, iptvurl.String(), nil)
	for header := range req.Header {
		if strings.HasPrefix(strings.ToLower(header), "x-iptv-") {
			continue
		}
		r.Header.Add(header, req.Header.Get(header))
	}

	fmt.Printf("req: %+v\n", iptvurl)
	fmt.Printf("query: %+v\n", q.Encode())

	c := http.Client{}
	resp, err := c.Do(r)

	if err != nil {
		log.Fatalf("Error on request: %+v", err)
	}

	bytesRead := 0
	buf := make([]byte, oneMB*4)
	res.Header().Add("Content-Type", resp.Header.Get("Content-Type"))

	// copy all headers back to the response
	for header := range resp.Header {
		res.Header().Add(header, resp.Header.Get(header))
	}

	defer resp.Body.Close()

	// Read the response body
	for {
		n, err := resp.Body.Read(buf)
		bytesRead += n

		// write to the response
		res.Write(buf[:n])

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal("Error reading HTTP response: ", err.Error())
		}
	}
}
