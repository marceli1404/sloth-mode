//go:build windows

package main

import (
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestProbeCompanionHealthAuthenticatesResponse(t *testing.T){
 cases:=[]struct{name string;status int;contentType,body string;want bool}{
  {"valid connector",200,"application/json; charset=utf-8",`{"ok":true,"name":"Sloth Mode Connector"}`,true},
  {"unrelated localhost JSON",200,"application/json",`{"ok":true,"name":"Other Program"}`,false},
  {"not healthy",200,"application/json",`{"ok":false,"name":"Sloth Mode Connector"}`,false},
  {"forbidden",403,"application/json",`{"ok":true,"name":"Sloth Mode Connector"}`,false},
  {"html fallback",200,"text/html",`<html>OK</html>`,false},
  {"bad JSON",200,"application/json",`{"ok":`,false},
 }
 for _,tc:=range cases {
  t.Run(tc.name,func(t *testing.T){
   ts:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
    if r.Header.Get("Origin")!="https://sloth.marceli.dev" || r.Method!=http.MethodGet {t.Error("missing required Sloth origin/method")}
    w.Header().Set("Content-Type",tc.contentType)
    w.WriteHeader(tc.status)
    _,_=w.Write([]byte(tc.body))
   }))
   defer ts.Close()
   if got:=probeCompanionHealth(ts.Client(),ts.URL+"/v1/health");got!=tc.want {t.Fatalf("ready=%v, want %v",got,tc.want)}
  })
 }
}
