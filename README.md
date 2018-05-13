# req 

> A http request library for Go.

[![ReportCard][reportcard-image]][reportcard-url] [![GoDoc][godoc-image]][godoc-url] [![License][license-image]][license-url]

## Get

```
go get -u -v github.com/LyricTian/req
```

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/LyricTian/req"
)

func main() {
	req.SetOptions(
		req.SetBaseURL("http://localhost:8080/api"),
	)

	resp, err := req.Get(context.Background(), "/foo", nil)
	if err != nil {
		panic(err)
	}

	body, err := resp.String()
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
}
```


## MIT License

    Copyright (c) 2018 Lyric

[reportcard-url]: https://goreportcard.com/report/github.com/LyricTian/req
[reportcard-image]: https://goreportcard.com/badge/github.com/LyricTian/req
[godoc-url]: https://godoc.org/github.com/LyricTian/req
[godoc-image]: https://godoc.org/github.com/LyricTian/req?status.svg
[license-url]: http://opensource.org/licenses/MIT
[license-image]: https://img.shields.io/npm/l/express.svg
