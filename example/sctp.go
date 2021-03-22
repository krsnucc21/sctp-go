package main

import (
    "github.com/ishidawataru/sctp"
    "net"
    "net/http"
    "strings"
    "flag"
    "bytes"
    "math/rand"
    "time"
    "fmt"
    "os"
    "io/ioutil"
    "log"
)

func main() {
    var server = flag.Bool("server", false, "server")
    var ip = flag.String("ip", "0.0.0.0", "ip address")
    var port = flag.Int("port", 0, "port")
    var lport = flag.Int("lport", 0, "local port")

    var sndbuf = flag.Int("sndbuf", 0, "send buffer size")
    var rcvbuf = flag.Int("rcvbuf", 0, "receive buffer size")

    var prt = flag.Int("print", 2, "debug print")

    flag.Parse()

    ips := []net.IPAddr{}

    for _, i := range strings.Split(*ip, ",") {
        if a, err := net.ResolveIPAddr("ip", i); err == nil {
	    if *prt < 2 {
                log.Printf("resolved address '%s' to %s", i, a)
	    }
            ips = append(ips, *a)
        } else {
            log.Printf("error resolving address '%s': %v", i, err)
        }
    }

    addr := &sctp.SCTPAddr{
        IPAddrs: ips,
        Port:    *port,
    }

    if *prt < 2 {
        log.Printf("raw addr: %+v\n", addr.ToRawSockAddrBuf())
    }

    if *server {
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

	    if *prt == 0 {
                log.Printf("Accepted Connection from RemoteAddr: %s", conn.RemoteAddr())
                log.Printf("SndBufSize: %d, RcvBufSize: %d", *sndbuf, *rcvbuf)
	    }

            go serveClient(wconn, *prt)
        }
    } else {
        var laddr *sctp.SCTPAddr
        if *lport != 0 {
            laddr = &sctp.SCTPAddr{
                Port: *lport,
            }
        }
        conn, err := sctp.DialSCTP("sctp", laddr, addr)
        if err != nil {
            log.Fatalf("failed to dial: %v", err)
        }

	if *prt < 2 {
            log.Printf("dial localAddr: %s; remoteAddr: %s", conn.LocalAddr(), conn.RemoteAddr())
        }

        if *sndbuf != 0 {
            err = conn.SetWriteBuffer(*sndbuf)
            if err != nil {
                log.Fatalf("failed to set write buf: %v", err)
            }
        }
        if *rcvbuf != 0 {
            err = conn.SetReadBuffer(*rcvbuf)
            if err != nil {
                log.Fatalf("failed to set read buf: %v", err)
            }
        }

        *sndbuf, err = conn.GetWriteBuffer()
        if err != nil {
            log.Fatalf("failed to get write buf: %v", err)
        }
        *rcvbuf, err = conn.GetReadBuffer()
        if err != nil {
            log.Fatalf("failed to get read buf: %v", err)
        }

	if *prt < 2 {
            log.Printf("SndBufSize: %d, RcvBufSize: %d", *sndbuf, *rcvbuf)
	}

	rand.Seed(time.Now().UnixNano())

        ppid := 0
        for i := 0; i < 2; i++ {
            info := &sctp.SndRcvInfo{
                Stream: uint16(ppid),
                PPID:   uint32(ppid),
            }
            ppid += 1
            conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)

	    cell := rand.Uint32() % 256
	    user := rand.Uint32() % 1000
	    num := (rand.Uint32() % 99) + 1

	    var postString = fmt.Sprintf("{\"cellname\":\"%d\",\"username\":\"%d\",\"rsrp\":%d}", cell, user, num)
            buf := []byte(postString)

	    if *prt == 0 {
		fmt.Println(postString)
		fmt.Println(buf)
	    }

	    n, err := conn.SCTPWrite(buf, info)
            if err != nil {
                log.Fatalf("failed to write: %v", err)
            }

	    if *prt == 0 {
	        log.Printf("(%d) write: len %d", i, n)
	    }
        }
	time.Sleep(time.Second)
    }
}

func serveClient(conn net.Conn, prt int) error {
    httpposturl := "http://" + os.Getenv("LB_ADDR") + "/rsrp"
    i := 0
    for {
        buf := make([]byte, 256) // add overhead of SCTPSndRcvInfoWrappedConn
        n, err := conn.Read(buf)
        if err != nil || n <= 32 {
            log.Printf("read failed: %v (%d)", err, n)
            conn.Close()
            return err
        }

	if prt == 0 {
	    fmt.Printf("%d: %s\n", n, string(buf))
	    fmt.Println(buf)
	    fmt.Println(buf[32:n])
	}

	request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(buf[32:n]))
        request.Header.Set("Content-Type", "application/json; charset=UTF-8")

        client := &http.Client{}
        response, error := client.Do(request)
        if error != nil {
            log.Printf("write failed: %v", err)
            response.Body.Close()
            conn.Close()
	    return err
        }

        if prt == 0 || response.StatusCode != 200 {
            fmt.Println("HTTP JSON POST URL:", httpposturl)
            fmt.Println("response Headers:", response.Header)
            body, _ := ioutil.ReadAll(response.Body)
            fmt.Println("response Body:", string(body))
        }

        response.Body.Close()
        i++
    }
}
