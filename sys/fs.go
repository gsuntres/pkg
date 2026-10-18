package sys

import (
  "log"
  "io"
  "os"
  "errors"
  "runtime"
  "path/filepath"
)

func Rootify(path string, rootPath... string) string {
  var root string
  withRoot := false
  if len(rootPath) > 0 {
    root = rootPath[0]
    withRoot = true
  }
  
  var err error

  if !withRoot {
    root, err = GetProjectRoot(1)
    if err != nil {
      log.Fatal(err)
    }  
  }
  
  return filepath.Join(root, path)
}

func GetProjectRoot(caller int) (string, error) {
  _, filename, _, ok := runtime.Caller(caller)
  if !ok {
    return "", errors.New("could not determine source file")
  }

  return filepath.Dir(filename), nil
}

func GetDir(caller int) (string, error) {
  _, filename, _, ok := runtime.Caller(caller)
  if !ok {
    return "", errors.New("could not determine source file")
  }

  return filepath.Dir(filename), nil
}

// func GetProjectRoot(dir... bool) (string, error) {
//   exeDir, err := os.Executable()
//   if err != nil {
//     return "", err
//   }

//   exeDir = filepath.Dir(exeDir)

//   for i, dir := range dir {
//     if dir {
//       exeDir = filepath.Dir(exeDir)
//     }

//     // up to 3 
//     if i > 3 {
//       break
//     }
//   }

//   return exeDir, nil
// }

func CopyFile(src, dst string) error {
  in, err := os.Open(src)
  if err != nil {
    return err
  }
  defer in.Close()

  out, err := os.Create(dst)
  if err != nil {
    return err
  }
  defer out.Close()

  _, err = io.Copy(out, in)
  if err != nil {
    return err
  }

  return out.Sync()
}

