package main
import ("archive/zip"; "fmt"; "os"; "path/filepath"; "strings"; "testing")
func TestRejectZipWithExcessDirectoryEntries(t *testing.T) {
 file:=filepath.Join(t.TempDir(),"too-many-directories.zip")
 out,err:=os.Create(file);if err!=nil{t.Fatal(err)}
 zw:=zip.NewWriter(out)
 for i:=0;i<701;i++{if _,err=zw.Create(fmt.Sprintf("dir%03d/",i));err!=nil{t.Fatal(err)}}
 if err=zw.Close();err!=nil{t.Fatal(err)}
 if err=out.Close();err!=nil{t.Fatal(err)}
 if err=extractRuntime(file,filepath.Join(t.TempDir(),"runtime"));err==nil||!strings.Contains(err.Error(),"too many runtime archive entries"){t.Fatalf("directory header count unbounded: %v",err)}
}
func TestRejectZipWithTraversalDirectory(t *testing.T) {
 file:=filepath.Join(t.TempDir(),"unsafe-directory.zip")
 out,err:=os.Create(file);if err!=nil{t.Fatal(err)}
 zw:=zip.NewWriter(out)
 if _,err=zw.Create("../outside/");err!=nil{t.Fatal(err)}
 if _,err=zw.Create("llama-server.exe");err!=nil{t.Fatal(err)}
 if err=zw.Close();err!=nil{t.Fatal(err)}
 if err=out.Close();err!=nil{t.Fatal(err)}
 if err=extractRuntime(file,filepath.Join(t.TempDir(),"runtime"));err==nil||!strings.Contains(err.Error(),"unsafe runtime zip path"){t.Fatalf("unsafe directory header accepted: %v",err)}
}
