package main

import (
    "log"
    "sort"
)

func init() {
    Command["help"] = HelpCmd
}

func HelpCmd(args []string, remote bool, output *[]string) error {
    if remote {
        return ErrNoRemoteCall
    }

    log.Println("Available commands:")
    cmds := []string{}
    for k, _ := range CommandHelp {
        if k == "help" {
            continue
        }
        cmds = append(cmds, k)
    }
    sort.Strings(cmds)
    for _, k := range cmds {
        log.Printf("  %s: %s", k, CommandHelp[k])
    }
    log.Println("  help: Show this help and exit")

    return nil
}
