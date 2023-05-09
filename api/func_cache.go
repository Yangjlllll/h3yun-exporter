package api

import (
	"fmt"
	"reflect"
	"time"
)

func FuncCache(duration time.Duration, fn interface{}) interface{} {
	cache := make(map[string][2]interface{})

	return reflect.MakeFunc(reflect.TypeOf(fn),
		func(args []reflect.Value) []reflect.Value {
			now := time.Now().Unix()
			key := fmt.Sprintf("%v", args)
			if val, ok := cache[key]; ok && now-val[0].(int64) < int64(duration.Seconds()) {
				fmt.Println("Using cached result")
				return []reflect.Value{reflect.ValueOf(val[1])}
			}
			ret := reflect.ValueOf(fn).Call(args)
			cache[key] = [2]interface{}{now, ret[0].Interface()}
			return ret
		}).Interface()
}
