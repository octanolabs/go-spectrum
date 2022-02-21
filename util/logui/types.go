package logui

import (
	"fmt"
	"io"
	"time"

	"github.com/ubiq/go-ubiq/v7/log"
)

type keyValuePair map[string]string

func (kvp keyValuePair) key() (key string) {
	for k := range kvp {
		key = k
	}
	return
}

func (kvp keyValuePair) value() (val string) {
	for _, v := range kvp {
		val = v
	}
	return
}

func makeKvp(key string, val interface{}) (kvp keyValuePair, ok bool) {

	var (
		valString string
	)

	kvp = make(map[string]string, 1)

	if valString, ok = val.(string); !ok {
		return keyValuePair{}, false
	}

	kvp[key] = valString
	return
}

type LoggerContext []keyValuePair

func (lc LoggerContext) Equals(ctx LoggerContext) bool {

	if len(lc) == len(ctx) {
		for k, v := range lc {
			if v.key() != ctx[k].key() {
				return false
			}

			if v.value() != ctx[k].value() {
				return false
			}
		}
	} else {
		return false
	}

	return true
}

func (lc LoggerContext) String() (str string) {

	for _, v := range lc {
		if len(str) == 0 {

			str = v.value()
		} else {
			str = str + " -> " + v.value()
		}
	}
	return
}

func (lc LoggerContext) LoggerString() (str []interface{}) {

	for _, v := range lc {
		str = append(str, v.key(), v.value())
	}
	return
}

func getContext(ctx []interface{}) (modules LoggerContext) {

	modules = LoggerContext{}

loop:
	for i, v := range ctx {
		if i%2 == 0 {

			//TODO: remove hardcoded keys
			// "pkg"
			// "module"
			// "dependency"

			if key, ok := v.(string); ok && (key == "pkg" || key == "module" || key == "dependency") {
				kvp, ok := makeKvp(key, ctx[i+1])
				//if value isn't a string we break the loop
				if !ok {
					break loop
				}
				modules = append(modules, kvp)
			} else {
				//if key isn't a string we break the loop
				break loop
			}
		}

	}
	return
}

func makeNewHandler(lc LoggerContext, writer io.Writer) (handler log.Handler) {

	handler = log.StreamHandler(writer, log.TerminalFormat(false))

	handler = log.FilterHandler(func(r *log.Record) (match bool) {
		ctx := getContext(r.Ctx)
		if ctx.Equals(lc) {
			match = true
		}
		return
	}, handler)

	err := handler.Log(&log.Record{
		Time: time.Now(),
		Lvl:  3,
		Msg:  "Created handler",
		Ctx:  lc.LoggerString(),
	})

	if err != nil {
		fmt.Printf("err loggin to handler: %v", err)
	}

	return
}
