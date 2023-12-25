package webkitgtk

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

type logger struct {
	prefix string
	writer io.Writer
}

func (l *logger) log(msg interface{}, keyvals ...interface{}) {
	var s strings.Builder
	s.WriteString(l.prefix)
	s.WriteString(": ")
	s.WriteString(fmt.Sprintf("%v", msg))
	for i := 0; i < len(keyvals); i += 2 {
		s.WriteString(" ")
		s.WriteString(fmt.Sprintf("%v", keyvals[i]))
		s.WriteString("=")
		s.WriteString(fmt.Sprintf("%v", keyvals[i+1]))
	}
	s.WriteString("\n")
	l.writer.Write([]byte(s.String()))
}

type runOnce struct {
	mutex     sync.Mutex
	running   bool
	runnables []runnable
}

type runnable interface {
	run()
}

func (r *runOnce) add(runnable runnable) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.running {
		runnable.run()
	} else {
		r.runnables = append(r.runnables, runnable)
	}
}

func (r *runOnce) invoke(inRoutine bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.running = true
	for _, runnable := range r.runnables {
		if inRoutine {
			go runnable.run()
		} else {
			runnable.run()
		}
	}
	r.runnables = nil
}
