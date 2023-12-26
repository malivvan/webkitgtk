package webkitgtk

import (
	"net/http"
)

type AppOptions struct {

	// The name of the app.
	Name string

	// Debug mode.
	Debug bool

	// Hold the app open after the last window is closed.
	Hold bool

	// Handle internal app:// requests.
	Handle map[string]http.Handler

	// Ephemeral mode disables all persistent storage.
	Ephemeral bool

	// DataDir is the directory where persistent data is stored.
	DataDir string

	// CacheDir is the directory where persistent cache is stored.
	CacheDir string

	// CacheModel is the cache model used by the webview.
	CacheModel WebkitCacheModel

	// CookiePolicy is the cookie store used by the webview.
	CookiePolicy WebkitCookiePolicy
}

type WebkitSettings struct {
	EnableJavascript                          bool
	AutoLoadImages                            bool
	LoadIconsIgnoringImageLoadSetting         bool
	EnableOfflineWebApplicationCache          bool
	EnableHtml5LocalStorage                   bool
	EnableHtml5Database                       bool
	EnableXssAuditor                          bool
	EnableFrameFlattening                     bool
	EnablePlugins                             bool
	EnableJava                                bool
	JavascriptCanOpenWindowsAutomatically     bool
	EnableHyperlinkAuditing                   bool
	DefaultFontFamily                         string
	MonospaceFontFamily                       string
	SerifFontFamily                           string
	SansSerifFontFamily                       string
	CursiveFontFamily                         string
	FantasyFontFamily                         string
	PictographFontFamily                      string
	DefaultFontSize                           uint32
	DefaultMonospaceFontSize                  uint32
	MinimumFontSize                           uint32
	DefaultCharset                            string
	EnableDeveloperExtras                     bool
	EnableResizableTextAreas                  bool
	EnableTabsToLinks                         bool
	EnableDnsPrefetching                      bool
	EnableCaretBrowsing                       bool
	EnableFullscreen                          bool
	PrintBackgrounds                          bool
	EnableWebAudio                            bool
	EnableWebgl                               bool
	AllowModalDialogs                         bool
	ZoomTextOnly                              bool
	JavascriptCanAccessClipboard              bool
	MediaPlaybackRequiresUserGesture          bool
	MediaPlaybackAllowsInline                 bool
	DrawCompositingIndicators                 bool
	EnableSiteSpecificQuirks                  bool
	EnablePageCache                           bool
	UserAgent                                 string
	EnableSmoothScrolling                     bool
	EnableAccelerated2DCanvas                 bool
	EnableWriteConsoleMessagesToStdout        bool
	EnableMediaStream                         bool
	EnableMockCaptureDevices                  bool
	EnableSpatialNavigation                   bool
	EnableMediaSource                         bool
	EnableEncryptedMedia                      bool
	EnableMediaCapabilities                   bool
	AllowFileAccessFromFileUrls               bool
	AllowUniversalAccessFromFileUrls          bool
	AllowTopNavigationToDataUrls              bool
	HardwareAccelerationPolicy                int
	EnableBackForwardNavigationGestures       bool
	EnableJavascriptMarkup                    bool
	EnableMedia                               bool
	MediaContentTypesRequiringHardwareSupport string
	EnableWebRTC                              bool
	DisableWebSecurity                        bool
}

//	var webkitDefault = WebkitSettings{
//		EnableJavascript:                          true,
//		AutoLoadImages:                            true,
//		LoadIconsIgnoringImageLoadSetting:         false,
//		EnableOfflineWebApplicationCache:          true,
//		EnableHtml5LocalStorage:                   true,
//		EnableHtml5Database:                       true,
//		EnableXssAuditor:                          false,
//		EnableFrameFlattening:                     false,
//		EnablePlugins:                             false,
//		EnableJava:                                false,
//		JavascriptCanOpenWindowsAutomatically:     false,
//		EnableHyperlinkAuditing:                   true,
//		DefaultFontFamily:                         "sans-serif",
//		MonospaceFontFamily:                       "monospace",
//		SerifFontFamily:                           "serif",
//		SansSerifFontFamily:                       "sans-serif",
//		CursiveFontFamily:                         "serif",
//		FantasyFontFamily:                         "serif",
//		PictographFontFamily:                      "serif",
//		DefaultFontSize:                           16,
//		DefaultMonospaceFontSize:                  13,
//		MinimumFontSize:                           0,
//		DefaultCharset:                            "iso-8859-1",
//		EnablePrivateBrowsing:                     false,
//		EnableDeveloperExtras:                     false,
//		EnableResizableTextAreas:                  true,
//		EnableTabsToLinks:                         true,
//		EnableDnsPrefetching:                      false,
//		EnableCaretBrowsing:                       false,
//		EnableFullscreen:                          true,
//		PrintBackgrounds:                          true,
//		EnableWebAudio:                            true,
//		EnableWebgl:                               true,
//		AllowModalDialogs:                         false,
//		ZoomTextOnly:                              false,
//		JavascriptCanAccessClipboard:              false,
//		MediaPlaybackRequiresUserGesture:          false,
//		MediaPlaybackAllowsInline:                 true,
//		DrawCompositingIndicators:                 false,
//		EnableSiteSpecificQuirks:                  true,
//		EnablePageCache:                           true,
//		UserAgent:                                 "Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.4 Safari/605.1.15",
//		EnableSmoothScrolling:                     true,
//		EnableAccelerated2DCanvas:                 false,
//		EnableWriteConsoleMessagesToStdout:        false,
//		EnableMediaStream:                         true,
//		EnableMockCaptureDevices:                  false,
//		EnableSpatialNavigation:                   false,
//		EnableMediaSource:                         true,
//		EnableEncryptedMedia:                      false,
//		EnableMediaCapabilities:                   false,
//		AllowFileAccessFromFileUrls:               false,
//		AllowUniversalAccessFromFileUrls:          false,
//		AllowTopNavigationToDataUrls:              false,
//		HardwareAccelerationPolicy:                1,
//		EnableBackForwardNavigationGestures:       false,
//		EnableJavascriptMarkup:                    true,
//		EnableMedia:                               true,
//		MediaContentTypesRequiringHardwareSupport: "",
//		EnableWebRTC:                              false,
//		DisableWebSecurity:                        false,
//	}
var defaultWebkitSettings = WebkitSettings{
	EnableJavascript:                          true,
	AutoLoadImages:                            true,
	LoadIconsIgnoringImageLoadSetting:         false,
	EnableOfflineWebApplicationCache:          true,
	EnableHtml5LocalStorage:                   true,
	EnableHtml5Database:                       true,
	EnableXssAuditor:                          false,
	EnableFrameFlattening:                     false,
	EnablePlugins:                             false,
	EnableJava:                                false,
	JavascriptCanOpenWindowsAutomatically:     false,
	EnableHyperlinkAuditing:                   true,
	DefaultFontFamily:                         "sans-serif",
	MonospaceFontFamily:                       "monospace",
	SerifFontFamily:                           "serif",
	SansSerifFontFamily:                       "sans-serif",
	CursiveFontFamily:                         "serif",
	FantasyFontFamily:                         "serif",
	PictographFontFamily:                      "serif",
	DefaultFontSize:                           16,
	DefaultMonospaceFontSize:                  13,
	MinimumFontSize:                           0,
	DefaultCharset:                            "UTF-8",
	EnableDeveloperExtras:                     false,
	EnableResizableTextAreas:                  true,
	EnableTabsToLinks:                         true,
	EnableDnsPrefetching:                      false,
	EnableCaretBrowsing:                       false,
	EnableFullscreen:                          true,
	PrintBackgrounds:                          true,
	EnableWebAudio:                            true,
	EnableWebgl:                               true,
	AllowModalDialogs:                         false,
	ZoomTextOnly:                              false,
	JavascriptCanAccessClipboard:              false,
	MediaPlaybackRequiresUserGesture:          false,
	MediaPlaybackAllowsInline:                 true,
	DrawCompositingIndicators:                 false,
	EnableSiteSpecificQuirks:                  true,
	EnablePageCache:                           true,
	UserAgent:                                 "Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.4 Safari/605.1.15",
	EnableSmoothScrolling:                     true,
	EnableAccelerated2DCanvas:                 false,
	EnableWriteConsoleMessagesToStdout:        false,
	EnableMediaStream:                         true,
	EnableMockCaptureDevices:                  false,
	EnableSpatialNavigation:                   false,
	EnableMediaSource:                         true,
	EnableEncryptedMedia:                      false,
	EnableMediaCapabilities:                   false,
	AllowFileAccessFromFileUrls:               false,
	AllowUniversalAccessFromFileUrls:          false,
	AllowTopNavigationToDataUrls:              false,
	HardwareAccelerationPolicy:                1,
	EnableBackForwardNavigationGestures:       false,
	EnableJavascriptMarkup:                    true,
	EnableMedia:                               true,
	MediaContentTypesRequiringHardwareSupport: "",
	EnableWebRTC:                              false,
	DisableWebSecurity:                        false,
}

func (settings WebkitSettings) apply(settingsPtr webkitSettingsPtr) {
	lib.webkitSettings.SetEnableJavascript(settingsPtr, settings.EnableJavascript)
	lib.webkitSettings.SetAutoLoadImages(settingsPtr, settings.AutoLoadImages)
	lib.webkitSettings.SetLoadIconsIgnoringImageLoadSetting(settingsPtr, settings.LoadIconsIgnoringImageLoadSetting)
	lib.webkitSettings.SetEnableOfflineWebApplicationCache(settingsPtr, settings.EnableOfflineWebApplicationCache)
	lib.webkitSettings.SetEnableHtml5LocalStorage(settingsPtr, settings.EnableHtml5LocalStorage)
	lib.webkitSettings.SetEnableHtml5Database(settingsPtr, settings.EnableHtml5Database)
	lib.webkitSettings.SetEnableXssAuditor(settingsPtr, settings.EnableXssAuditor)
	lib.webkitSettings.SetEnableFrameFlattening(settingsPtr, settings.EnableFrameFlattening)
	lib.webkitSettings.SetEnablePlugins(settingsPtr, settings.EnablePlugins)
	lib.webkitSettings.SetEnableJava(settingsPtr, settings.EnableJava)
	lib.webkitSettings.SetJavascriptCanOpenWindowsAutomatically(settingsPtr, settings.JavascriptCanOpenWindowsAutomatically)
	lib.webkitSettings.SetEnableHyperlinkAuditing(settingsPtr, settings.EnableHyperlinkAuditing)
	lib.webkitSettings.SetDefaultFontFamily(settingsPtr, settings.DefaultFontFamily)
	lib.webkitSettings.SetMonospaceFontFamily(settingsPtr, settings.MonospaceFontFamily)
	lib.webkitSettings.SetSerifFontFamily(settingsPtr, settings.SerifFontFamily)
	lib.webkitSettings.SetSansSerifFontFamily(settingsPtr, settings.SansSerifFontFamily)
	lib.webkitSettings.SetCursiveFontFamily(settingsPtr, settings.CursiveFontFamily)
	lib.webkitSettings.SetFantasyFontFamily(settingsPtr, settings.FantasyFontFamily)
	lib.webkitSettings.SetPictographFontFamily(settingsPtr, settings.PictographFontFamily)
	lib.webkitSettings.SetDefaultFontSize(settingsPtr, settings.DefaultFontSize)
	lib.webkitSettings.SetDefaultMonospaceFontSize(settingsPtr, settings.DefaultMonospaceFontSize)
	lib.webkitSettings.SetMinimumFontSize(settingsPtr, settings.MinimumFontSize)
	lib.webkitSettings.SetDefaultCharset(settingsPtr, settings.DefaultCharset)
	lib.webkitSettings.SetEnableDeveloperExtras(settingsPtr, settings.EnableDeveloperExtras)
	lib.webkitSettings.SetEnableResizableTextAreas(settingsPtr, settings.EnableResizableTextAreas)
	lib.webkitSettings.SetEnableTabsToLinks(settingsPtr, settings.EnableTabsToLinks)
	lib.webkitSettings.SetEnableDnsPrefetching(settingsPtr, settings.EnableDnsPrefetching)
	lib.webkitSettings.SetEnableCaretBrowsing(settingsPtr, settings.EnableCaretBrowsing)
	lib.webkitSettings.SetEnableFullscreen(settingsPtr, settings.EnableFullscreen)
	lib.webkitSettings.SetPrintBackgrounds(settingsPtr, settings.PrintBackgrounds)
	lib.webkitSettings.SetEnableWebaudio(settingsPtr, settings.EnableWebAudio)
	lib.webkitSettings.SetEnableWebgl(settingsPtr, settings.EnableWebgl)
	lib.webkitSettings.SetAllowModalDialogs(settingsPtr, settings.AllowModalDialogs)
	lib.webkitSettings.SetZoomTextOnly(settingsPtr, settings.ZoomTextOnly)
	lib.webkitSettings.SetJavascriptCanAccessClipboard(settingsPtr, settings.JavascriptCanAccessClipboard)
	lib.webkitSettings.SetMediaPlaybackRequiresUserGesture(settingsPtr, settings.MediaPlaybackRequiresUserGesture)
	lib.webkitSettings.SetMediaPlaybackAllowsInline(settingsPtr, settings.MediaPlaybackAllowsInline)
	lib.webkitSettings.SetDrawCompositingIndicators(settingsPtr, settings.DrawCompositingIndicators)
	lib.webkitSettings.SetEnableSiteSpecificQuirks(settingsPtr, settings.EnableSiteSpecificQuirks)
	lib.webkitSettings.SetEnablePageCache(settingsPtr, settings.EnablePageCache)
	lib.webkitSettings.SetUserAgent(settingsPtr, settings.UserAgent)
	lib.webkitSettings.SetEnableSmoothScrolling(settingsPtr, settings.EnableSmoothScrolling)
	lib.webkitSettings.SetEnableAccelerated2DCanvas(settingsPtr, settings.EnableAccelerated2DCanvas)
	lib.webkitSettings.SetEnableWriteConsoleMessagesToStdout(settingsPtr, settings.EnableWriteConsoleMessagesToStdout)
	lib.webkitSettings.SetEnableMediaStream(settingsPtr, settings.EnableMediaStream)
	lib.webkitSettings.SetEnableMockCaptureDevices(settingsPtr, settings.EnableMockCaptureDevices)
	lib.webkitSettings.SetEnableSpatialNavigation(settingsPtr, settings.EnableSpatialNavigation)
	lib.webkitSettings.SetEnableMediasource(settingsPtr, settings.EnableMediaSource)
	lib.webkitSettings.SetEnableEncryptedMedia(settingsPtr, settings.EnableEncryptedMedia)
	lib.webkitSettings.SetEnableMediaCapabilities(settingsPtr, settings.EnableMediaCapabilities)
	lib.webkitSettings.SetAllowFileAccessFromFileUrls(settingsPtr, settings.AllowFileAccessFromFileUrls)
	lib.webkitSettings.SetAllowUniversalAccessFromFileUrls(settingsPtr, settings.AllowUniversalAccessFromFileUrls)
	lib.webkitSettings.SetAllowTopNavigationToDataUrls(settingsPtr, settings.AllowTopNavigationToDataUrls)
	lib.webkitSettings.SetHardwareAccelerationPolicy(settingsPtr, settings.HardwareAccelerationPolicy)
	lib.webkitSettings.SetEnableBackForwardNavigationGestures(settingsPtr, settings.EnableBackForwardNavigationGestures)
	lib.webkitSettings.SetEnableJavascriptMarkup(settingsPtr, settings.EnableJavascriptMarkup)
	lib.webkitSettings.SetEnableMedia(settingsPtr, settings.EnableMedia)
	lib.webkitSettings.SetMediaContentTypesRequiringHardwareSupport(settingsPtr, settings.MediaContentTypesRequiringHardwareSupport)
	lib.webkitSettings.SetEnableWebrtc(settingsPtr, settings.EnableWebRTC)
	lib.webkitSettings.SetDisableWebSecurity(settingsPtr, settings.DisableWebSecurity)
}

func toWebkitSettings(settingsPtr webkitSettingsPtr) WebkitSettings {
	var settings WebkitSettings
	settings.EnableJavascript = lib.webkitSettings.GetEnableJavascript(settingsPtr)
	settings.AutoLoadImages = lib.webkitSettings.GetAutoLoadImages(settingsPtr)
	settings.LoadIconsIgnoringImageLoadSetting = lib.webkitSettings.GetLoadIconsIgnoringImageLoadSetting(settingsPtr)
	settings.EnableOfflineWebApplicationCache = lib.webkitSettings.GetEnableOfflineWebApplicationCache(settingsPtr)
	settings.EnableHtml5LocalStorage = lib.webkitSettings.GetEnableHtml5LocalStorage(settingsPtr)
	settings.EnableHtml5Database = lib.webkitSettings.GetEnableHtml5Database(settingsPtr)
	settings.EnableXssAuditor = lib.webkitSettings.GetEnableXssAuditor(settingsPtr)
	settings.EnableFrameFlattening = lib.webkitSettings.GetEnableFrameFlattening(settingsPtr)
	settings.EnablePlugins = lib.webkitSettings.GetEnablePlugins(settingsPtr)
	settings.EnableJava = lib.webkitSettings.GetEnableJava(settingsPtr)
	settings.JavascriptCanOpenWindowsAutomatically = lib.webkitSettings.GetJavascriptCanOpenWindowsAutomatically(settingsPtr)
	settings.EnableHyperlinkAuditing = lib.webkitSettings.GetEnableHyperlinkAuditing(settingsPtr)
	settings.DefaultFontFamily = lib.webkitSettings.GetDefaultFontFamily(settingsPtr)
	settings.MonospaceFontFamily = lib.webkitSettings.GetMonospaceFontFamily(settingsPtr)
	settings.SerifFontFamily = lib.webkitSettings.GetSerifFontFamily(settingsPtr)
	settings.SansSerifFontFamily = lib.webkitSettings.GetSansSerifFontFamily(settingsPtr)
	settings.CursiveFontFamily = lib.webkitSettings.GetCursiveFontFamily(settingsPtr)
	settings.FantasyFontFamily = lib.webkitSettings.GetFantasyFontFamily(settingsPtr)
	settings.PictographFontFamily = lib.webkitSettings.GetPictographFontFamily(settingsPtr)
	settings.DefaultFontSize = lib.webkitSettings.GetDefaultFontSize(settingsPtr)
	settings.DefaultMonospaceFontSize = lib.webkitSettings.GetDefaultMonospaceFontSize(settingsPtr)
	settings.MinimumFontSize = lib.webkitSettings.GetMinimumFontSize(settingsPtr)
	settings.DefaultCharset = lib.webkitSettings.GetDefaultCharset(settingsPtr)
	settings.EnableDeveloperExtras = lib.webkitSettings.GetEnableDeveloperExtras(settingsPtr)
	settings.EnableResizableTextAreas = lib.webkitSettings.GetEnableResizableTextAreas(settingsPtr)
	settings.EnableTabsToLinks = lib.webkitSettings.GetEnableTabsToLinks(settingsPtr)
	settings.EnableDnsPrefetching = lib.webkitSettings.GetEnableDnsPrefetching(settingsPtr)
	settings.EnableCaretBrowsing = lib.webkitSettings.GetEnableCaretBrowsing(settingsPtr)
	settings.EnableFullscreen = lib.webkitSettings.GetEnableFullscreen(settingsPtr)
	settings.PrintBackgrounds = lib.webkitSettings.GetPrintBackgrounds(settingsPtr)
	settings.EnableWebAudio = lib.webkitSettings.GetEnableWebaudio(settingsPtr)
	settings.EnableWebgl = lib.webkitSettings.GetEnableWebgl(settingsPtr)
	settings.AllowModalDialogs = lib.webkitSettings.GetAllowModalDialogs(settingsPtr)
	settings.ZoomTextOnly = lib.webkitSettings.GetZoomTextOnly(settingsPtr)
	settings.JavascriptCanAccessClipboard = lib.webkitSettings.GetJavascriptCanAccessClipboard(settingsPtr)
	settings.MediaPlaybackRequiresUserGesture = lib.webkitSettings.GetMediaPlaybackRequiresUserGesture(settingsPtr)
	settings.MediaPlaybackAllowsInline = lib.webkitSettings.GetMediaPlaybackAllowsInline(settingsPtr)
	settings.DrawCompositingIndicators = lib.webkitSettings.GetDrawCompositingIndicators(settingsPtr)
	settings.EnableSiteSpecificQuirks = lib.webkitSettings.GetEnableSiteSpecificQuirks(settingsPtr)
	settings.EnablePageCache = lib.webkitSettings.GetEnablePageCache(settingsPtr)
	settings.UserAgent = lib.webkitSettings.GetUserAgent(settingsPtr)
	settings.EnableSmoothScrolling = lib.webkitSettings.GetEnableSmoothScrolling(settingsPtr)
	settings.EnableAccelerated2DCanvas = lib.webkitSettings.GetEnableAccelerated2DCanvas(settingsPtr)
	settings.EnableWriteConsoleMessagesToStdout = lib.webkitSettings.GetEnableWriteConsoleMessagesToStdout(settingsPtr)
	settings.EnableMediaStream = lib.webkitSettings.GetEnableMediaStream(settingsPtr)
	settings.EnableMockCaptureDevices = lib.webkitSettings.GetEnableMockCaptureDevices(settingsPtr)
	settings.EnableSpatialNavigation = lib.webkitSettings.GetEnableSpatialNavigation(settingsPtr)
	settings.EnableMediaSource = lib.webkitSettings.GetEnableMediasource(settingsPtr)
	settings.EnableEncryptedMedia = lib.webkitSettings.GetEnableEncryptedMedia(settingsPtr)
	settings.EnableMediaCapabilities = lib.webkitSettings.GetEnableMediaCapabilities(settingsPtr)
	settings.AllowFileAccessFromFileUrls = lib.webkitSettings.GetAllowFileAccessFromFileUrls(settingsPtr)
	settings.AllowUniversalAccessFromFileUrls = lib.webkitSettings.GetAllowUniversalAccessFromFileUrls(settingsPtr)
	settings.AllowTopNavigationToDataUrls = lib.webkitSettings.GetAllowTopNavigationToDataUrls(settingsPtr)
	settings.HardwareAccelerationPolicy = lib.webkitSettings.GetHardwareAccelerationPolicy(settingsPtr)
	settings.EnableBackForwardNavigationGestures = lib.webkitSettings.GetEnableBackForwardNavigationGestures(settingsPtr)
	settings.EnableJavascriptMarkup = lib.webkitSettings.GetEnableJavascriptMarkup(settingsPtr)
	settings.EnableMedia = lib.webkitSettings.GetEnableMedia(settingsPtr)
	settings.MediaContentTypesRequiringHardwareSupport = lib.webkitSettings.GetMediaContentTypesRequiringHardwareSupport(settingsPtr)
	settings.EnableWebRTC = lib.webkitSettings.GetEnableWebrtc(settingsPtr)
	settings.DisableWebSecurity = lib.webkitSettings.GetDisableWebSecurity(settingsPtr)
	return settings
}
