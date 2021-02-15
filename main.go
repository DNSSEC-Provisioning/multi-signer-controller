package main

import (
    "flag"
    "log"
    "os"
    "runtime"
    "runtime/pprof"
    "time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memstats = flag.Bool("memstats", false, "output memstats")

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

func run() int {
    return 0
}
