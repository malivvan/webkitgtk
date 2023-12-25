package webkitgtk

import (
	"fmt"
	"github.com/ebitengine/purego"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type (
	ptr                   uintptr
	gsListPtr             *gsList
	webkitSettingsPtr     uintptr
	webviewPtr            uintptr
	windowPtr             uintptr
	userContentManagerPtr uintptr
	gsList                struct {
		data ptr
		next *gsList
	}
	gList struct {
		data ptr
		next *gList
		prev *gList
	}
	gdkGeometry struct {
		minWidth   int32
		minHeight  int32
		maxWidth   int32
		maxHeight  int32
		baseWidth  int32
		baseHeight int32
		widthInc   int32
		heightInc  int32
		padding    int32
		minAspect  float64
		maxAspect  float64
		GdkGravity int32
	}
)

const (
	gSourceRemove int = 0

	gdkHintMinSize = 1 << 1
	gdkHintMaxSize = 1 << 2

	gdkWindowStateIconified  = 1 << 1
	gdkWindowStateMaximized  = 1 << 2
	gdkWindowStateFullscreen = 1 << 4

	gtkOrientationVertical = 1
)

var libs = [][]string{
	{"gtk-3", "webkit2gtk-4.1"},
	{"gtk-4", "webkitgtk-6.0"},
}

var lib struct {
	Target  string
	Version int

	GTK    uintptr
	Webkit uintptr

	g struct {
		ApplicationHold      func(ptr)
		ApplicationQuit      func(ptr)
		ApplicationRegister  func(ptr, ptr, ptr)
		ApplicationActivate  func(ptr)
		GetApplicationName   func() string
		ApplicationRelease   func(ptr)
		ApplicationRun       func(ptr, int, []string) int
		BytesNewStatic       func(uintptr, int) uintptr
		BytesUnref           func(uintptr)
		Free                 func(ptr)
		IdleAdd              func(uintptr)
		ObjectRefSink        func(ptr)
		ObjectUnref          func(ptr)
		SignalConnectData    func(ptr, string, uintptr, ptr, bool, int) int
		SignalConnectObject  func(ptr, string, ptr, ptr, int) uint
		SignalHandlerBlock   func(ptr, uint)
		SignalHandlerUnblock func(ptr, uint)
		ThreadSelf           func() uint64
	}
	gdk struct {
		DisplayGetMonitor         func(ptr, int) ptr
		DisplayGetMonitorAtWindow func(ptr, windowPtr) ptr
		DisplayGetNMonitors       func(ptr) int
		MonitorGetGeometry        func(ptr, ptr) ptr
		MonitorGetScaleFactor     func(ptr) int
		MonitorIsPrimary          func(ptr) int
		PixbufNewFromBytes        func(uintptr, int, int, int, int, int, int) ptr
		RgbaParse                 func(ptr, string) bool
		ScreenGetRgbaVisual       func(ptr) ptr
		ScreenIsComposited        func(ptr) int
		WindowGetState            func(windowPtr) int
		WindowGetDisplay          func(ptr) ptr
	}
	gtk struct {
		ApplicationNew             func(string, uint) ptr
		ApplicationGetActiveWindow func(ptr) ptr

		ApplicationGetWindows        func(ptr) *gList
		ApplicationWindowNew         func(ptr) windowPtr
		BoxNew                       func(int, int) ptr
		BoxPackStart                 func(ptr, ptr, int, int, int)
		CheckMenuItemGetActive       func(ptr) int
		CheckMenuItemNewWithLabel    func(string) ptr
		CheckMenuItemSetActive       func(ptr, int)
		ContainerAdd                 func(windowPtr, ptr)
		CssProviderLoadFromData      func(ptr, string, int, ptr)
		CssProviderNew               func() ptr
		DialogAddButton              func(ptr, string, int)
		DialogGetContentArea         func(ptr) ptr
		DialogRun                    func(ptr) int
		DialogSetDefaultResponse     func(ptr, int)
		DragDestSet                  func(webviewPtr, uint, ptr, uint, uint)
		FileChooserAddFilter         func(ptr, ptr)
		FileChooserDialogNew         func(string, ptr, int, string, int, string, int, ptr) ptr
		FileChooserGetFilenames      func(ptr) *gsList
		FileChooserSetAction         func(ptr, int)
		FileChooserSetCreateFolders  func(ptr, bool)
		FileChooserSetCurrentFolder  func(ptr, string)
		FileChooserSetSelectMultiple func(ptr, bool)
		FileChooserSetShowHidden     func(ptr, bool)
		FileFilterAddPattern         func(ptr, string)
		FileFilterNew                func() ptr
		FileFilterSetName            func(ptr, string)
		ImageNewFromPixbuf           func(ptr) ptr
		MenuBarNew                   func() ptr
		MenuItemNewWithLabel         func(string) ptr
		MenuItemSetLabel             func(ptr, string)
		MenuItemSetSubmenu           func(ptr, ptr)
		MenuNew                      func() ptr
		MenuShellAppend              func(ptr, ptr)
		MessageDialogNew             func(ptr, int, int, int, string) ptr
		//RadioMenuItemGetGroup        func(ptr) gsListPtr
		//RadioMenuItemNewWithLabel    func(gsListPtr, string) ptr
		SeparatorMenuItemNew    func() ptr
		StyleContextAddProvider func(ptr, ptr, int)
		TargetEntryFree         func(ptr)
		TargetEntryNew          func(string, int, uint) ptr
		WidgetDestroy           func(windowPtr)
		WidgetGetDisplay        func(windowPtr) ptr
		WidgetGetScreen         func(windowPtr) ptr
		WidgetGetStyleContext   func(windowPtr) ptr
		WidgetGetWindow         func(windowPtr) windowPtr
		WidgetHide              func(ptr)
		WidgetIsVisible         func(ptr) bool
		WidgetShow              func(ptr)
		WidgetShowAll           func(windowPtr)
		WidgetSetAppPaintable   func(windowPtr, int)
		WidgetSetName           func(ptr, string)
		WidgetSetSensitive      func(ptr, int)
		WidgetSetTooltipText    func(windowPtr, string)
		WidgetSetVisual         func(windowPtr, ptr)
		WindowClose             func(windowPtr)
		WindowFullscreen        func(windowPtr)
		WindowGetPosition       func(windowPtr, *int, *int) bool
		WindowGetSize           func(windowPtr, *int, *int)
		WindowHasToplevelFocus  func(windowPtr) int
		//WindowKeepAbove              func(pointer, bool)
		WindowMaximize         func(windowPtr)
		WindowIconify          func(windowPtr)
		WindowMove             func(windowPtr, int, int)
		WindowPresent          func(windowPtr)
		WindowResize           func(windowPtr, int, int)
		WindowSetDecorated     func(windowPtr, int)
		WindowSetGeometryHints func(windowPtr, ptr, ptr, int)
		WindowSetKeepAbove     func(windowPtr, bool)
		WindowSetResizable     func(windowPtr, bool)
		WindowSetTitle         func(windowPtr, string)
		WindowUnfullscreen     func(windowPtr)
		WindowUnmaximize       func(windowPtr)
	}
	webkitSettings struct {
		GetEnableJavascript                          func(webkitSettingsPtr) bool
		SetEnableJavascript                          func(webkitSettingsPtr, bool)
		GetAutoLoadImages                            func(webkitSettingsPtr) bool
		SetAutoLoadImages                            func(webkitSettingsPtr, bool)
		GetLoadIconsIgnoringImageLoadSetting         func(webkitSettingsPtr) bool
		SetLoadIconsIgnoringImageLoadSetting         func(webkitSettingsPtr, bool)
		GetEnableOfflineWebApplicationCache          func(webkitSettingsPtr) bool
		SetEnableOfflineWebApplicationCache          func(webkitSettingsPtr, bool)
		GetEnableHtml5LocalStorage                   func(webkitSettingsPtr) bool
		SetEnableHtml5LocalStorage                   func(webkitSettingsPtr, bool)
		GetEnableHtml5Database                       func(webkitSettingsPtr) bool
		SetEnableHtml5Database                       func(webkitSettingsPtr, bool)
		GetEnableXssAuditor                          func(webkitSettingsPtr) bool
		SetEnableXssAuditor                          func(webkitSettingsPtr, bool)
		GetEnableFrameFlattening                     func(webkitSettingsPtr) bool
		SetEnableFrameFlattening                     func(webkitSettingsPtr, bool)
		GetEnablePlugins                             func(webkitSettingsPtr) bool
		SetEnablePlugins                             func(webkitSettingsPtr, bool)
		GetEnableJava                                func(webkitSettingsPtr) bool
		SetEnableJava                                func(webkitSettingsPtr, bool)
		GetJavascriptCanOpenWindowsAutomatically     func(webkitSettingsPtr) bool
		SetJavascriptCanOpenWindowsAutomatically     func(webkitSettingsPtr, bool)
		GetEnableHyperlinkAuditing                   func(webkitSettingsPtr) bool
		SetEnableHyperlinkAuditing                   func(webkitSettingsPtr, bool)
		GetDefaultFontFamily                         func(webkitSettingsPtr) string
		SetDefaultFontFamily                         func(webkitSettingsPtr, string)
		GetMonospaceFontFamily                       func(webkitSettingsPtr) string
		SetMonospaceFontFamily                       func(webkitSettingsPtr, string)
		GetSerifFontFamily                           func(webkitSettingsPtr) string
		SetSerifFontFamily                           func(webkitSettingsPtr, string)
		GetSansSerifFontFamily                       func(webkitSettingsPtr) string
		SetSansSerifFontFamily                       func(webkitSettingsPtr, string)
		GetCursiveFontFamily                         func(webkitSettingsPtr) string
		SetCursiveFontFamily                         func(webkitSettingsPtr, string)
		GetFantasyFontFamily                         func(webkitSettingsPtr) string
		SetFantasyFontFamily                         func(webkitSettingsPtr, string)
		GetPictographFontFamily                      func(webkitSettingsPtr) string
		SetPictographFontFamily                      func(webkitSettingsPtr, string)
		GetDefaultFontSize                           func(webkitSettingsPtr) uint32
		SetDefaultFontSize                           func(webkitSettingsPtr, uint32)
		GetDefaultMonospaceFontSize                  func(webkitSettingsPtr) uint32
		SetDefaultMonospaceFontSize                  func(webkitSettingsPtr, uint32)
		GetMinimumFontSize                           func(webkitSettingsPtr) uint32
		SetMinimumFontSize                           func(webkitSettingsPtr, uint32)
		GetDefaultCharset                            func(webkitSettingsPtr) string
		SetDefaultCharset                            func(webkitSettingsPtr, string)
		GetEnableDeveloperExtras                     func(webkitSettingsPtr) bool
		SetEnableDeveloperExtras                     func(webkitSettingsPtr, bool)
		GetEnableResizableTextAreas                  func(webkitSettingsPtr) bool
		SetEnableResizableTextAreas                  func(webkitSettingsPtr, bool)
		GetEnableTabsToLinks                         func(webkitSettingsPtr) bool
		SetEnableTabsToLinks                         func(webkitSettingsPtr, bool)
		GetEnableDnsPrefetching                      func(webkitSettingsPtr) bool
		SetEnableDnsPrefetching                      func(webkitSettingsPtr, bool)
		GetEnableCaretBrowsing                       func(webkitSettingsPtr) bool
		SetEnableCaretBrowsing                       func(webkitSettingsPtr, bool)
		GetEnableFullscreen                          func(webkitSettingsPtr) bool
		SetEnableFullscreen                          func(webkitSettingsPtr, bool)
		GetPrintBackgrounds                          func(webkitSettingsPtr) bool
		SetPrintBackgrounds                          func(webkitSettingsPtr, bool)
		GetEnableWebaudio                            func(webkitSettingsPtr) bool
		SetEnableWebaudio                            func(webkitSettingsPtr, bool)
		GetEnableWebgl                               func(webkitSettingsPtr) bool
		SetEnableWebgl                               func(webkitSettingsPtr, bool)
		SetAllowModalDialogs                         func(webkitSettingsPtr, bool)
		GetAllowModalDialogs                         func(webkitSettingsPtr) bool
		SetZoomTextOnly                              func(webkitSettingsPtr, bool)
		GetZoomTextOnly                              func(webkitSettingsPtr) bool
		GetJavascriptCanAccessClipboard              func(webkitSettingsPtr) bool
		SetJavascriptCanAccessClipboard              func(webkitSettingsPtr, bool)
		GetMediaPlaybackRequiresUserGesture          func(webkitSettingsPtr) bool
		SetMediaPlaybackRequiresUserGesture          func(webkitSettingsPtr, bool)
		GetMediaPlaybackAllowsInline                 func(webkitSettingsPtr) bool
		SetMediaPlaybackAllowsInline                 func(webkitSettingsPtr, bool)
		GetDrawCompositingIndicators                 func(webkitSettingsPtr) bool
		SetDrawCompositingIndicators                 func(webkitSettingsPtr, bool)
		GetEnableSiteSpecificQuirks                  func(webkitSettingsPtr) bool
		SetEnableSiteSpecificQuirks                  func(webkitSettingsPtr, bool)
		GetEnablePageCache                           func(webkitSettingsPtr) bool
		SetEnablePageCache                           func(webkitSettingsPtr, bool)
		GetUserAgent                                 func(webkitSettingsPtr) string
		SetUserAgent                                 func(webkitSettingsPtr, string)
		SetUserAgentWithApplicationDetails           func(webkitSettingsPtr, string, string)
		GetEnableSmoothScrolling                     func(webkitSettingsPtr) bool
		SetEnableSmoothScrolling                     func(webkitSettingsPtr, bool)
		GetEnableAccelerated2DCanvas                 func(webkitSettingsPtr) bool  `name:"webkit_settings_get_enable_accelerated_2d_canvas"`
		SetEnableAccelerated2DCanvas                 func(webkitSettingsPtr, bool) `name:"webkit_settings_set_enable_accelerated_2d_canvas"`
		GetEnableWriteConsoleMessagesToStdout        func(webkitSettingsPtr) bool
		SetEnableWriteConsoleMessagesToStdout        func(webkitSettingsPtr, bool)
		GetEnableMediaStream                         func(webkitSettingsPtr) bool
		SetEnableMediaStream                         func(webkitSettingsPtr, bool)
		GetEnableMockCaptureDevices                  func(webkitSettingsPtr) bool
		SetEnableMockCaptureDevices                  func(webkitSettingsPtr, bool)
		GetEnableSpatialNavigation                   func(webkitSettingsPtr) bool
		SetEnableSpatialNavigation                   func(webkitSettingsPtr, bool)
		GetEnableMediasource                         func(webkitSettingsPtr) bool
		SetEnableMediasource                         func(webkitSettingsPtr, bool)
		GetEnableEncryptedMedia                      func(webkitSettingsPtr) bool
		SetEnableEncryptedMedia                      func(webkitSettingsPtr, bool)
		GetEnableMediaCapabilities                   func(webkitSettingsPtr) bool
		SetEnableMediaCapabilities                   func(webkitSettingsPtr, bool)
		GetAllowFileAccessFromFileUrls               func(webkitSettingsPtr) bool
		SetAllowFileAccessFromFileUrls               func(webkitSettingsPtr, bool)
		GetAllowUniversalAccessFromFileUrls          func(webkitSettingsPtr) bool
		SetAllowUniversalAccessFromFileUrls          func(webkitSettingsPtr, bool)
		GetAllowTopNavigationToDataUrls              func(webkitSettingsPtr) bool
		SetAllowTopNavigationToDataUrls              func(webkitSettingsPtr, bool)
		GetHardwareAccelerationPolicy                func(webkitSettingsPtr) int
		SetHardwareAccelerationPolicy                func(webkitSettingsPtr, int)
		GetEnableBackForwardNavigationGestures       func(webkitSettingsPtr) bool
		SetEnableBackForwardNavigationGestures       func(webkitSettingsPtr, bool)
		FontSizeToPoints                             func(uint32) uint32
		FontSizeToPixels                             func(uint32) uint32
		GetEnableJavascriptMarkup                    func(webkitSettingsPtr) bool
		SetEnableJavascriptMarkup                    func(webkitSettingsPtr, bool)
		GetEnableMedia                               func(webkitSettingsPtr) bool
		SetEnableMedia                               func(webkitSettingsPtr, bool)
		GetMediaContentTypesRequiringHardwareSupport func(webkitSettingsPtr) string
		SetMediaContentTypesRequiringHardwareSupport func(webkitSettingsPtr, string)
		GetEnableWebrtc                              func(webkitSettingsPtr) bool
		SetEnableWebrtc                              func(webkitSettingsPtr, bool)
		GetDisableWebSecurity                        func(webkitSettingsPtr) bool
		SetDisableWebSecurity                        func(webkitSettingsPtr, bool)
	}
	jsc struct {
		ValueToString        func(ptr) string
		ValueToStringAsBytes func(ptr) string
		ValueIsString        func(ptr) bool
	}
	webkit struct {
		//WebViewNew                       func() ptr
		WebViewNewWithUserContentManager func(userContentManagerPtr) webviewPtr
		WebContextRegisterUriScheme      func(ptr, string, ptr, int, int)
		//webkitSettingsGetEnableDeveloperExtras                  func(pointer) bool
		//webkitSettingsSetHardwareAccelerationPolicy             func(pointer, int)
		//webkitSettingsSetEnableDeveloperExtras                  func(pointer, bool)
		//webkitSettingsSetUserAgentWithApplicationDetails        func(pointer, string, string)
		UserContentManagerNew                          func() userContentManagerPtr
		UserContentManagerRegisterScriptMessageHandler func(userContentManagerPtr, string)

		//	WebsiteDataManagerNew               func(...string) ptr
		//WebContextNewWithWebsiteDataManager func(ptr) ptr
		//web_context_get_cookie_manager
		// void webkit_cookie_manager_set_storage
		CookieManagerSetPersistentStorage func(ptr, string, int)
		WebContextGetCookieManager        func(ptr) ptr
		WebViewGetUserContentManager      func(webviewPtr) userContentManagerPtr

		WebsiteDataManagerNew               func(...string) ptr
		WebContextNewWithWebsiteDataManager func(ptr) ptr
		WebContextGetSandboxEnabled         func(ptr) bool
		WebContextSetSandboxEnabled         func(ptr, bool)
		WebContextAddPathToSandbox          func(ptr, string)
		WebContextGetSpellCheckingEnabled   func(ptr) bool
		WebContextSetSpellCheckingEnabled   func(ptr, bool)

		JavascriptResultGetJsValue func(ptr) ptr
		WebContextGetDefault       func() ptr

		WebContextGetSecurityManager                      func(ptr) ptr
		SecurityManagerRegisterUriSchemeAsSecure          func(ptr, string)
		SecurityManagerRegisterUriSchemeAsNoAccess        func(ptr, string)
		SecurityManagerRegisterUriSchemeAsDisplayIsolated func(ptr, string)
		SecurityManagerRegisterUriSchemeAsCorsEnabled     func(ptr, string)
		SecurityManagerRegisterUriSchemeAsLocal           func(ptr, string)

		WebViewEvaluateJavascript func(webviewPtr, string, int, ptr, string, ptr, ptr, ptr)
		WebViewGetSettings        func(webviewPtr) webkitSettingsPtr
		WebViewGetZoomLevel       func(webviewPtr) float64
		WebViewLoadAlternateHtml  func(webviewPtr, string, string, *string)
		WebViewLoadUri            func(webviewPtr, string)
		WebViewSetBackgroundColor func(webviewPtr, ptr)
		WebViewSetSettings        func(webviewPtr, webkitSettingsPtr)
		WebViewSetZoomLevel       func(webviewPtr, float64)
		WebViewLoadBytes          func(webviewPtr, []byte, string, string, string)
	}
}

func registerFunctions(lib uintptr, prefix string, v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Pointer {
		return fmt.Errorf("v must be a struct pointer")
	}
	vElem := reflect.ValueOf(v).Elem()
	if vElem.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a struct pointer")
	}
	vType := vElem.Type()
	re := regexp.MustCompile("(\\p{Lu}\\P{Lu}*)")
	for i := 0; i < vElem.NumField(); i++ {
		field := vElem.Field(i)
		if field.Kind() == reflect.Func {
			name := vType.Field(i).Tag.Get("name")
			if name == "" {
				name = prefix + strings.ToLower(re.ReplaceAllString(vType.Field(i).Name, "_${1}"))
			}
			sym, err := purego.Dlsym(lib, name)
			if err != nil {
				return err
			}
			purego.RegisterFunc(field.Addr().Interface(), sym)
		}
	}
	return nil
}

func getLibTarget() string {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return "x86_64-linux-gnu"
	case "linux/arm64":
		return "aarch64-linux-gnu"
	case "freebsd/amd64":
		return "x86_64-unknown-freebsd"
	case "freebsd/arm64":
		return "aarch64-unknown-freebsd"
	case "darwin/amd64":
		return "x86_64-apple-darwin"
	case "darwin/arm64":
		return "aarch64-apple-darwin"
	case "ios/amd64":
		return "x86_64-apple-ios"
	case "ios/arm64":
		return "aarch64-apple-ios"
	default:
		panic("unsupported architecture: " + runtime.GOARCH)
	}
}

func findSharedLib(target string, names []string) []string {
	var paths []string
	for _, name := range names {

		// 2.1.2 build library paths
		libDirs := []string{
			"/lib/" + target + "/",
			"/lib64/" + target + "/",
			"/usr/lib/" + target + "/",
			"/usr/lib64/" + target + "/",
			"/usr/local/lib64/" + target + "/",
			"/usr/local/lib/" + target + "/",
		}

		// 2.1.3 determine library path
		for _, libDir := range libDirs {
			if info, err := os.Stat(libDir); os.IsNotExist(err) || !info.IsDir() {
				continue
			}
			libPath := filepath.Join(libDir, "lib"+name+".so")
			if info, err := os.Stat(libPath); !os.IsNotExist(err) && !info.IsDir() {
				paths = append(paths, libPath)
				break
			}
			matches, err := filepath.Glob(filepath.Join(libDir, "lib"+name+".so") + ".*")
			if err != nil {
				continue
			}
			if len(matches) == 0 {
				continue
			}
			paths = append(paths, matches[0])
			break
		}
	}
	if len(paths) != len(names) {
		return nil
	}
	return paths
}

func (a *App) loadSharedLibs() error {
	a.log("loading shared libraries", "GOOS", runtime.GOOS, "GOARCH", runtime.GOARCH)
	loadTime := time.Now()

	// 1. Locate shared libraries
	target := getLibTarget()
	var libPaths []string
	for i, names := range libs {
		a.log("locating shared libraries", "target", target, "libs", names)
		paths := findSharedLib(target, names)
		if paths != nil {
			libPaths = paths
			lib.Version = i
			break
		}
	}
	if libPaths == nil {
		return fmt.Errorf("unable to locate shared libraries for %s", target)
	}

	// 2. Load shared libraries
	var err error
	a.log("loading gtk library", "path", libPaths[0])
	lib.GTK, err = purego.Dlopen(libPaths[0], purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("unable to load gtk library: %w", err)
	}
	a.log("loading webkit library", "path", libPaths[1])
	lib.Webkit, err = purego.Dlopen(libPaths[1], purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("unable to load webkit library: %w", err)
	}

	// 3. Register functions
	err = registerFunctions(lib.GTK, "g", &lib.g)
	if err != nil {
		return fmt.Errorf("unable to register g functions: %w", err)
	}

	err = registerFunctions(lib.GTK, "gdk", &lib.gdk)
	if err != nil {
		return fmt.Errorf("unable to register gdk functions: %w", err)
	}
	err = registerFunctions(lib.GTK, "gtk", &lib.gtk)
	if err != nil {
		return fmt.Errorf("unable to register gtk functions: %w", err)
	}
	err = registerFunctions(lib.Webkit, "jsc", &lib.jsc)
	if err != nil {
		return fmt.Errorf("unable to register jsc functions: %w", err)
	}
	err = registerFunctions(lib.Webkit, "webkit", &lib.webkit)
	if err != nil {
		a.log("unable to register webkit functions", "error", err)
	}
	err = registerFunctions(lib.Webkit, "webkit_settings", &lib.webkitSettings)
	if err != nil {
		a.log("unable to register webkit_settings functions", "error", err)
	}

	a.log("shared libraries loaded", "in", time.Since(loadTime), "paths", libPaths)
	return nil
}
