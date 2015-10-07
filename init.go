package glweb

import (
	"crypto/sha256"
	"net/http"
	"time"

	"github.com/nzlov/glweb/totp"

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
