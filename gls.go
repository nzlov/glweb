package glweb

import (
	"io"
	"net/http"
	"sync"
	"time"

	log "github.com/nzlov/glog"
	"github.com/nzlov/go/gllog"
	"github.com/nzlov/go/glstr"

	"github.com/layeh/gopher-json"
	"github.com/layeh/gopher-luar"
	"github.com/yuin/gopher-lua"
)

type Task struct {
	Name    string
	Method  string
	Action  string
	Version string
	Fn      string
	Info    string
}

func (self *Task) exec(l *lua.LState, w http.ResponseWriter, r *http.Request, m map[string]string) {
	log.Infoln("Task", self.Name, self.Version, self.Method,
		self.Action, self.Fn, m, "Exec")

	if err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal(self.Fn),
		NRet:    3,
		Protect: true,
	}, luar.New(l, luaResponseWriter{w}), luar.New(l, r), luar.New(l, m)); err != nil {
		io.WriteString(w, "Error:"+err.Error())
		log.Errorln("Task", self.Name, self.Version, self.Method,
			self.Action, self.Fn, "Call", err)
	} else {
		log.Infoln("Task", self.Name, self.Version, self.Method,
			self.Action, self.Fn, m, "Exec OK")
	}
}

type Tasks struct {
	Name string
	Task []*Task
}

type glf struct {
	name      string
	path      string
	script    string
	tasks     map[string]*Task
	m         sync.Mutex
	lstates   []*lua.LState
	t         *time.Ticker
	isRunning bool
}

func newglf(p, script string) *glf {
	g := &glf{path: p, script: script}
	g.tasks = make(map[string]*Task)
	g.m = sync.Mutex{}
	g.lstates = make([]*lua.LState, 0)
	g.t = time.NewTicker(time.Second)
	g.isRunning = true
	go g.cleanl()
	return g
}

func (self *glf) cleanl() {
	for self.isRunning {
		select {
		case <-self.t.C:
			self.m.Lock()
			o := len(self.lstates)
			dr := self.lstates[0 : o/2]
			self.lstates = self.lstates[o/2 : o]
			self.m.Unlock()
			for _, l := range dr {
				l.Close()
			}
		}
	}
}

func (self *glf) addtask(t *Task) {
	self.tasks[t.Name] = t
}
func (self *glf) close() {
	self.m.Lock()
	self.isRunning = false
	for _, l := range self.lstates {
		l.Close()
	}
	self.m.Unlock()
}
func (self *glf) put(L *lua.LState) {
	self.m.Lock()
	L.SetTop(0)
	if self.isRunning {
		self.lstates = append(self.lstates, L)
	} else {
		L.Close()
	}
	self.m.Unlock()
}
func (self *glf) get() *lua.LState {
	if !self.isRunning {
		return nil
	}
	self.m.Lock()
	n := len(self.lstates)
	if n == 0 {
		self.m.Unlock()
		return self.new()
	}
	x := self.lstates[n-1]
	self.lstates = self.lstates[0 : n-1]
	self.m.Unlock()
	return x
}
func (self *glf) new() *lua.LState {
	l := lua.NewState()
	json.Preload(l)
	gllog.Preload(l)
	glstr.Preload(l)
	l.PreloadModule("http", glhttp.Loader)
	//	l.SetGlobal("VPASS", luar.New(l, vpass))
	if db != nil {
		l.SetGlobal("GLDB", luar.New(l, db))
	}
	if err := l.DoString(self.script); err != nil {
		log.Errorln("GLF", self.path, "DoString", err)
		return nil
	}
	return l
}

func (self *glf) exec(name string, w http.ResponseWriter, r *http.Request, m map[string]string) {
	if t, ok := self.tasks[name]; ok {
		log.Debugln("glf", "lstate", len(self.lstates))
		l := self.get()
		if l == nil {
			http.NotFound(w, r)
			return
		}
		t.exec(l, w, r, m)
		self.put(l)
	}
}
