package main

import (
    "fmt"
    "log"
    "net"
    "net/http"
    "net/rpc"
    "strings"
    "sync"
)

var IsDaemon bool
var DaemonConf string
var DaemonLock sync.Mutex

type Rpc struct {
}

func (r *Rpc) Call(args []string, reply *[]string) error {
    // lock the daemon so only one command is executed at the same time
    DaemonLock.Lock()
    defer DaemonLock.Unlock()

    cmd, ok := Command[args[0]]
    if !ok {
        log.Println("Invalid call:", args[0])
        return fmt.Errorf("Command does not exist: ", args[0])
    }

    log.Println("Calling command", args)
    WsConsole("Calling command " + strings.Join(args, " "))
    if err := cmd(args[1:], true, reply); err != nil {
        WsConsole("Command " + args[0] + " error: " + err.Error())
        return fmt.Errorf("Command ", args[0], " error: ", err)
    }
    for _, r := range *reply {
        WsConsole(" " + r)
        log.Println("", r)
    }

    if err := Config.Store(DaemonConf); err != nil {
        log.Fatal(err)
    }

    return nil
}

func init() {
    Command["daemon"] = DaemonCmd
    CommandHelp["daemon"] = "Run the daemon and listen for RPC, requires: [server|ip]:port"
}

func DaemonCmd(args []string, remote bool, output *[]string) error {
    if remote {
        return ErrNoRemoteCall
    }

    if len(args) < 1 {
        return fmt.Errorf("server/ip and port for RPC required")
    }

    IsDaemon = true

    r := new(Rpc)
    rpc.Register(r)
    rpc.HandleHTTP()
    l, e := net.Listen("tcp", args[0])
    if e != nil {
        return fmt.Errorf("listen error:", e)
    }
    log.Println("Listening for RPC on", l.Addr().String())
    AutomateAutostart()
    // Start listening for gRPC, this won't really return
    http.Serve(l, nil)

    return nil
}
