package exp

import "expvar"

func GetMap(name string) *expvar.Map {
	if p := expvar.Get(name); p != nil {
		return p.(*expvar.Map)
	}

	return expvar.NewMap(name)
}
