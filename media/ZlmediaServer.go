package main

import (
	"log"

	"syscall"
)

type MkApi struct {
	EnvInit2  *syscall.Proc
	IniCreate *syscall.Proc
}

func main() {
	dll, err := syscall.LoadDLL("mk_api.dll")
	if err != nil {
		log.Fatal(err)
	}
	lib := MkApi{}
	//查找函数
	lib.EnvInit2, err = dll.FindProc("mk_env_init2")
	if err != nil {
		log.Fatal(err)
	}
	lib.EnvInit2, err = dll.FindProc("mk_ini_create")
	if err != nil {
		log.Fatal(err)
	}

}
