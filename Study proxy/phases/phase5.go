package main

import (
    "log"
    "net/rpc"
    "sort"
    "sync"
)

var (
    requestBytes = make(map[string]int64)
    requestLock  = &sync.Mutex{}
)

type Empty struct{}
type Stats struct {
    RequestBytes map[string]int64
}
type RpcServer struct{}

type RequestStats struct {
    Path  string
    Bytes int64
}
type RequestStatsSlice []*RequestStats

func (r RequestStatsSlice) Less(i, j int) bool {
    return r[i].Bytes < r[j].Bytes
}

func (r RequestStatsSlice) Swap(i, j int) {
    r[i], r[j] = r[j], r[i]
}

func (r RequestStatsSlice) Len() int {
    return len(r)
}

func (r *RpcServer) GetStats(args *Empty, reply *Stats) error {
    requestLock.Lock()
    defer requestLock.Unlock()

    reply.RequestBytes = make(map[string]int64)
    for k, v := range requestBytes {
        reply.RequestBytes[k] = v
    }
    
    return nil
}

func main() {
    client, err := rpc.Dial("tcp", "127.0.0.1:8081")
    if err != nil {
        log.Fatalf("Failed to dial: %s", err)
    }

    var reply Stats
    err = client.Call("RpcServer.GetStats", &Empty{}, &reply)
    if err != nil {
        log.Fatalf("Failed to GetStats: %s", err)
    }

    rss := make(RequestStatsSlice, 0)
    for k, v := range reply.RequestBytes {
        rss = append(rss, &RequestStats{Path: k, Bytes: v})
    }
    sort.Sort(rss)

    for i := len(rss) - 1; i > len(rss)-10; i-- {
        log.Printf("%10d %s", rss[i].Bytes, rss[i].Path)
    }
}
