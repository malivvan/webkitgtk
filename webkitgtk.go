package webkitgtk

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/ebitengine/purego"
	"image"
	"image/draw"
	"image/png"
	"os"
	"strconv"
	"strings"
	"sync"
)

var _RELEASE = false

const uriScheme = "app"

var PanicHandler = func(v any) {
	panic(v)
}

func panicHandlerRecover() {
	h := PanicHandler
	if h == nil {
		return
	}
	if err := recover(); err != nil {
		h(err)
	}
}

//go:embed examples/systray/icon.png
var defaultIcon []byte

type iconPX struct {
	W, H int
	Pix  []byte
}

func iconToPX(icon []byte) (iconPX, error) {
	img, err := pngToImage(icon)
	if err != nil {
		return iconPX{}, err
	}
	w, h, pix := imageToARGB(img)
	return iconPX{
		W:   w,
		H:   h,
		Pix: pix,
	}, nil
}
func pngToImage(data []byte) (*image.RGBA, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba, nil
}

func imageToARGB(img *image.RGBA) (int, int, []byte) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	data := make([]byte, w*h*4)
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			data[i] = byte(a)
			data[i+1] = byte(r)
			data[i+2] = byte(g)
			data[i+3] = byte(b)
			i += 4
		}
	}
	return w, h, data
}

type WebkitCookiePolicy int

const (
	CookiesAcceptAll WebkitCookiePolicy = iota
	CookiesRejectAll
	CookiesNoThirdParty
)

type WebkitCacheModel int

const (
	CacheNone WebkitCacheModel = iota
	CacheLite
	CacheFull
)

var LogWriter = os.Stderr

type logFunc func(msg interface{}, keyvals ...interface{})

func newLogFunc(prefix string) logFunc {
	return func(msg interface{}, keyvals ...interface{}) {
		if LogWriter == nil {
			return
		}
		var s strings.Builder
		s.WriteString(prefix)
		s.WriteString(": ")
		s.WriteString(fmt.Sprintf("%v", msg))
		for i := 0; i < len(keyvals); i += 2 {
			s.WriteString(" ")
			s.WriteString(fmt.Sprintf("%v", keyvals[i]))
			s.WriteString("=")
			s.WriteString(fmt.Sprintf("%v", keyvals[i+1]))
		}
		s.WriteString("\n")
		_, _ = LogWriter.Write([]byte(s.String()))
	}
}

type deferredRunner struct {
	mutex     sync.Mutex
	running   bool
	runnables []runnable
}

type runnable interface {
	run()
}

func (r *deferredRunner) run(runnable runnable) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.running {
		runnable.run()
	} else {
		r.runnables = append(r.runnables, runnable)
	}
}

func (r *deferredRunner) invoke() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.running = true
	for _, runnable := range r.runnables {
		go runnable.run()
	}
	r.runnables = nil
}

type mainThread struct {
	sync.Mutex
	id    uint64
	fnMap map[uint16]func()
}

func newMainThread() *mainThread {
	return &mainThread{
		id:    lib.g.ThreadSelf(),
		fnMap: make(map[uint16]func()),
	}
}

func (mt *mainThread) ID() uint64 {
	return mt.id

}

func (mt *mainThread) Running() bool {
	return mt.id == lib.g.ThreadSelf()
}

func (mt *mainThread) register(fn func()) uint16 {
	mt.Lock()
	defer mt.Unlock()

	var id uint16
	for {
		_, exist := mt.fnMap[id]
		if !exist {
			mt.fnMap[id] = fn
			return id
		}
		id++
		if id == 0 {
			panic("FATAL: Too many functions have been dispatched to the main thread")
			os.Exit(1)
		}
	}
}
func (mt *mainThread) dispatch(fn func()) {
	if mt.Running() {
		fn()
		return
	}
	id := mt.register(fn)
	lib.g.IdleAdd(purego.NewCallback(func(ptr) int {
		mt.Lock()
		fn, exist := mt.fnMap[id]
		if !exist {
			mt.Unlock()
			println("FATAL: main thread dispatch called with invalid id: " + strconv.Itoa(int(id)))
			os.Exit(1)
		}
		delete(mt.fnMap, id)
		mt.Unlock()
		fn()
		return 0 // gSourceRemove
	}))
}

func (mt *mainThread) InvokeSync(fn func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	mt.dispatch(func() {
		defer panicHandlerRecover()
		fn()
		wg.Done()
	})
	wg.Wait()
}

func (mt *mainThread) InvokeAsync(fn func()) {
	mt.dispatch(func() {
		defer panicHandlerRecover()
		fn()
	})
}

func (mt *mainThread) InvokeSyncWithResult(fn func() any) (res any) {
	var wg sync.WaitGroup
	wg.Add(1)
	mt.dispatch(func() {
		defer panicHandlerRecover()
		res = fn()
		wg.Done()
	})
	wg.Wait()
	return res
}

func (mt *mainThread) InvokeSyncWithError(fn func() error) (err error) {
	var wg sync.WaitGroup
	wg.Add(1)
	mt.dispatch(func() {
		defer panicHandlerRecover()
		err = fn()
		wg.Done()
	})
	wg.Wait()
	return
}

func (mt *mainThread) InvokeSyncWithResultAndError(fn func() (any, error)) (res any, err error) {
	var wg sync.WaitGroup
	wg.Add(1)
	mt.dispatch(func() {
		defer panicHandlerRecover()
		res, err = fn()
		wg.Done()
	})
	wg.Wait()
	return res, err
}
