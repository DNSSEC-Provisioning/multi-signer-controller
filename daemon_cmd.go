package main

import (
    "fmt"
    "log"
    "net"
    "net/http"
    "net/rpc"
)

type Rpc struct {
}

func (r *Rpc) Call(args []string, reply *[]string) error {
    cmd, ok := Command[args[0]]
    if !ok {
        log.Println("Invalid call:", args[0])
        return fmt.Errorf("Command does not exist: ", args[0])
    }

    log.Println("Calling command", args[0])
    if err := cmd(args[1:], true, reply); err != nil {
        return fmt.Errorf("Command ", args[0], " error: ", err)
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

    r := new(Rpc)
    rpc.Register(r)
    rpc.HandleHTTP()
    l, e := net.Listen("tcp", args[0])
    if e != nil {
        return fmt.Errorf("listen error:", e)
    }
    log.Println("Listening for RPC on", l.Addr().String())
    http.Serve(l, nil)

    return nil
}
