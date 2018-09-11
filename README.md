# req 

> A http request library for Go.

[![Build][Build-Status-Image]][Build-Status-Url] [![Coverage][Coverage-Image]][Coverage-Url] [![ReportCard][reportcard-image]][reportcard-url] [![GoDoc][godoc-image]][godoc-url] [![License][license-image]][license-url]

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
		req.SetBaseURL("https://jsonplaceholder.typicode.com"),
	)

	resp, err := req.Get(context.Background(), "/posts/42", nil)
	if err != nil {
		panic(err)
	}

	body, err := resp.String()
	if err != nil {
		panic(err)
	}
	fmt.Printf("status:%d,body:\n%s\n", resp.Response().StatusCode, body)
}
```

> output:

```
status:200,body:
{
  "userId": 5,
  "id": 42,
  "title": "commodi ullam sint et excepturi error explicabo praesentium voluptas",
  "body": "odio fugit voluptatum ducimus earum autem est incidunt voluptatem\nodit reiciendis aliquam sunt sequi nulla dolorem\nnon facere repellendus voluptates quia\nratione harum vitae ut"
}
```


## MIT License

    Copyright (c) 2018 Lyric

[Build-Status-Url]: https://travis-ci.org/LyricTian/req
[Build-Status-Image]: https://travis-ci.org/LyricTian/req.svg?branch=master
[Coverage-Url]: https://coveralls.io/github/LyricTian/req?branch=master
[Coverage-Image]: https://coveralls.io/repos/github/LyricTian/req/badge.svg?branch=master
[reportcard-url]: https://goreportcard.com/report/github.com/LyricTian/req
[reportcard-image]: https://goreportcard.com/badge/github.com/LyricTian/req
[godoc-url]: https://godoc.org/github.com/LyricTian/req
[godoc-image]: https://godoc.org/github.com/LyricTian/req?status.svg
[license-url]: http://opensource.org/licenses/MIT
[license-image]: https://img.shields.io/npm/l/express.svg
