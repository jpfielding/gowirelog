# gowirelog
Simple wire logging utils targeted to http.Client


```go
package main

import (
  "net/http"

  "github.com/jpfielding/gowirelog/wirelog"
}

func main() {
	transport := wirelog.NewHTTPTransport()
	logname := "/tmp/logs/http.log"
	wirelog.LogToFile(transport, logname, true, true)
	client := &http.Client{
		Transport: transport,
	}
	client.Jar, _ = cookiejar.New(nil)
	login, _ := url.Parse(login)
	req, _ := http.NewRequest("GET", login.String(), nil)
	resp, _ := ctxhttp.Do(ctx, client, req1)
	defer resp.Body.Close()
}

```
