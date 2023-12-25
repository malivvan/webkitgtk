package webkitgtk

import (
	"fmt"
	"github.com/ebitengine/purego"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

var globalApplication *App

func init() {
	runtime.LockOSThread()
}

type App struct {
	options AppOptions
	pointer ptr
	pid     int
	ident   string
	logger  *logger

	windows     map[uint]*Window
	windowsLock sync.RWMutex

	runOnce runOnce
}

func New(options AppOptions) *App {
	if globalApplication != nil {
		return globalApplication
	} ///////////////////////////

	// Apply defaults
	if options.Name == "" {
		options.Name = "undefined"
	} else {
		options.Name = strings.ToLower(options.Name)
	}
	if options.PanicHandler == nil {
		options.PanicHandler = func(v any) {
			panic(v)
		}
	}

	// Create app
	app := &App{
		options: options,
		pid:     syscall.Getpid(),
		ident:   fmt.Sprintf("org.webkit2gtk.%s", strings.Replace(options.Name, " ", "-", -1)),
		windows: make(map[uint]*Window),
	}

	// Setup debug logger
	if options.Debug {
		logger := &logger{
			prefix: "webkit2gtk: " + options.Name,
			writer: os.Stdout,
		}
		if options.LogWriter != nil {
			logger.writer = options.LogWriter
		}
		app.logger = logger
	}

	globalApplication = app // !important

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

func (a *App) Quit() {
	appDestroy(a.pointer)
}

func (a *App) Run() error {
	defer processPanicHandlerRecover()
	startupTime := time.Now()
	a.log("application startup...", "identifier", a.ident, "main_thread", mainThreadId, "pid", a.pid)

	// 1. Fix console spam (USR1)
	if err := os.Setenv("JSC_SIGNAL_FOR_GC", "20"); err != nil {
		return err
	}
	//
	//// 2. Connect foreground signal (USR2)
	//// TODO: Is this working?
	//foreground := make(chan os.Signal, 1)
	//signal.Notify(foreground, syscall.SIGUSR2)
	//go func() {
	//	for {
	//		<-foreground
	//		for _, window := range a.windows {
	//			windowPresent(window.pointer)
	//
	//		}
	//		a.info("application foregrounded", "main_thread", mainThreadId)
	//	}
	//}()

	// 3. Load shared libraries
	if err := a.loadSharedLibs(); err != nil {
		return err
	}

	// 4. Get Main Thread and create GTK Application
	mainThreadId = lib.g.ThreadSelf()
	a.pointer = lib.gtk.ApplicationNew(a.ident, uint(0))

	//dbusCli, err := newDBUSClient(a.ident)
	//if err != nil {
	//	return err
	//}
	//running, pid := dbusCli.IsAppRunning()
	//if running {
	//	// send signal
	//	proc, err := os.FindProcess(int(pid))
	//	if err != nil {
	//		return err
	//	}
	//	err = proc.Signal(syscall.SIGUSR2)
	//	if err != nil {
	//		return err
	//	}
	//	a.info("application already running", "pid", pid)
	//	return fmt.Errorf("application already running")
	//}

	//lib.g.ApplicationRegister(a.pointer, 0, 0)

	// 3. Run deferred functions
	a.runOnce.invoke(true)

	// 4. Setup activate signal ipc
	app := ptr(a.pointer)
	activate := func() {
		a.log("application startup complete", "since_startup", time.Since(startupTime))
		lib.g.ApplicationHold(app) // allow running without a pointer
	}
	lib.g.SignalConnectData(
		ptr(a.pointer),
		"activate",
		purego.NewCallback(activate),
		app,
		false,
		0)

	// 5. Run GTK Application
	status := lib.g.ApplicationRun(a.pointer, 0, nil)
	/////////////////////////////////////////////////

	// 6. Shutdown
	shutdownTime := time.Now()
	a.log("application shutdown...", "status", status)
	lib.g.ApplicationRelease(a.pointer)
	lib.g.ObjectUnref(ptr(a.pointer))
	var err error
	if status != 0 {
		err = fmt.Errorf("exit code: %d", status)
	}
	a.log("application shutdown done", "since_shutdown", time.Since(shutdownTime))
	return err
}

func appDestroy(application ptr) {
	lib.g.ApplicationQuit(application)
}

func (a *App) log(msg interface{}, kv ...interface{}) {
	if a.logger == nil {
		return
	}
	a.logger.log(msg, kv...)
}

func fatal(message string, args ...interface{}) {
	println("*********************** FATAL ***********************")
	println(fmt.Sprintf(message, args...))
	println("*********************** FATAL ***********************")
	os.Exit(1)
}
