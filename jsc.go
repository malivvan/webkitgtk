package webkitgtk

import (
	"errors"
	"github.com/ebitengine/purego"
)

// TODO: check for memory leaks in jsc.go

func (w *Window) JSPromise(js string, fn func(interface{})) func() {
	js = "return new Promise((resolve, reject) => { \n" + js + "\n });"
	return w.JSCall(js, fn)
}

func (w *Window) JSCall(js string, fn func(interface{})) func() {
	cancelable := lib.g.CancellableNew()
	lib.webkit.WebViewCallAsyncJavascriptFunction(
		w.webview, js, len(js), 0, 0, 0, cancelable,
		ptr(purego.NewCallback(func(webview webviewPtr, result ptr) {
			var gErr *gError
			jsc := lib.webkit.WebViewCallAsyncJavascriptFunctionFinish(webview, result, gErr)
			fn(parseJSC(jsc, cancelable, gErr))
			//lib.webkit.JavascriptResultUnref(result)
		})), 0)
	return func() {
		lib.g.CancellableCancel(cancelable)
	}
}

func (w *Window) JSEval(js string, fn func(interface{})) func() {
	cancelable := lib.g.CancellableNew()
	lib.webkit.WebViewEvaluateJavascript(
		w.webview, js, len(js), 0, 0, cancelable,
		ptr(purego.NewCallback(func(webview webviewPtr, result ptr) {
			var gErr *gError
			jsc := lib.webkit.WebViewEvaluateJavascriptFinish(webview, result, gErr)
			fn(parseJSC(jsc, cancelable, gErr))
			//lib.webkit.JavascriptResultUnref(result)
		})), 0)
	return func() {
		lib.g.CancellableCancel(cancelable)
	}
}

type JSObject struct {
	pointer ptr
}

func (jso *JSObject) Get(name string) interface{} {
	value := lib.jsc.ValueObjectGetProperty(jso.pointer, name)
	return parseJSC(value, 0, nil)
}

func parseJSC(value ptr, cancelable ptr, gErr *gError) interface{} {
	if value == 0 {
		if lib.g.CancellableIsCancelled(cancelable) {
			return errors.New("call canceled")
		}
		err := errors.New("call error")
		if gErr != nil {
			err = gErr.toError(err.Error())
			lib.g.ErrorFree(gErr)
		}
		return err
	}
	exception := lib.jsc.ContextGetException(lib.jsc.ValueGetContext(value))
	if exception != 0 {
		return errors.New(lib.jsc.ExceptionGetMessage(exception))
	}
	if lib.jsc.ValueIsNumber(value) {
		return lib.jsc.ValueToDouble(value)
	} else if lib.jsc.ValueIsBoolean(value) {
		return lib.jsc.ValueToBoolean(value)
	} else if lib.jsc.ValueIsString(value) {
		return lib.jsc.ValueToString(value)
	} else if lib.jsc.ValueIsObject(value) {
		return &JSObject{value}
	}
	return errors.New("jscToValue: unknown type")
}
