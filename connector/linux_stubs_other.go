//go:build !linux
package main
import (
 "net/http"
 "errors"
)
func linuxCommand(_ string,_ string)(bool,error){return false,nil}
func linuxServe(_ *http.Server,_ chan struct{},_ string){}
func extractLinuxRuntime(_ string,_ string)error{return errors.New("Linux runtime extraction unavailable on this platform")}
