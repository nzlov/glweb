package glweb

import (
	"crypto/sha256"
	"net/http"
	"os"
	"time"

	"github.com/nzlov/glweb/totp"

	log "github.com/nzlov/glog"
	"github.com/nzlov/glog/listener/console"
	"github.com/nzlov/glog/listener/file"
	//gvpass "github.com/nzlov/gvpass/v1"

	"github.com/nzlov/gluahttp"
)

var db *DB
var debugmode bool
var BasicRealm = "GLWEB API Authorization"
var totpo *totp.Options
var glhttp *gluahttp.HttpModule
var cors http.HandlerFunc

//var vpass *gvpass.GVPass

func init() {

	if len(os.Args) > 1 {
		debugmode = (os.Args[1] == "debug")

	}
	if debugmode {
		log.Register(&console.Console{})
	}
	f := "log/" + time.Now().Format("2006-01-02 150405") + ".log"
	logf, err := file.New(f)
	if err != nil {
		log.Errorln("Create Log file ", f)
		os.Exit(0)
	} else {
		log.Register(logf)
	}

	totpo = &totp.Options{
		Time:     time.Now,
		TimeStep: 30 * time.Second,
		Digits:   6,
		Hash:     sha256.New,
	}

	//	Publish("cmdline", Func(cmdline))
	//	Publish("memstats", Func(memstats))

	//	vpass = gvpass.New("http://112.253.7.54:1210/")

	glhttp = gluahttp.NewHttpModule(&http.Client{})
	cors = Allow(&Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"DELETE"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
}
