//go:build windows

package main

import "testing"

func TestAccountLinkCodeInputIsBounded(t *testing.T){
 for _,tc:=range []struct{length uintptr;accept bool}{{0,false},{1,true},{21,true},{64,true},{65,false},{65536,false},{^uintptr(0),false}}{
  if got:=validAccountLinkInputLength(tc.length);got!=tc.accept {t.Fatalf("length %d: got %v, want %v",tc.length,got,tc.accept)}
 }
}
