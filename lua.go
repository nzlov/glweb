package glweb

import (
	"io"
	"net/http"
)

type luaGlobal struct {
}

type luaResponseWriter struct {
	resp http.ResponseWriter
}

func (self luaResponseWriter) Add(k, v string) {
	self.resp.Header().Add(k, v)
}
func (self luaResponseWriter) Del(k string) {
	self.resp.Header().Del(k)
}
func (self luaResponseWriter) Get(k string) string {
	return self.resp.Header().Get(k)
}
func (self luaResponseWriter) Set(k, v string) {
	self.resp.Header().Set(k, v)
}

func (self luaResponseWriter) Write(w string) {
	io.WriteString(self.resp, w)
}
func (self luaResponseWriter) WriteHeader(i int) {
	self.resp.WriteHeader(i)
}
