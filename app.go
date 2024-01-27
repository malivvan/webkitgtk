package webkitgtk

import (
	"fmt"
	"github.com/ebitengine/purego"
	"net/http"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func init() {
	runtime.LockOSThread()
}

var _app *App

type App struct {
	log logFunc

	id   string // application id e.g. com.github.malivvan.webkitgtk.undefined
	pid  int    // process id of the application
	name string // application name e.g. Unnamed Application
	icon []byte // application icon used to create desktop file

	thread  *mainThread // thread is the mainthread runner
	pointer ptr         // gtk application pointer

	trayIcon []byte    // trayIcon is the system tray icon (if not set icon will be used)
	trayMenu *TrayMenu // trayMenu is the system tray menu

	systray  *dbusSystray // systray is the dbus systray
	notifier *dbusNotify  // notifier is the dbus notifier
	session  *dbusSession // session is the dbus session

	windows     map[uint]*Window // windows is the map of all windows
	windowsLock sync.RWMutex     // windowsLock is the lock for windows map

	dialogs     map[uint]interface{} // dialogs is the map of all dialogs
	dialogsLock sync.RWMutex         // dialogsLock is the lock for dialogs map

	handler     map[string]http.Handler // handler is the map of all http handlers
	handlerLock sync.RWMutex            // handlerLock is the lock for handler map

	webContext   ptr                // webContext is the global webkit web context
	hold         bool               // hold indicates if the application stays alive after the last window is closed
	ephemeral    bool               // ephemeral is the flag to indicate if the application is ephemeral
	dataDir      string             // dataDir is the directory where the application data is stored
	cacheDir     string             // cacheDir is the directory where the application cache is stored
	cookiePolicy WebkitCookiePolicy // cookiePolicy is the cookie policy for the application
	cacheModel   WebkitCacheModel   // cacheModel is the cache model for the application

	started deferredRunner // started is the deferred runner for post application startup
}

func (a *App) Menu(icon []byte) *TrayMenu {
	a.trayIcon = icon
	if a.trayMenu == nil {
		a.trayMenu = &TrayMenu{}
	}
	return a.trayMenu
}

func (a *App) Handle(host string, handler http.Handler) {
	a.handlerLock.Lock()
	a.handler[host] = handler
	a.handlerLock.Unlock()
}

func New(options AppOptions) *App {
	if _app != nil {
		return _app
	} ///////////////////////////

	// Apply defaults
	if options.ID == "" {
		options.ID = "com.github.malivvan.webkitgtk.undefined"
	}
	if options.Name == "" {
		options.Name = "Unnamed Application"
	}
	if options.Icon == nil {
		options.Icon = defaultIcon
	}

	// Create app
	app := &App{
		log:  newLogFunc("app"),
		pid:  syscall.Getpid(),
		id:   options.ID,
		name: options.Name,
		icon: options.Icon,

		windows: make(map[uint]*Window),
		dialogs: make(map[uint]interface{}),
		handler: make(map[string]http.Handler),

		hold:         options.Hold,
		ephemeral:    options.Ephemeral,
		dataDir:      options.DataDir,
		cacheDir:     options.CacheDir,
		cookiePolicy: options.CookiePolicy,
		cacheModel:   options.CacheModel,
	}

	/////////////////////////////////////
	_app = app // !important
	return app
}

func (a *App) CurrentWindow() *Window {
	if a.pointer == 0 {
		return nil
	}
	active := lib.gtk.ApplicationGetActiveWindow(a.pointer)
	if active == 0 {
		return nil
	}
	a.windowsLock.RLock()
	windows := a.windows
	a.windowsLock.RUnlock()
	for _, w := range windows {
		if w.pointer == windowPtr(active) {
			return w
		}
	}
	return nil
}

func (a *App) Run() (err error) {
	defer panicHandlerRecover()

	// >>> STARTUP
	startupTime := time.Now()
	a.log("application startup...", "identifier", a.id, "pid", a.pid)

	// 1. Fix console spam (USR1)
	if err := os.Setenv("JSC_SIGNAL_FOR_GC", "20"); err != nil {
		return fmt.Errorf("failed to set JSC_SIGNAL_FOR_GC: %w", err)
	}

	// 2. Load shared libraries
	if err := a.loadSharedLibs(); err != nil {
		return fmt.Errorf("failed to load shared libraries: %w", err)
	}

	// 3. Validate application identifier
	if !lib.g.ApplicationIdIsValid(a.id) {
		return fmt.Errorf("invalid application identifier: %s", a.id)
	}

	// 4. Get Main Thread and create GTK Application
	a.thread = newMainThread()
	a.pointer = lib.gtk.ApplicationNew(a.id, uint(0))
	a.log("application created", "pointer", a.pointer, "thread", a.thread.ID())

	// 5. Establish DBUS session
	var dbusPlugins []dbusPlugin
	if a.trayMenu != nil {
		a.systray = a.trayMenu.toTray(a.id, a.trayIcon)
		dbusPlugins = append(dbusPlugins, a.systray)
	}
	a.notifier = &dbusNotify{
		appName: a.id,
	}
	dbusPlugins = append(dbusPlugins, a.notifier)
	a.session, err = newDBusSession(dbusPlugins)
	if err != nil {
		return fmt.Errorf("failed to create dbus session: %w", err)
	}

	// 5. Setup activate signal ipc
	lib.g.SignalConnectData(
		a.pointer,
		"activate",
		purego.NewCallback(func() {

			// 7. Allow running without a window
			lib.g.ApplicationHold(a.pointer)

			// 8. Invoke deferred runners
			a.started.invoke()

			// <<< STARTUP
			a.log("application startup complete", "since_startup", time.Since(startupTime))
		}),
		a.pointer,
		false,
		0)

	// 6. Run GTK Application
	status := lib.g.ApplicationRun(a.pointer, 0, nil) // BLOCKING

	// >>> SHUTDOWN
	shutdownTime := time.Now()
	a.log("application shutdown...", "status", status)

	// 1. Close dbus session
	a.session.close()

	// 2. Release GTK Application and dereference application pointer
	lib.g.ApplicationRelease(a.pointer)
	lib.g.ObjectUnref(a.pointer)

	// 3. Handle exit status
	if status == 0 {
		err = nil
	} else {
		err = fmt.Errorf("exit code: %d", status)
	}

	// <<< SHUTDOWN
	a.log("application shutdown done", "error", err, "since_shutdown", time.Since(shutdownTime))
	return err
}

func (a *App) Quit() {
	a.thread.InvokeSync(func() {
		lib.g.ApplicationQuit(a.pointer)
	})
}
