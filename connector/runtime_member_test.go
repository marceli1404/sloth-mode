package main

import (
 "strings"
 "testing"
)

func TestNativeRuntimeArchiveNameValidation(t *testing.T) {
 valid := map[string]string{
  "llama-server.exe": "llama-server.exe",
  "bin/llama-server.exe": "llama-server.exe",
  "./build/bin/libggml-base.so.0": "libggml-base.so.0",
  "lib/libllama.so": "libllama.so",
  "windows/ggml-cpu-x64.dll": "ggml-cpu-x64.dll",
 }
 for input,want := range valid {
  got,err:=safeRuntimeArchiveName(input)
  if err!=nil||got!=want {t.Errorf("valid %q: got %q %v",input,got,err)}
 }
 invalid := []string{
  "", "../private-notes", "bin/../../private-notes", "bin//llama-server.exe",
  "/etc/passwd", "C:/Windows/win.ini", "C:\\Windows\\win.ini",
  "bin\\llama-server.exe", "bin/key:payload", "CON", "CON.txt", "aux.DLL",
  "nul", "LPT1.so", "..", ".", "lib/special name.so", "lib/evil\nfile",
  "lib/llama-server.", "lib/evil☠.so", strings.Repeat("x",130),
 }
 for _,input:=range invalid {
  if out,err:=safeRuntimeArchiveName(input);err==nil {t.Errorf("unsafe archive entry accepted: %q -> %q",input,out)}
 }
}

func TestNativeRuntimePathConstructionRejectsUntrustedSelectors(t *testing.T) {
 base:=t.TempDir()
 for _,backend:=range []string{"", "../other", "../../tmp", "cpu/../../notes", "GPU", "vulkan/../cpu"} {
  if path:=runtimePath(base,backend);path!="" {
   t.Errorf("invalid backend %q produced native path %q",backend,path)
  }
 }
 for _,quant:=range []string{"", "../models", "Q4_K_M/../notes", "untrusted"} {
  if path:=modelPathFor(base,quant);path!="" {
   t.Errorf("invalid quantization %q produced path %q",quant,path)
  }
 }
 for _,backend:=range []string{"cpu","vulkan"} {
  if path:=runtimePath(base,backend);path=="" {
   t.Errorf("valid backend %q lost its runtime path",backend)
  }
 }
 if path:=modelPathFor(base,"Q4_K_M");path=="" {
  t.Fatal("valid model has no path")
 }
}
