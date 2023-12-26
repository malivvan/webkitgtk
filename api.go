package webkitgtk

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

func apiClient(bindings map[string]apiBinding) string {
	calls := make(map[string][]string)
	for api, binding := range bindings {
		for fn := range binding {
			calls[api] = append(calls[api], fn)
		}
	}
	var buf strings.Builder
	if err := apiClientTmpl.Execute(&buf, struct {
		Calls map[string][]string
	}{
		Calls: calls,
	}); err != nil {
		panic(err)
	}
	return buf.String()
}

var apiClientTmpl = template.Must(template.New("api.js").Parse(`(function(document,window) {
class WebkitAPI {
	constructor() {
		this._id = 0;
		this._calls = new Map();
	}
	resolve(id, data) {
		this._calls.get(id)[0](JSON.parse(data));
		this._calls.delete(id);
	}
	reject(id, err) {
		this._calls.get(id)[1](err);
		this._calls.delete(id);
	}
	request(api, fn, obj) {
		let id = this._id++;
		let self = this;
		let msg = id.toString()+" "+api+" "+fn;
		if (obj) msg += " "+JSON.stringify(obj);
		return new Promise((resolve, reject) => {
			self._calls.set(id, [resolve, reject]);
			window.webkit.messageHandlers.api.postMessage(msg);
		});
	}
}
window.webkitAPI = new WebkitAPI();
{{range $api, $calls := .Calls}}window.{{$api}} = {};
{{range $calls}}window.{{$api}}.{{.}} = (obj) => window.webkitAPI.request("{{$api}}", "{{.}}", obj);{{end}}{{end}}
})(document.cloneNode(),globalThis.window);`))

func apiHandler(bindings map[string]apiBinding, eval func(string), log func(interface{}, ...interface{})) func(string) {
	return func(req string) {
		var id, api, fn string
		var cur int
		for i, c := range req {
			if c == ' ' {
				if id == "" {
					id = req[:i]
					cur = i + 1
				} else if api == "" {
					api = req[cur:i]
					cur = i + 1
				} else {
					fn = req[cur:i]
					cur = i + 1
					break
				}
			}
		}
		if fn == "" {
			fn = req[cur:]
			cur = len(req)
		}
		if id == "" || api == "" || fn == "" {
			log("api error", "error", "invalid request", "request", req)
			return
		}

		log("api request", "id", id, "api", api, "fn", fn)
		binding, ok := bindings[api]
		if !ok {
			eval("webkitAPI.reject(" + string(id) + ",'api not found')")
			return
		}
		reply, err := binding.call(fn, req[cur:])
		if err != nil {
			log("api reject", "id", id, "error", err)
			eval("webkitAPI.reject(" + string(id) + ",'" + err.Error() + "')")
			return
		}
		log("api resolve", "id", id, "reply", reply)
		eval("webkitAPI.resolve(" + id + ",'" + reply + "')")
	}
}

type apiBinding map[string]func(string) (string, error)

func apiBind(api interface{}) (apiBinding, error) {
	value := reflect.ValueOf(api)
	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("api is not a struct pointer")
	}
	binding := make(apiBinding)
	for i := 0; i < value.NumMethod(); i++ {
		method := value.Method(i)

		fn := value.Type().Method(i).Name
		fn = strings.ToLower(string(fn[0])) + fn[1:]
		if _, exists := binding[fn]; exists {
			return nil, fmt.Errorf("function %s already exists", fn)
		}
		println(fn)

		var hasInput, hasOutput bool
		var inputType reflect.Type
		inputCount := method.Type().NumIn()
		if inputCount > 0 {
			if inputCount > 1 {
				return nil, fmt.Errorf("function has too many inputs")
			}
			hasInput = true
			inputType = method.Type().In(0)
		}
		outputCount := method.Type().NumOut()
		if outputCount < 1 || !method.Type().Out(outputCount-1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return nil, fmt.Errorf("function requires an error as last output")
		}
		if outputCount > 1 {
			if outputCount > 2 {
				return nil, fmt.Errorf("function has too many outputs")
			}
			hasOutput = true
		}

		binding[fn] = func(s string) (string, error) {
			var inputs []reflect.Value
			if hasInput {
				input := reflect.New(inputType).Elem()
				if err := json.Unmarshal([]byte(s), input.Addr().Interface()); err != nil {
					return "", err
				}
				inputs = []reflect.Value{input}
			}
			outputs := method.Call(inputs)
			if len(outputs) != outputCount {
				return "", fmt.Errorf("call did not return enought values")
			}
			if err := outputs[outputCount-1].Interface(); err != nil {
				return "", err.(error)
			}
			if hasOutput {
				output, err := json.Marshal(outputs[0].Interface())
				if err != nil {
					return "", err
				}
				return string(output), nil
			}
			return "", nil
		}
	}
	return binding, nil
}

func (api apiBinding) call(name string, input string) (string, error) {
	if fn, ok := api[name]; ok {
		return fn(input)
	}
	return "", fmt.Errorf("function %s not found", name)
}
