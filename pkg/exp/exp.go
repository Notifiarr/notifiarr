//nolint:gochecknoglobals
package exp

import (
	"expvar"
	"strings"
)

var mainMap = expvar.NewMap("notifiarr").Init()

var (
	LogFiles      = GetMap("Log File Information").Init()
	APIHits       = GetMap("Incoming API Requests").Init()
	HTTPRequests  = GetMap("Incoming HTTP Requests").Init()
	TimerEvents   = GetMap("Triggers and Timers Executed").Init()
	NotifiarrCom  = GetMap("Outbound Requests to Notifiarr.com").Init()
	ServiceChecks = GetMap("Service Check Responses").Init() //nolint:gochecknoglobals
)

type AllData struct {
	LogFiles      map[string]string
	APIHits       map[string]string
	HTTPRequests  map[string]string
	TimerEvents   map[string]map[string]string
	NotifiarrCom  map[string]string
	ServiceChecks map[string]map[string]string
}

func GetAllData() AllData {
	return AllData{
		LogFiles:      GetKeys(LogFiles),
		APIHits:       GetKeys(APIHits),
		HTTPRequests:  GetKeys(HTTPRequests),
		TimerEvents:   GetSplitKeys(TimerEvents),
		NotifiarrCom:  GetKeys(NotifiarrCom),
		ServiceChecks: GetSplitKeys(ServiceChecks),
	}
}

func GetMap(name string) *expvar.Map {
	if p := mainMap.Get(name); p != nil {
		return p.(*expvar.Map)
	}

	newMap := expvar.NewMap(name)
	mainMap.Set(name, newMap)

	return newMap
}

func GetKeys(mapName *expvar.Map) map[string]string {
	output := make(map[string]string)

	mapName.Do(func(kv expvar.KeyValue) {
		output[kv.Key] = kv.Value.String()
	})

	return output
}

//nolint:gomnd
func GetSplitKeys(mapName *expvar.Map) map[string]map[string]string {
	output := make(map[string]map[string]string)

	mapName.Do(func(keyval expvar.KeyValue) {
		keys := strings.SplitN(keyval.Key, "&&", 2)
		if len(keys) != 2 {
			return
		}

		if output[keys[0]] == nil {
			output[keys[0]] = make(map[string]string)
		}

		output[keys[0]][keys[1]] = keyval.Value.String()
	})

	return output
}
