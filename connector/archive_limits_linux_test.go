//go:build linux

package main
import ("archive/tar";"fmt";"path/filepath";"strings";"testing")
type archiveTestItem struct{name string;kind byte;link string;text string}
func TestRejectTarWithExcessDirectoryEntries(t *testing.T){
 items:=make([]struct{name string;kind byte;link string;text string},701)
 for i:=range items{items[i].name=fmt.Sprintf("folder%03d",i);items[i].kind=tar.TypeDir}
 archive:=writeTestTar(t,items)
 if err:=extractLinuxRuntime(archive,filepath.Join(t.TempDir(),"runtime"));err==nil||!strings.Contains(err.Error(),"too many archive entries"){t.Fatalf("TAR directories unbounded: %v",err)}
}
func TestRejectTarWithUnsafeDirectoryHeader(t *testing.T){
 archive:=writeTestTar(t,[]struct{name string;kind byte;link string;text string}{{name:"../outside",kind:tar.TypeDir},{name:"bin/llama-server",kind:tar.TypeReg,text:"binary"}})
 if err:=extractLinuxRuntime(archive,filepath.Join(t.TempDir(),"runtime"));err==nil||!strings.Contains(err.Error(),"unsafe archive member"){t.Fatalf("TAR traversal directory accepted: %v",err)}
}
func TestRejectTarCyclicRuntimeAliases(t *testing.T){
 archive:=writeTestTar(t,[]struct{name string;kind byte;link string;text string}{
 {name:"bin/llama-server",kind:tar.TypeReg,text:"binary"},
 {name:"lib/a.so",kind:tar.TypeSymlink,link:"b.so"},
 {name:"lib/b.so",kind:tar.TypeSymlink,link:"a.so"}})
 if err:=extractLinuxRuntime(archive,filepath.Join(t.TempDir(),"runtime"));err==nil||!strings.Contains(err.Error(),"cyclic"){t.Fatalf("TAR symlink cycle accepted: %v",err)}
}
