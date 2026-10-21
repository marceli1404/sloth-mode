package main

import (
 "errors"
 "path/filepath"
 "regexp"
 "strings"
)

// Validate every archive-derived filename BEFORE joining it with a writable
// destination. The native packages deliberately flatten archive members into
// a fresh private directory; nested source paths never become output paths.
// Reject Windows device names too, even when tested on Linux CI.
var runtimeFilename = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)

func safeRuntimeArchiveName(member string) (string, error) {
 if member == "" || strings.Contains(member, "\\") || strings.Contains(member, ":") || !filepath.IsLocal(filepath.FromSlash(member)) {
  return "", errors.New("unsafe runtime archive path")
 }
 for _, part := range strings.Split(member, "/") {
  if part == "" || part == ".." {
   return "", errors.New("unsafe runtime archive segment")
  }
 }
 name := filepath.Base(filepath.FromSlash(member))
 if name == "." || name == ".." || !runtimeFilename.MatchString(name) || strings.HasSuffix(name, ".") {
  return "", errors.New("unsafe runtime filename")
 }
 upper := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
 switch upper {
 case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
  return "", errors.New("reserved Windows runtime filename")
 }
 return name, nil
}
