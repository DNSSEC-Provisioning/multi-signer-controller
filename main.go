package main

import (
    "flag"
    "log"
    "net/http"
    "net/rpc"
    "os"
    "os/signal"
    "runtime"
    "runtime/pprof"
    "time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memstats = flag.Bool("memstats", false, "output memstats")
var conf = flag.String("conf", "", "config file to use")
var remote = flag.String("remote", "", "specify remote daemon to execute commands on [<server|ip>]:<port>")
var httpAddr = flag.String("http", "", "http service address")

func main() {
    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }
    if *memstats == true {
        go func() {
            m := &runtime.MemStats{}
            for {
                runtime.ReadMemStats(m)
                log.Printf("Alloc %v / %v   Live %v   Heap %v / %v / %v / %v   Objs %v   Next %v",
                    m.Alloc, m.Sys,
                    m.Mallocs-m.Frees,
                    m.HeapAlloc, m.HeapSys, m.HeapIdle, m.HeapInuse,
                    m.HeapObjects,
                    m.NextGC,
                )
                time.Sleep(1 * time.Second)
            }
        }()
    }
    if e := run(); e != 0 {
        os.Exit(e)
    }
    if *memstats == true {
        time.Sleep(10 * time.Second)
        m := &runtime.MemStats{}
        runtime.ReadMemStats(m)
        log.Printf("Alloc %v / %v   Live %v   Heap %v / %v / %v / %v   Objs %v   Next %v",
            m.Alloc, m.Sys,
            m.Mallocs-m.Frees,
            m.HeapAlloc, m.HeapSys, m.HeapIdle, m.HeapInuse,
            m.HeapObjects,
            m.NextGC,
        )
    }
}

func serveHome(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }
    if r.Method != "GET" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    http.ServeFile(w, r, "home.html")
}

func run() int {
    args := flag.Args()

    if len(args) < 1 {
        log.Println("At least one command must be given")
        HelpCmd([]string{}, false, nil)
        return 1
    }

    if args[0] == "help" {
        HelpCmd([]string{}, false, nil)
        return 0
    }

    if *httpAddr != "" {
        go func() {
            http.HandleFunc("/", serveHome)
            http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
                serveWs(w, r)
            })
            log.Println("HTTP on", *httpAddr)
            err := http.ListenAndServe(*httpAddr, nil)
            if err != nil {
                log.Fatal("ListenAndServe: ", err)
            }
        }()
    }

    if *remote != "" {
        client, err := rpc.DialHTTP("tcp", *remote)
        if err != nil {
            log.Fatal("dialing:", err)
        }
        defer client.Close()

        var reply []string
        err = client.Call("Rpc.Call", args, &reply)
        if err != nil {
            log.Fatal("call error:", err.Error())
        }

        for _, v := range reply {
            log.Println(v)
        }
        return 0
    }

    if *conf == "" {
        log.Fatal("-conf <file> must be specified")
    }
    if _, err := os.Stat(*conf); !os.IsNotExist(err) {
        log.Println("Loading config", *conf)
        if err := Config.Load(*conf); err != nil {
            log.Fatal(err)
        }
    }
    // Daemon needs to know what config is used, it will save changes after each command
    DaemonConf = *conf

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        for range c {
            log.Println("Caught SIGINT")

            if err := Config.Store(*conf); err != nil {
                log.Fatal(err)
            }

            os.Exit(1)
        }
    }()

    cmd, ok := Command[args[0]]
    if !ok {
        log.Fatal("Command does not exist: ", args[0])
    }

    var out []string
    if err := cmd(args[1:], false, &out); err != nil {
        log.Fatal("Command ", args[0], " error: ", err)
    }

    for _, v := range out {
        log.Println(v)
    }

    if err := Config.Store(*conf); err != nil {
        log.Fatal(err)
    }

    return 0
}
