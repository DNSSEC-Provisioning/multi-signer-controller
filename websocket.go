package main

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 512
)

var (
    newline = []byte{'\n'}
    space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type Client struct {
    conn *websocket.Conn
    send chan []byte
}

var Clients map[string]*Client
var ClientsLock sync.Mutex

func init() {
    Clients = make(map[string]*Client)
}

type console struct {
    Log string `json:"log"`
}

func WsConsole(s string) {
    b, err := json.Marshal(&console{Log: s})
    if err != nil {
        log.Fatal(err)
    }
    ClientsLock.Lock()
    for _, c := range Clients {
        c.send <- b
    }
    ClientsLock.Unlock()
}

type signerStatus struct {
    Name    string `json:"name"`
    Leaving bool   `json:"leaving"`
}
type groupStatus struct {
    Fqdn    string         `json:"fqdn"`
    Stage   string         `json:"stage"`
    Signers []signerStatus `json:"signers"`
}

func WsStatus(fqdn, stage string, signers map[string]bool) {
    status := &groupStatus{Fqdn: fqdn, Stage: stage}
    for s, l := range signers {
        status.Signers = append(status.Signers, signerStatus{Name: s, Leaving: l})
    }
    b, err := json.Marshal(status)
    if err != nil {
        log.Fatal(err)
    }
    ClientsLock.Lock()
    for _, c := range Clients {
        c.send <- b
    }
    ClientsLock.Unlock()
}

type waitUntil struct {
    Fqdn string `json:"fqdn"`
    Left string `json:"left"`
}

func WsWaitUntil(fqdn, left string) {
    b, err := json.Marshal(&waitUntil{Fqdn: fqdn, Left: left})
    if err != nil {
        log.Fatal(err)
    }
    ClientsLock.Lock()
    for _, c := range Clients {
        c.send <- b
    }
    ClientsLock.Unlock()
}

func (c *Client) readPump() {
    defer func() {
        log.Println("lost websocket connection", c.conn.RemoteAddr().String())
        c.conn.Close()
        ClientsLock.Lock()
        delete(Clients, c.conn.RemoteAddr().String())
        ClientsLock.Unlock()
    }()
    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }
        message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
        // received messages are just ignored
        // c.hub.broadcast <- message
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write(newline)
                w.Write(<-c.send)
            }

            if err := w.Close(); err != nil {
                return
            }
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func serveWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    client := &Client{conn: conn, send: make(chan []byte, 256)}
    log.Println("new websocket connection from", conn.RemoteAddr().String())
    ClientsLock.Lock()
    Clients[conn.RemoteAddr().String()] = client
    ClientsLock.Unlock()

    go client.writePump()
    go client.readPump()

    DaemonLock.Lock()
    for _, g := range Config.ListGet("groups") {
        signers := make(map[string]bool)
        for _, s := range Config.ListGet("signers:" + g) {
            if Config.Get("signer-leaving:"+s, "") == "" {
                signers[s] = false
            } else {
                signers[s] = true
            }
        }
        WsStatus(g, Config.Get("automate-stage:"+g, ""), signers)
    }
    DaemonLock.Unlock()
}
