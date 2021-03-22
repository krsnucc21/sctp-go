package main

import (
    "github.com/ishidawataru/sctp"
    "net/http"
    "net/url"
    "encoding/json"
    "strings"
    "flags"
    "fmt"
    "os"
    "log"
)

func main() {
    var ip = flag.String("ip", "0.0.0.0", "")
    var port = flag.Int("port", 0, "")
    var lport = flag.Int("lport", 0, "")
    var bufsize = flag.Int("bufsize", 256, "")

    var sndbuf = flag.Int("sndbuf", 0, "")
    var rcvbuf = flag.Int("rcvbuf", 0, "")

    flag.Parse()

    ips := []net.IPAddr{}

    for _, i := range strings.Split(*ip, ",") {
        if a, err := net.ResolveIPAddr("ip", i); err == nil {
            log.Printf("Resolved address '%s' to %s", i, a)
            ips = append(ips, *a)
        } else {
            log.Printf("Error resolving address '%s': %v", i, err)
        }
    }

    addr := &sctp.SCTPAddr{
        IPAddrs: ips,
        Port:    *port,
    }
    log.Printf("raw addr: %+v\n", addr.ToRawSockAddrBuf())

    ln, err := sctp.ListenSCTP("sctp", addr)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    log.Printf("Listen on %s", ln.Addr())

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Fatalf("failed to accept: %v", err)
        }
        log.Printf("Accepted Connection from RemoteAddr: %s", conn.RemoteAddr())

        wconn := sctp.NewSCTPSndRcvInfoWrappedConn(conn.(*sctp.SCTPConn))
        if *sndbuf != 0 {
            err = wconn.SetWriteBuffer(*sndbuf)
            if err != nil {
                log.Fatalf("failed to set write buf: %v", err)
            }
        }
        if *rcvbuf != 0 {
            err = wconn.SetReadBuffer(*rcvbuf)
            if err != nil {
                log.Fatalf("failed to set read buf: %v", err)
            }
        }
        *sndbuf, err = wconn.GetWriteBuffer()
        if err != nil {
            log.Fatalf("failed to get write buf: %v", err)
        }
        *rcvbuf, err = wconn.GetWriteBuffer()
        if err != nil {
            log.Fatalf("failed to get read buf: %v", err)
        }
        log.Printf("SndBufSize: %d, RcvBufSize: %d", *sndbuf, *rcvbuf)

        go serveClient(wconn, *bufsize)
    }
}

func postMessage(msg string) {
    url := os.Getenv("LB_ADDR")
    resp, err := http.PostForm(url, data)

    if err != nil {
        log.Fatal(err)
    }

    var res map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&res)
    fmt.Println(res["form"])
}

func serveClient(conn net.Conn, bufsize int) error {
    i := 0
    for {
        buf := make([]byte, bufsize+128) // add overhead of SCTPSndRcvInfoWrappedConn
        n, err := conn.Read(buf)
        if err != nil {
            log.Printf("read failed: %v", err)
            conn.Close();
            return err
        }
        log.Printf("(%d) read: %d", i, n)
        n, err = postMessage(buf[:n])
        if err != nil {
            log.Printf("write failed: %v", err)
            conn.Close();
            return err
        }
        log.Printf("(%d) write: %d", i, n)
        i++
    }
}

