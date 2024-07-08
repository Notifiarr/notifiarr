//nolint:gochecknoglobals
package mnd

import (
	"expvar"
	"strings"
)

/* This file powers all the exported metrics. */

var mainMap = expvar.NewMap("notifiarr").Init()

var (
	LogFiles      = GetMap("Log File Information").Init()
	APIHits       = GetMap("Incoming API Requests").Init()
	HTTPRequests  = GetMap("Incoming HTTP Requests").Init()
	TimerEvents   = GetMap("Triggers and Timers Executed").Init()
	TimerCounts   = GetMap("Triggers and Timers Counters").Init()
	Website       = GetMap("Outbound Requests to Website").Init()
	ServiceChecks = GetMap("Service Check Responses").Init()
	Apps          = GetMap("Starr App Requests").Init()
	FileWatcher   = GetMap("File Watcher").Init()
)

type AllData struct {
	LogFiles      map[string]interface{}
	APIHits       map[string]interface{}
	HTTPRequests  map[string]interface{}
	TimerEvents   map[string]map[string]interface{}
	TimerCounts   map[string]interface{}
	Website       map[string]interface{}
	ServiceChecks map[string]map[string]interface{}
	Apps          map[string]map[string]interface{}
	FileWatcher   map[string]interface{}
}

func GetAllData() AllData {
	return AllData{
		LogFiles:      GetKeys(LogFiles),
		APIHits:       GetKeys(APIHits),
		HTTPRequests:  GetKeys(HTTPRequests),
		TimerEvents:   GetSplitKeys(TimerEvents),
		TimerCounts:   GetKeys(TimerCounts),
		Website:       GetKeys(Website),
		ServiceChecks: GetSplitKeys(ServiceChecks),
		Apps:          GetSplitKeys(Apps),
		FileWatcher:   GetKeys(FileWatcher),
	}
}

func GetMap(name string) *expvar.Map {
	if p := mainMap.Get(name); p != nil {
		pp, _ := p.(*expvar.Map)
		return pp
	}

	newMap := expvar.NewMap(name)
	mainMap.Set(name, newMap)

	return newMap
}

func GetKeys(mapName *expvar.Map) map[string]interface{} {
	output := make(map[string]interface{})

	mapName.Do(func(keyval expvar.KeyValue) {
		switch v := keyval.Value.(type) {
		case *expvar.Int:
			output[keyval.Key] = v.Value()
		case expvar.Func:
			output[keyval.Key], _ = v.Value().(int64)
		default:
			output[keyval.Key] = keyval.Value
		}
	})

	return output
}

//nolint:mnd
func GetSplitKeys(mapName *expvar.Map) map[string]map[string]interface{} {
	output := make(map[string]map[string]interface{})

	mapName.Do(func(keyval expvar.KeyValue) {
		keys := strings.SplitN(keyval.Key, "&&", 2)
		if len(keys) != 2 {
			return
		}

		if output[keys[0]] == nil {
			output[keys[0]] = make(map[string]interface{})
		}

		switch v := keyval.Value.(type) {
		case *expvar.Int:
			output[keys[0]][keys[1]] = v.Value()
		case *expvar.Func:
			output[keys[0]][keys[1]], _ = v.Value().(int64)
		default:
			output[keys[0]][keys[1]] = keyval.Value
		}
	})

	return output
}
