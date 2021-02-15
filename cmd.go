package main

import (
    "fmt"
)

type CmdFunc func(args []string, remote bool, output *[]string) error

var Command = make(map[string]CmdFunc)
var CommandHelp = make(map[string]string)

var ErrNoRemoteCall = fmt.Errorf("Can not be called remotely")
