// Package aeinfo registers an HTTP handler which provides information and statistics for appengine applications.
package aeinfo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"appengine"
	"appengine/memcache"
	"appengine/module"
	aeruntime "appengine/runtime"
	"appengine/taskqueue"
	"appengine/user"
)

const serveURL = "/_ah/aeinfo/"

func init() {
	http.HandleFunc(serveURL, aeinfoHandler)
}

func aeinfoHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if appengine.IsDevAppServer() {
		// noop
	} else if u := user.Current(c); u == nil {
		if loginURL, err := user.LoginURL(c, r.URL.String()); err == nil {
			http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		} else {
			serveError(w, err)
		}
		return
	} else if !u.Admin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	i, err := Gather(c, r)
	if err != nil {
		serveError(w, err)
	}

	if err = json.NewEncoder(w).Encode(i); err != nil {
		serveError(w, err)
	}
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func Gather(c appengine.Context, r *http.Request) (*Info, error) {
	var ms Memcache
	memstats, err := memcache.Stats(c)
	if err != nil && err != memcache.ErrNoStats {
		return nil, err
	}

	if err == nil {
		ms = Memcache{
			Hits:     memstats.Hits,
			Misses:   memstats.Misses,
			ByteHits: memstats.ByteHits,
			Items:    memstats.Items,
			Bytes:    memstats.Bytes,
			Oldest:   memstats.Oldest,
		}
	}

	taskstats, err := taskqueue.QueueStats(c, []string{"default"}, 0)
	if err != nil || len(taskstats) == 0 {
		return nil, fmt.Errorf("unable to gather taskqueue stats")
	}

	mnames, err := module.List(c)
	if err != nil {
		return nil, err
	}

	modules := make([]Module, 0, len(mnames))
	for _, n := range mnames {
		versions, err := module.Versions(c, n)
		if err != nil {
			return nil, err
		}
		m := Module{
			Name:     n,
			Versions: versions,
		}
		modules = append(modules, m)
	}

	aestats, err := aeruntime.Stats(c)
	if err != nil {
		return nil, err
	}

	i := &Info{
		AppID:                  appengine.AppID(c),
		Datacenter:             appengine.Datacenter(),
		DefaultVersionHostname: appengine.DefaultVersionHostname(c),
		InstanceID:             appengine.InstanceID(),
		IsDevAppServer:         appengine.IsDevAppServer(),
		ModuleName:             appengine.ModuleName(c),
		ServerSoftware:         appengine.ServerSoftware(),
		VersionID:              appengine.VersionID(c),
		ServerTime:             time.Now(),
		GoVersion:              runtime.Version(),
		CPU: CPU{
			Total:   aestats.CPU.Total,
			Rate1M:  aestats.CPU.Rate1M,
			Rate10M: aestats.CPU.Rate10M,
		},
		RAM: RAM{
			Current:    aestats.RAM.Current,
			Average1M:  aestats.RAM.Average1M,
			Average10M: aestats.RAM.Average10M,
		},
		Modules:  modules,
		Memcache: ms,
		Taskqueue: Taskqueue{
			Name:            "default",
			Tasks:           taskstats[0].Tasks,
			OldestETA:       taskstats[0].OldestETA,
			InFlight:        taskstats[0].InFlight,
			Executed1Minute: taskstats[0].Executed1Minute,
			EnforcedRate:    taskstats[0].EnforcedRate,
		},
		Caller: Caller{
			RemoteAddr:  r.RemoteAddr,
			UserAgent:   r.UserAgent(),
			City:        r.Header.Get("X-AppEngine-City"),
			CityLatLong: r.Header.Get("X-AppEngine-CityLatLong"),
			Country:     r.Header.Get("X-AppEngine-Country"),
			Region:      r.Header.Get("X-AppEngine-Region"),
			Email:       user.Current(c).Email,
		},
	}

	return i, nil
}

type Info struct {
	AppID                  string    `json:"appID"`
	Datacenter             string    `json:"datacenter"`
	DefaultVersionHostname string    `json:"defaultVersionHostname"`
	InstanceID             string    `json:"instanceID"`
	IsDevAppServer         bool      `json:"isDevAppServer"`
	ModuleName             string    `json:"moduleName"`
	ServerSoftware         string    `json:"serverSoftware"`
	VersionID              string    `json:"versionID"`
	ServerTime             time.Time `json:"serverTime"`
	GoVersion              string    `json:"goVersion"`
	CPU                    CPU       `json:"cpu"`
	RAM                    RAM       `json:"ram"`
	Modules                []Module  `json:"modules"`
	Memcache               Memcache  `json:"memcache"`
	Taskqueue              Taskqueue `json:"taskqueue"`
	Caller                 Caller    `json:"caller"`
}

type Module struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

type CPU struct {
	Total   float64 `json:"total"`
	Rate1M  float64 `json:"rate1M"`  // consumption rate over one minute
	Rate10M float64 `json:"rate10M"` // consumption rate over ten minutes
}

// RAM records the memory used by the instance, in megabytes.
type RAM struct {
	Current    float64 `json:"current"`
	Average1M  float64 `json:"average1M"`  // average usage over one minute
	Average10M float64 `json:"average10M"` // average usage over ten minutes
}

type Memcache struct {
	Hits     uint64 `json:"hits"`     // Counter of cache hits
	Misses   uint64 `json:"misses"`   // Counter of cache misses
	ByteHits uint64 `json:"byteHits"` // Counter of bytes transferred for gets
	Items    uint64 `json:"items"`    // Items currently in the cache
	Bytes    uint64 `json:"bytes"`    // Size of all items currently in the cache
	Oldest   int64  `json:"oldest"`   // Age of access of the oldest item, in seconds
}

type Taskqueue struct {
	Name            string    `json:"name"`
	Tasks           int       `json:"tasks"`           //  may be an approximation
	OldestETA       time.Time `json:"oldestETA"`       // zero if there are no pending tasks
	Executed1Minute int       `json:"executed1Minute"` //  tasks executed in the last minute
	InFlight        int       `json:"inFlight"`        //  tasks executing now
	EnforcedRate    float64   `json:"enforcedRate"`    //  requests per second
}

type Caller struct {
	RemoteAddr  string `json:"remoteAddr"`
	UserAgent   string `json:"userAgent"`
	Country     string `json:"country"`
	Region      string `json:"region"`
	City        string `json:"city"`
	CityLatLong string `json:"cityLatLong"`
	Email       string `json:"email"`
}
