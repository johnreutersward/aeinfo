# aeinfo

aeinfo is a Go package that registers an HTTP handler which provides information and statistics for appengine applications.

[![GoDoc](https://godoc.org/github.com/rojters/aeinfo?status.svg)](https://godoc.org/github.com/rojters/aeinfo)

## Usage

Use goapp get to install the package

```
$ goapp get github.com/rojters/aeinfo
```

Import for side effects

```go
package app

import (
	"net/http"

	_ "github.com/rojters/aeinfo"
)

func init() {
	http.HandleFunc("/hello", hello)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}
```

As admin, go to ``/_ah/aeinfo``

```
{
	{
		appID: "myapp",
		datacenter: "us3",
		defaultVersionHostname: "myapp.appspot.com",
		instanceID: "99c61b179c95d1b2e81b224bcfbff97dd6ab90d6",
		isDevAppServer: false,
		moduleName: "default",
		serverSoftware: "Google App Engine/1.9.17",
		versionID: "1.382093188328700785",
		serverTime: "2015-02-08T14:59:25.186210687Z",
		goVersion: "release-branch.go1.4 (appengine-1.9.17)",
	cpu: {
		total: 50,
		rate1M: 0.8333333333333333,
		rate10M: 0.08333333333333334
	},
	ram: {
		current: 4.30859375,
		average1M: 0,
		average10M: 0
	},
	modules: [
		{
			name: "default",
			versions: [
				"1",
				"ah-builtin-datastoreservice",
				"ah-builtin-python-bundle"
			]
		},
		{
		name: "frontend",
			versions: [
				"1"
			]
		}
	],
	memcache: {
		hits: 0,
		misses: 0,
		byteHits: 0,
		items: 0,
		bytes: 0,
		oldest: 0
	},
	taskqueue: {
		name: "default",
		tasks: 0,
		oldestETA: "0001-01-01T00:00:00Z",
		executed1Minute: 0,
		inFlight: 0,
		enforcedRate: 5
	},
	caller: {
		remoteAddr: "189.178.178.12",
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.111 Safari/537.36",
		country: "CC",
		region: "",
		city: "MyCity",
		cityLatLong: "51.328930,12.064910",
		email: "foo@bar.com"
		}
	}
}
```

## License

MIT