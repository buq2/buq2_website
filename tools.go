package main

import (
    "os"
    "path"
)

func executable() (string, error) {
    return os.Readlink("/proc/self/exe")
}

func executablePath() (string) {
    exe_path, _ := executable()
    return path.Dir(exe_path)
}
