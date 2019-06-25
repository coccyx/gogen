package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	ttemplate "text/template"
)

var (
	cache map[string]*ttemplate.Template
	tmpl  *ttemplate.Template
	err   error
)

func init() {
	cache = make(map[string]*ttemplate.Template)
}

// New creates a template and caches it
func New(name string, template string) error {
	if _, ok := cache[name]; !ok {
		funcMap := ttemplate.FuncMap{
			"json": func(v interface{}) string {
				a, _ := json.Marshal(v)
				return string(a)
			},
			"splunkhec": func(v interface{}) string {
				if _, ok := v.(map[string]string)["_raw"]; ok {
					v.(map[string]string)["event"] = v.(map[string]string)["_raw"]
					delete(v.(map[string]string), "_raw")
				}
				if _, ok = v.(map[string]string)["_time"]; ok {
					v.(map[string]string)["time"] = v.(map[string]string)["_time"]
					delete(v.(map[string]string), "_time")
				}
				a, _ := json.Marshal(v)
				return string(a)
			},
			"modinput": func(v interface{}) string {
				ret := "<event>"
				tv := v.(map[string]string)
				if _, ok := tv["_raw"]; ok {
					ret += "<data>" + tv["_raw"] + "</data>"
				}
				if _, ok = tv["_time"]; ok {
					ret += "<time>" + tv["_time"] + "</time>"
				}
				if _, ok = tv["index"]; ok {
					ret += "<index>" + tv["index"] + "</index>"
				}
				if _, ok = tv["host"]; ok {
					ret += "<host>" + tv["host"] + "</host>"
				}
				if _, ok = tv["source"]; ok {
					ret += "<source>" + tv["source"] + "</source>"
				}
				if _, ok = tv["sourcetype"]; ok {
					ret += "<sourcetype>" + tv["sourcetype"] + "</sourcetype>"
				}
				ret += "</event>"
				return ret
			},
			"keys": func(m map[string]string) []string {
				keys := make([]string, len(m))
				i := 0
				for k := range m {
					keys[i] = k
					i++
				}
				sort.Strings(keys)
				return keys
			},
			"values": func(m map[string]string) []string {
				keys := make([]string, len(m))
				values := make([]string, len(m))
				i := 0
				for k := range m {
					keys[i] = k
					i++
				}
				sort.Strings(keys)
				i = 0
				for _, k := range keys {
					values[i] = m[k]
					i++
				}
				return values
			},
			"join": func(arg string, value []string) string {
				return strings.Join(value, arg)
			},
		}
		// Create template, add Func map
		tmpl, err = ttemplate.New(name).Funcs(funcMap).Parse(template)
		if err != nil {
			return err
		}
		cache[name] = tmpl
	}
	return nil
}

// Exists checks whether a given template has been created
func Exists(name string) bool {
	if _, ok := cache[name]; !ok {
		return false
	}
	return true
}

// Exec returns a fully executed template substituted with a string map of row
func Exec(name string, row map[string]string) (string, error) {
	if _, ok := cache[name]; ok {
		tmpl = cache[name]
	} else {
		return "", fmt.Errorf("Exec called for template '%s' but not found in cache", name)
	}
	buf := bytes.NewBufferString("")
	err := tmpl.Execute(buf, row)

	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
