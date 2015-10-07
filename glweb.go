package glweb

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"net/http/pprof"

	log "github.com/nzlov/glog"
	"github.com/nzlov/go/filemonitor"
	"github.com/nzlov/go/gllog"
	"github.com/nzlov/go/glstr"

	"github.com/layeh/gopher-json"
	"github.com/yuin/gluamapper"
	"github.com/yuin/gopher-lua"
)

type GLWeb struct {
	filemonitor.FileMonitorAdaptor
	glspath string

	fobserver *filemonitor.FileMonitorObserver
	fmonitor  *filemonitor.FileMonitor

	glsroute map[string]*route
	gls      map[string]*glf
	glps     map[string]string
}

func NewGLWeb(glspath string) *GLWeb {
	g := &GLWeb{}
	g.glspath = glspath
	return g
}
func (self *GLWeb) DB(db_user, db_pwd, db_host, db_database string) {
	db = &DB{}
	db.Open(db_user, db_pwd, db_host, db_database)
}
func (self *GLWeb) Run(hp string) {

	log.Infoln("GLWeb Running")
	log.SetLevel(log.DebugLevel)

	self.glsroute = make(map[string]*route)
	self.gls = make(map[string]*glf)
	self.glps = make(map[string]string)

	self.loadgls()
	self.fnotify()

	log.Infoln("GLWeb ListenAndServe...")
	go func() {
		if err := http.ListenAndServe(hp, self); err != nil {
			log.Errorln("GLWeb ListenAndServe Error.", err)
		}
	}()
	signalChan := make(chan os.Signal)               //创建一个信号量的chan，缓存为1，（0,1）意义不大
	signal.Notify(signalChan, os.Interrupt, os.Kill) //让进城收集信号量。
	<-signalChan
	log.Infoln("GLWeb Stop.")

	self.fmonitor.End()
	if db != nil {
		db.Close()
	}
	for _, g := range self.gls {
		g.close()
	}
	log.Infoln("GLWeb Close")
	log.Close()
}

func (self *GLWeb) loadgls() {
	gls, err := WalkDir(self.glspath, ".gl")
	log.Debugln("gls:", gls)
	if err != nil {
		log.Errorln("GLWeb loadgls  WalkDir Error.", err)
		return
	}

	for _, g := range gls {
		self.loadgl(g)
	}
}

func (self *GLWeb) loadgl(p string) {

	fbody, err := ioutil.ReadFile(p)
	if err != nil {
		log.Errorln("GLWeb loadgl ReadFile "+p+" Error.", err)
		return
	}
	L := lua.NewState()
	defer func() {
		L.SetTop(0)
		L.Close()
	}()

	json.Preload(L)
	gllog.Preload(L)
	glstr.Preload(L)
	if err := L.DoString(string(fbody)); err != nil {
		log.Errorln("GLWeb loadgl DoFile  Error.", err)
		return
	}

	ltask := L.GetGlobal("tasks")
	if ltask.Type() == lua.LTNil {
		log.Errorln("GLWeb loadgl Get Tasks  Error.", p, " not tasks table.")
		return
	}
	var tasks *Tasks
	if err := gluamapper.Map(L.GetGlobal("tasks").(*lua.LTable), &tasks); err != nil {
		log.Errorln("GLWeb loadgl Get Tasks  Error.", err)
		return
	}
	log.Infoln("GLWeb gls load ", p, tasks.Name)

	g := newglf(p, string(fbody))
	if tasks.Name == "" {
		log.Errorln("GLWEB gls load Error tasks name is nil")
		return
	}
	if _, ok := self.gls[tasks.Name]; ok {
		log.Errorln("GLWeb gls load Error tasks name ", tasks.Name, " exist")
		return
	}
	g.name = tasks.Name
	for _, t := range tasks.Task {
		//		log.WithFields(log.Fields{
		//			"name":     t.Name,
		//			"method":   t.Method,
		//			"action":   t.Action,
		//			"version":  t.Version,
		//			"function": t.Fn,
		//		}).Infoln("load task")

		if t.Name == "nil" || t.Method == "nil" || t.Action == "nil" ||
			t.Version == "nil" || t.Fn == "nil" {
			log.Errorln("GLWeb gls load Error name method action version fn is nil")
			continue
		}

		for n, r := range self.glsroute {
			if n == t.Name {
				log.Errorln("GLWeb gls load Error task name ", t.Name, " exist")
				continue
			}
			if r.pattern == t.Action {
				log.Errorln("GLWeb gls load Error action ", t.Action, " exist")
				continue
			}
		}
		self.glsroute[t.Name] = newRoute(g, t.Method, t.Action)
		g.addtask(t)
	}
	self.gls[p] = g
	self.glps[tasks.Name] = p
}

func (self *GLWeb) fnotify() {

	rootfile, err := filemonitor.NewFileEntry(self.glspath)
	log.Debugln("rootfile", rootfile.Path())
	if err != nil {
		log.Errorln("file Error:", err)
		return
	}
	fileFilter := filemonitor.NewFileFilter()
	fileFilter.AddFilter(".gl")
	self.fobserver = filemonitor.NewFileMonitorObserverByFileEntryAndFileFileter(
		rootfile,
		fileFilter,
	)
	self.fobserver.AddListener(self)
	self.fmonitor = filemonitor.NewFileMonitorByDt(self.fobserver, time.Second)
	self.fmonitor.Start()
	log.Infoln("GLWeb FileMonitor Start...")
}

func (self *GLWeb) FileCreate(file *filemonitor.FileEntry) {
	log.Infoln("FileCreate", file)

	self.loadgl(file.Path())
}

func (self *GLWeb) FileModify(file *filemonitor.FileEntry) {
	log.Infoln("FileModify", file)

	self.FileDelete(file)
	self.FileCreate(file)

}
func (self *GLWeb) FileDelete(file *filemonitor.FileEntry) {
	log.Infoln("FileDelete", file)

	if gls, ok := self.gls[file.Path()]; ok {
		for _, t := range gls.tasks {
			delete(self.glsroute, t.Name)
		}
		delete(self.gls, file.Path())
		delete(self.glps, gls.name)
		gls.close()
	}
}
func (self *GLWeb) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infoln("Request ", r.Method, r.URL.Path)
	w.Header().Add("Access-Control-Allow-Origin", "*")

	if r.Method == "GET" && r.URL.Path == "/" {
		if b := auth(w, r); b {
			htasks(self.gls, self.glps, w, r)
		}
		return
	}
	if debugmode && r.Method == "GET" {
		switch r.URL.Path {
		case "/debug/pprof/cmdline":
			if b := auth(w, r); b {
				pprof.Cmdline(w, r)
			}
			return
		case "/debug/pprof/profile":
			if b := auth(w, r); b {
				pprof.Profile(w, r)
			}
			return
		case "/debug/pprof/symbol":
			if b := auth(w, r); b {
				pprof.Symbol(w, r)
			}
			return
			//		case "/debug/vars":
			//			expvarHandler(w, r)
			//			return
		}
		if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
			if b := auth(w, r); b {
				pprof.Index(w, r)
			}
			return
		}
	}

	cors(w, r)

	for n, route := range self.glsroute {
		match, pmap := route.Match(r.Method, r.URL.Path)
		if match == NoMatch {
			continue
		}
		//		log.Infoln("Test", n, r.URL.Path, match, pmap, "OK")

		route.g.exec(n, w, r, pmap)
		return
	}
	http.NotFound(w, r)
}
