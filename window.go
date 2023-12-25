package webkitgtk

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/ebitengine/purego"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var URLScheme = "app"
var UserAgent = "webkit2gtk"

var windowID uint
var windowIDLock sync.RWMutex

//var showDevTools = func(pointer unsafe.Pointer) {}

type dragInfo struct {
	XRoot       int
	YRoot       int
	DragTime    int
	MouseButton uint
}

func getWindowID() uint {
	windowIDLock.Lock()
	defer windowIDLock.Unlock()
	windowID++
	return windowID
}

type WindowState int

const (
	WindowStateNormal WindowState = iota
	WindowStateMinimised
	WindowStateMaximised
	WindowStateFullscreen
)

type WindowOptions struct {
	// Name is a unique identifier of the window.
	Name string

	// Title is the title of the page.
	Title string

	// URL is the URL of the page to load.
	URL string

	// HTML Content to load (if URL is not set).
	HTML string

	// CSS to load after the page has loaded.
	CSS []string

	// JS to load after the page has loaded.
	JS []string

	// Define global Variables and APIs.
	Define map[string]interface{}

	// Width is the starting width of the window.
	Width int

	// Height is the starting height of the window.
	Height int

	// Overlay will make the window float above other windows.
	Overlay bool

	// Resizeable indicates if the window can be resized.
	// Resizeable bool

	// Frameless will remove the window frame.
	Frameless bool

	// MinWidth is the minimum width of the window.
	MinWidth int

	// MinHeight is the minimum height of the window.
	MinHeight int

	// MaxWidth is the maximum width of the window.
	MaxWidth int

	// MaxHeight is the maximum height of the window.
	MaxHeight int

	// State indicates the state of the window when it is first shown.
	// Default: WindowStateNormal
	State WindowState

	// Centered will Center the window on the screen.
	Centered bool

	// Color specified the window color as a hex string (#RGB, #RGBA, #RRGGBB, #RRGGBBAA)
	Color string
	// BackgroundType is the type of background to use for the window.
	// Default: BackgroundTypeSolid
	//BackgroundType BackgroundType

	// BackgroundColour is the colour to use for the window background.
	//BackgroundColour RGBA

	// X is the starting X position of the window.
	X int

	// Y is the starting Y position  of the window.
	Y int

	// Hidden will Hide the window when it is first created.
	Hidden bool

	// Zoom is the initial Zoom level of the window.
	Zoom float64

	// ZoomControlEnabled will enable the Zoom control.
	ZoomControlEnabled bool

	// OpenInspectorOnStartup will open the inspector when the window is first shown.
	OpenInspectorOnStartup bool

	// Focused indicates the window should be focused when initially shown
	Focused bool

	// If true, the window's devtools will be available
	//DevToolsEnabled bool

	/////////////////

	// HideOnClose will hide the window when it is closed instead of destroying it.
	HideOnClose bool

	// WebkitSettings exposes the underlying WebkitSettings object.
	WebkitSettings WebkitSettings
}

type Window struct {
	options    WindowOptions
	pointer    windowPtr
	id         uint
	app        *App
	logger     *logger
	webview    webviewPtr
	vbox       ptr
	lastWidth  int
	lastHeight int

	bindings  map[string]apiBinding
	constants map[string]string
}

type assetHandler struct {
	registered bool
}

func (ah *assetHandler) handleURISchemeRequest(request uintptr) {
	println("handleURISchemeRequest", request)
}

// Open opens a new window with the given options.
func (a *App) Open(options WindowOptions) *Window {
	if options.Width == 0 {
		options.Width = 800
	}
	if options.Height == 0 {
		options.Height = 600
	}
	if options.Color == "" {
		options.Color = "#FFFFFF"
	}

	newWindow := &Window{
		app:     a,
		id:      getWindowID(),
		options: options,
	}

	if a.logger != nil {
		newWindow.logger = &logger{
			prefix: a.logger.prefix + "-",
			writer: a.logger.writer,
		}
		if options.Name != "" {
			newWindow.logger.prefix += options.Name
		} else {
			newWindow.logger.prefix += strconv.Itoa(int(newWindow.id))
		}
	}

	if options.Define != nil && len(options.Define) > 0 {
		newWindow.constants = make(map[string]string)
		newWindow.bindings = make(map[string]apiBinding)
		for name, v := range options.Define {
			value := reflect.ValueOf(v)
			if value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Struct {
				binding, err := apiBind(v)
				if err != nil {
					panic(err)
				}
				newWindow.bindings[name] = binding
			} else {
				constant, err := json.Marshal(v)
				if err != nil {
					panic(err)
				}
				newWindow.constants[name] = string(constant)
			}

		}
	}

	a.runOnce.add(newWindow)
	//	a.runOrDeferToAppRun(newWindow)
	return newWindow
}

func (w *Window) ID() uint {
	return w.id
}

func (w *Window) run() {
	invokeSync(w.create)
}

// //////////////////////////////////////////////
var registered = false

func (w *Window) create() {
	id := w.ID()

	w.app.windowsLock.Lock()
	w.app.windows[id] = w
	w.app.windowsLock.Unlock()

	openTime := time.Now()
	w.log("creating pointer", "id", w.id, "name", w.options.Name)

	w.pointer = lib.gtk.ApplicationWindowNew(w.app.pointer)
	lib.g.ObjectRefSink(ptr(w.pointer))

	manager := lib.webkit.UserContentManagerNew()

	if w.bindings != nil {
		manager.registerScriptMessageHandler("api", apiHandler(w.bindings, w.ExecJS, w.log))
	}

	w.webview = lib.webkit.WebViewNewWithUserContentManager(manager)

	if !registered {
		webContext := lib.webkit.WebContextGetDefault()
		//dataManager := lib.webkit.WebsiteDataManagerNew()
		cookieManager := lib.webkit.WebContextGetCookieManager(webContext)
		lib.webkit.CookieManagerSetPersistentStorage(cookieManager, "/home/malivvan/webkitdata/xxx", 0)

		//dataManager := lib.webkit.WebsiteDataManagerNew("base-data-directory","~/xxxxxx/")
		//	webContext  := lib.webkit.WebContextNewWithWebsiteDataManager(dataManager)
		//	//	println("sandboxing webview", lib.webkit.WebContextGetSandboxEnabled(webContext))
		//	securityManager := lib.webkit.WebContextGetSecurityManager(webContext)
		//	lib.webkit.SecurityManagerRegisterUriSchemeAsLocal(securityManager, URLScheme)
		//	lib.webkit.SecurityManagerRegisterUriSchemeAsSecure(securityManager, URLScheme)
		//	//webkitSecurityManagerRegisterUriSchemeAsCorsEnabled(securityManager, URLScheme)
		//	//	webkitSecurityManagerRegisterUriSchemeAsNoAccess(securityManager, URLScheme)
		//
		//	lib.webkit.WebContextRegisterUriScheme(
		//		webContext,
		//		URLScheme,
		//		ptr(purego.NewCallback(func(request uintptr) {
		//			//	globalApplication.server.handleURISchemeRequest(request)
		//
		//		})),
		//		0,
		//		0,
		//	)
		registered = true
	}

	settings := lib.webkit.WebViewGetSettings(w.webview)
	defaultWebkitSettings.apply(settings)
	lib.webkitSettings.SetUserAgentWithApplicationDetails(
		settings,
		UserAgent,
		"")
	lib.webkitSettings.SetHardwareAccelerationPolicy(settings, 1)
	lib.webkit.WebViewSetSettings(w.webview, settings)

	w.vbox = lib.gtk.BoxNew(gtkOrientationVertical, 0)
	lib.gtk.ContainerAdd(w.pointer, w.vbox)
	lib.gtk.WidgetSetName(w.vbox, "webview-box")
	lib.gtk.BoxPackStart(w.vbox, ptr(w.webview), 1, 1, 0)

	windowSetupSignalHandlers(w.id, w.pointer, w.webview)
	w.SetTitle(w.options.Title)

	// only set min/max GetSize if actually set
	if w.options.MinWidth != 0 &&
		w.options.MinHeight != 0 &&
		w.options.MaxWidth != 0 &&
		w.options.MaxHeight != 0 {
		w.SetMinMaxSize(
			w.options.MinWidth,
			w.options.MinHeight,
			w.options.MaxWidth,
			w.options.MaxHeight,
		)
	}
	w.SetSize(w.options.Width, w.options.Height)
	w.SetZoom(w.options.Zoom)
	w.SetOverlay(w.options.Overlay)
	w.setBackgroundColour(w.options.Color)
	//w.SetResizeable(!w.options.Resizeable)
	//if w.options.BackgroundType != BackgroundTypeSolid {
	//	w.setTransparent()
	//	w.setBackgroundColour(w.options.BackgroundColour)
	//}

	w.SetFrameless(w.options.Frameless)

	if w.options.X != 0 || w.options.Y != 0 {
		w.setRelativePosition(w.options.X, w.options.Y)
	} else {
		w.Center()
	}
	switch w.options.State {
	case WindowStateMaximised:
		w.Maximise()
	case WindowStateMinimised:
		w.Minimise()
	case WindowStateFullscreen:
		w.Fullscreen()
	}

	if w.options.URL != "" {
		w.SetURL(w.options.URL)
	} else {
		if w.options.HTML != "" {
			w.setHTML(w.options.HTML)
		} else {
			w.setHTML("<html><body>error: no url or html specified in window options</body></html>")
		}
	}

	if !w.options.Hidden {
		w.Show()
		if w.options.X != 0 || w.options.Y != 0 {
			w.setRelativePosition(w.options.X, w.options.Y)
		} else {
			w.Center() // needs to be queued until after GTK starts up!
		}
	}
	if w.app.options.Debug {
		w.ToggleDevTools()
	}

	//w.ExecJS(ipcJS)
	w.log("pointer created", "id", w.id, "name", w.options.Name, "since_open", time.Since(openTime))
}

// Window Callable Methods

func (w *Window) Focus() {
	windowPresent(w.pointer)
}

func (w *Window) Show() {
	windowShow(w.pointer)
}

func (w *Window) Hide() {
	windowHide(w.pointer)
}

func (w *Window) GetZoom() float64 {
	return windowZoom(w.webview)
}

func (w *Window) SetZoom(zoom float64) {
	windowZoomSet(w.webview, zoom)
}

// FIXME: this is not working properly
//func (w *Window) SetResizeable(resizeable bool) {
//	windowSetResizable(w.pointer, !resizeable)
//}

func (w *Window) ToggleDevTools() {
	windowToggleDevTools(w.webview)
}

func (w *Window) GetSize() (int, int) {
	return windowGetSize(w.pointer)
}
func (w *Window) Unfullscreen() {
	windowUnfullscreen(w.pointer)
	w.Unmaximise()
}

func (w *Window) Fullscreen() {
	w.Maximise()
	w.lastWidth, w.lastHeight = w.GetSize()
	x, y, width, height, scale := windowGetCurrentMonitorGeometry(w.pointer)
	if x == -1 && y == -1 && width == -1 && height == -1 {
		return
	}
	w.SetMinMaxSize(0, 0, width*scale, height*scale)
	w.SetSize(width*scale, height*scale)
	windowFullscreen(w.pointer)
	w.setRelativePosition(0, 0)
}

func (w *Window) Unminimise() {
	windowPresent(w.pointer)
}

func (w *Window) Unmaximise() {
	lib.gtk.WindowUnmaximize(w.pointer)
}

func (w *Window) Maximise() {
	windowMaximize(w.pointer)
}

func (w *Window) Minimise() {
	windowMinimize(w.pointer)
}

func (w *Window) SetOverlay(alwaysOnTop bool) {
	windowSetKeepAbove(w.pointer, alwaysOnTop)
}

func (w *Window) SetTitle(title string) {
	if !w.options.Frameless {
		windowSetTitle(w.pointer, title)
	}
}

func (w *Window) SetSize(width, height int) {
	lib.gtk.WindowResize(w.pointer, width, height)
}

func (w *Window) ZoomIn() {
	windowZoomIn(w.webview)
}

func (w *Window) ZoomOut() {
	windowZoomOut(w.webview)
}

func (w *Window) ZoomReset() {
	windowZoomSet(w.webview, 1.0)
}

func (w *Window) Center() {
	x, y, width, height, _ := windowGetCurrentMonitorGeometry(w.pointer)
	if x == -1 && y == -1 && width == -1 && height == -1 {
		return
	}
	windowWidth, windowHeight := windowGetSize(w.pointer)

	newX := ((width - int(windowWidth)) / 2) + x
	newY := ((height - int(windowHeight)) / 2) + y

	// Place the pointer at the Center of the monitor
	windowMove(w.pointer, newX, newY)
}

func (w *Window) SetFrameless(frameless bool) {
	windowSetFrameless(w.pointer, frameless)
}

func (w *Window) IsMinimised() bool {
	return windowIsMinimized(w.pointer)
}

func (w *Window) IsMaximised() bool {
	return windowIsMaximized(w.pointer)
}

func (w *Window) IsFocused() bool {
	return windowIsFocused(w.pointer)
}

func (w *Window) IsFullscreen() bool {
	return windowIsFullscreen(w.pointer)
}

func (w *Window) Close() {
	windowClose(w.pointer)
}

//////////////////////////////////////////////////////////////////////////////

// widgets
func widgetSetSensitive(widget ptr, enabled bool) {
	value := 0
	if enabled {
		value = 1
	}
	lib.gtk.WidgetSetSensitive(widget, value)
}

func widgetSetVisible(widget ptr, hidden bool) {
	if hidden {
		lib.gtk.WidgetHide(widget)
	} else {
		lib.gtk.WidgetShow(widget)
	}
}

func windowClose(window windowPtr) {
	lib.gtk.WindowClose(window)
}

func windowExecJS(webview webviewPtr, js string) {
	lib.webkit.WebViewEvaluateJavascript(
		webview,
		js,
		len(js),
		0,
		"",
		0,
		0,
		0)
}

func windowDestroy(window windowPtr) {
	// Should this truly 'destroy' ?
	lib.gtk.WindowClose(window)
}

func windowFullscreen(window windowPtr) {
	lib.gtk.WindowFullscreen(window)
}

func windowGetAbsolutePosition(window windowPtr) (int, int) {
	var x, y int
	lib.gtk.WindowGetPosition(window, &x, &y)
	return x, y
}

func windowGetCurrentMonitor(window windowPtr) ptr {
	// Get the monitor that the pointer is currently on
	display := lib.gtk.WidgetGetDisplay(window)
	window = lib.gtk.WidgetGetWindow(window)
	if window == 0 {
		return 0
	}
	return lib.gdk.DisplayGetMonitorAtWindow(display, window)
}

func windowGetCurrentMonitorGeometry(window windowPtr) (x int, y int, width int, height int, scale int) {
	monitor := windowGetCurrentMonitor(window)
	if monitor == 0 {
		return -1, -1, -1, -1, 1
	}

	result := struct {
		x      int32
		y      int32
		width  int32
		height int32
	}{}
	lib.gdk.MonitorGetGeometry(monitor, ptr(unsafe.Pointer(&result)))
	return int(result.x), int(result.y), int(result.width), int(result.height), lib.gdk.MonitorGetScaleFactor(monitor)
}

func windowGetRelativePosition(window windowPtr) (int, int) {
	absX, absY := windowGetAbsolutePosition(window)
	x, y, _, _, _ := windowGetCurrentMonitorGeometry(window)

	relX := absX - x
	relY := absY - y

	// TODO: Scale based on DPI
	return relX, relY
}

func windowGetSize(window windowPtr) (int, int) {
	// TODO: dispatchOnMainThread?
	var width, height int
	lib.gtk.WindowGetSize(window, &width, &height)
	return width, height
}

func windowGetPosition(window windowPtr) (int, int) {
	// TODO: dispatchOnMainThread?
	var x, y int
	lib.gtk.WindowGetPosition(window, &x, &y)
	return x, y
}

func windowHide(window windowPtr) {
	lib.gtk.WidgetHide(ptr(window))
}

func windowIsFocused(window windowPtr) bool {
	return lib.gtk.WindowHasToplevelFocus(window) == 1
}

func windowIsFullscreen(window windowPtr) bool {
	gdkwindow := lib.gtk.WidgetGetWindow(window)
	state := lib.gdk.WindowGetState(gdkwindow)
	return state&gdkWindowStateFullscreen > 0
}

func windowIsMaximized(window windowPtr) bool {
	gdkwindow := lib.gtk.WidgetGetWindow(window)
	state := lib.gdk.WindowGetState(gdkwindow)
	return state&gdkWindowStateMaximized > 0 && state&gdkWindowStateFullscreen == 0
}

func windowIsMinimized(window windowPtr) bool {
	gdkwindow := lib.gtk.WidgetGetWindow(window)
	state := lib.gdk.WindowGetState(gdkwindow)
	return state&gdkWindowStateIconified > 0
}

func windowIsVisible(window windowPtr) bool {
	// TODO: validate this works..  (used a `bool` in the registration)
	return lib.gtk.WidgetIsVisible(ptr(window))
}

func windowMaximize(window windowPtr) {
	lib.gtk.WindowMaximize(window)
}

func windowMinimize(window windowPtr) {
	lib.gtk.WindowIconify(window)
}

func (manager userContentManagerPtr) registerScriptMessageHandler(name string, handler func(string)) {
	lib.g.SignalConnectObject(ptr(manager), "script-message-received::"+name, ptr(purego.NewCallback(func(manager ptr, message ptr) {
		handler(lib.jsc.ValueToString(lib.webkit.JavascriptResultGetJsValue(message)))
	})), 0, 0)
	lib.webkit.UserContentManagerRegisterScriptMessageHandler(manager, name)
}

func windowPresent(window windowPtr) {
	lib.gtk.WindowPresent(window)
}

func windowReload(webview webviewPtr, address string) {
	lib.webkit.WebViewLoadUri(webview, address)
}

func windowResize(window windowPtr, width, height int) {
	lib.gtk.WindowResize(window, width, height)
}

func windowShow(window windowPtr) {
	lib.gtk.WidgetShowAll(window)
}

func windowSetGeometryHints(window windowPtr, minWidth, minHeight, maxWidth, maxHeight int) {
	size := gdkGeometry{
		minWidth:  int32(minWidth),
		minHeight: int32(minHeight),
		maxWidth:  int32(maxWidth),
		maxHeight: int32(maxHeight),
	}
	lib.gtk.WindowSetGeometryHints(
		window,
		ptr(0),
		ptr(unsafe.Pointer(&size)),
		gdkHintMinSize|gdkHintMaxSize)
}

func windowSetFrameless(window windowPtr, frameless bool) {
	decorated := 1
	if frameless {
		decorated = 0
	}
	lib.gtk.WindowSetDecorated(window, decorated)

}

// TODO: confirm this is working properly
func windowSetHTML(webview webviewPtr, html string) {
	lib.webkit.WebViewLoadAlternateHtml(webview, html, URLScheme+"://", nil)
}

func windowSetKeepAbove(window windowPtr, alwaysOnTop bool) {
	lib.gtk.WindowSetKeepAbove(window, alwaysOnTop)
}

// FIXME: this is not working properly
//func windowSetResizable(window windowPtr, resizable bool) {
//	width, height := windowGetSize(window)
//	lib.gtk.WindowSetResizable(
//		window,
//		resizable,
//	)
//	windowResize(window, width, height)
//}

func windowSetTitle(window windowPtr, title string) {
	lib.gtk.WindowSetTitle(window, title)
}

func windowSetTransparent(window windowPtr) {
	screen := lib.gtk.WidgetGetScreen(window)
	visual := lib.gdk.ScreenGetRgbaVisual(screen)
	if visual == 0 {
		return
	}
	if lib.gdk.ScreenIsComposited(screen) == 1 {
		lib.gtk.WidgetSetAppPaintable(window, 1)
		lib.gtk.WidgetSetVisual(window, visual)
	}
}

func windowSetURL(webview webviewPtr, uri string) {
	lib.webkit.WebViewLoadUri(webview, uri)
}

func windowSetupSignalHandlers(windowId uint, window windowPtr, webview webviewPtr) {
	handleDelete := purego.NewCallback(func(ptr) {

		globalApplication.windowsLock.RLock()
		appWindow := globalApplication.windows[windowId]
		globalApplication.windowsLock.RUnlock()

		if !appWindow.options.HideOnClose {
			windowDestroy(window)
			appWindow.log("pointer closed", "id", windowId, "name", appWindow.options.Name)

			globalApplication.windowsLock.Lock()
			delete(globalApplication.windows, windowId)
			windowCount := len(globalApplication.windows)
			globalApplication.windowsLock.Unlock()

			if windowCount == 0 && !globalApplication.options.HideOnLastWindowClosed {
				globalApplication.log("last pointer closed, quitting")
				globalApplication.Quit()
			}
		} else {
			appWindow.log("pointer hiding", "id", windowId, "name", appWindow.options.Name)
		}
	})
	lib.g.SignalConnectData(ptr(window), "delete-event", handleDelete, 0, false, 0)

	handleLoadChanged := purego.NewCallback(func(webview ptr, event int, data ptr) {

		switch event {
		case 0: // LOAD_STARTED
		case 1: // LOAD_REDIRECTED
		case 2: // LOAD_COMMITTED
		case 3: // LOAD_FINISHED
			globalApplication.windowsLock.RLock()
			w := globalApplication.windows[windowId]
			globalApplication.windowsLock.RUnlock()

			w.log("initial load finished", "id", windowId, "name", w.options.Name)

			for name, constant := range w.constants {
				w.ExecJS(fmt.Sprintf("const %s = JSON.parse('%s');", name, constant))
			}

			w.ExecJS(apiClient(w.bindings))

			for _, css := range w.options.CSS {
				w.AddCSS(css)
			}

			for _, js := range w.options.JS {
				w.ExecJS(js)
			}
		}
	})
	lib.g.SignalConnectData(ptr(webview), "load-changed", handleLoadChanged, 0, false, 0)

}

func windowToggleDevTools(webview webviewPtr) {
	settings := lib.webkit.WebViewGetSettings(webview)
	lib.webkitSettings.SetEnableDeveloperExtras(
		settings,
		!lib.webkitSettings.GetEnableDeveloperExtras(settings))
}

func windowUnfullscreen(window windowPtr) {
	lib.gtk.WindowUnfullscreen(window)
}

func windowZoom(webview webviewPtr) float64 {
	return lib.webkit.WebViewGetZoomLevel(webview)
}

func windowZoomIn(webview webviewPtr) {
	ZoomInFactor := 1.10
	windowZoomSet(webview, windowZoom(webview)*ZoomInFactor)
}

func windowZoomOut(webview webviewPtr) {
	ZoomOutFactor := -1.10
	windowZoomSet(webview, windowZoom(webview)*ZoomOutFactor)
}

func windowZoomSet(webview webviewPtr, zoom float64) {
	if zoom < 1.0 { // 1.0 is the smallest allowable
		zoom = 1.0
	}
	lib.webkit.WebViewSetZoomLevel(webview, zoom)
}

func windowMove(window windowPtr, x, y int) {
	lib.gtk.WindowMove(window, x, y)
}

////////////////////////////////////////////////

func (w *Window) isNormal() bool {
	return !w.IsMinimised() && !w.IsMaximised() && !w.IsFullscreen()
}

func (w *Window) isVisible() bool {
	return windowIsVisible(w.pointer)
}

func (w *Window) DisableSizeConstraints() {
	x, y, width, height, scale := windowGetCurrentMonitorGeometry(w.pointer)
	w.SetMinMaxSize(x, y, width*scale, height*scale)
}

func (w *Window) Restore() {
	// Restore pointer to normal GetSize
	// FIXME: never called!  - remove from webviewImpl interface
}

func (w *Window) ExecJS(js string) {
	windowExecJS(w.webview, js)
}

func (w *Window) AddCSS(css string) {
	w.ExecJS(fmt.Sprintf("var style = document.createElement('style'); style.innerHTML = `%s`; document.head.appendChild(style);", css))
}

func (w *Window) SetURL(uri string) {
	if uri != "" {
		url, err := url.Parse(uri)
		if err == nil && url.Scheme == "" && url.Host == "" {
			// TODO handle this in a central location, the scheme and host might be platform dependant.
			url.Scheme = URLScheme
			url.Host = URLScheme // TODO: maybe handle differently
			//url.Host = ""
			uri = url.String()
		}
	}
	windowSetURL(w.webview, uri)
}

func (w *Window) SetMinMaxSize(minWidth, minHeight, maxWidth, maxHeight int) {
	if minWidth == 0 {
		minWidth = -1
	}
	if minHeight == 0 {
		minHeight = -1
	}
	if maxWidth == 0 {
		maxWidth = -1
	}
	if maxHeight == 0 {
		maxHeight = -1
	}
	windowSetGeometryHints(w.pointer, minWidth, minHeight, maxWidth, maxHeight)
}

func (w *Window) SetMinSize(width, height int) {
	w.SetMinMaxSize(width, height, w.options.MaxWidth, w.options.MaxHeight)
}

func (w *Window) SetMaxSize(width, height int) {
	w.SetMinMaxSize(w.options.MinWidth, w.options.MinHeight, width, height)
}

func (w *Window) setRelativePosition(x, y int) {
	mx, my, _, _, _ := windowGetCurrentMonitorGeometry(w.pointer)
	windowMove(w.pointer, x+mx, y+my)
}

func (w *Window) setAbsolutePosition(x int, y int) {
	// Set the pointer's absolute position
	windowMove(w.pointer, x, y)
}

func (w *Window) absolutePosition() (int, int) {
	var x, y int
	x, y = windowGetAbsolutePosition(w.pointer)
	return x, y
}

func (w *Window) setTransparent() {
	windowSetTransparent(w.pointer)
}

func (w *Window) setBackgroundColour(s string) error {

	// 1. Decode the given hex colour string
	s = strings.TrimPrefix(s, "#")
	var red, green, blue uint8
	var alpha uint8 = 255
	switch len(s) {
	case 3:
		_, err := fmt.Sscanf(s, "%1x%1x%1x", &red, &green, &blue)
		if err != nil {
			return err
		}
		red *= 17
		green *= 17
		blue *= 17
	case 4:
		_, err := fmt.Sscanf(s, "%1x%1x%1x%1x", &red, &green, &blue, &alpha)
		if err != nil {
			return err
		}
		red *= 17
		green *= 17
		blue *= 17
		alpha *= 17
	case 6:
		_, err := fmt.Sscanf(s, "%2x%2x%2x", &red, &green, &blue)
		if err != nil {
			return err
		}
	case 8:
		_, err := fmt.Sscanf(s, "%2x%2x%2x%2x", &red, &green, &blue, &alpha)
		if err != nil {
			return err
		}
	}

	// 2. Set transparency based on alpha value and screen compositor capability
	screen := lib.gtk.WidgetGetScreen(w.pointer)
	visual := lib.gdk.ScreenGetRgbaVisual(screen)
	if visual != 0 && lib.gdk.ScreenIsComposited(screen) == 1 {
		lib.gtk.WidgetSetAppPaintable(w.pointer, 1)
		lib.gtk.WidgetSetVisual(w.pointer, visual)
	} else {
		alpha = 255 // fallback to solid colour
	}

	// 3. Set the background colour of the webview
	rgba := make([]byte, 4*8)
	rgbaPtr := ptr(unsafe.Pointer(&rgba[0]))
	if !lib.gdk.RgbaParse(rgbaPtr, fmt.Sprintf("rgba(%v,%v,%v,%v)", red, green, blue, float32(alpha)/255.0)) {
		return fmt.Errorf("invalid colour")
	}
	lib.webkit.WebViewSetBackgroundColor(w.webview, rgbaPtr)

	return nil
}

func (w *Window) relativePosition() (int, int) {
	var x, y int
	x, y = windowGetRelativePosition(w.pointer)
	return x, y
}

func (w *Window) destroy() {
	windowDestroy(w.pointer)
}

func (w *Window) setEnabled(enabled bool) {
	widgetSetSensitive(ptr(w.pointer), enabled)
}

func (w *Window) setHTML(html string) {
	windowSetHTML(w.webview, html)
}

func (w *Window) startResize(border string) error {
	// FIXME: what do we need to do here?
	return nil
}

func (w *Window) log(msg interface{}, kv ...interface{}) {
	if w.logger == nil {
		return
	}
	w.logger.log(msg, kv...)
}
