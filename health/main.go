package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"
)

// https://github.com/esnet/iperf/blob/bd1437791a63579d589e9bea7de9250a876a5c97/src/iperf.h#L134
const COOKIESIZE = 37

const (
	PARAM_EXCHANGE uint8 = 9
	ACCESS_DENIED  uint8 = 0xff
)

// makeCookie generates a cookie that looks like the one iperf uses
// https://github.com/esnet/iperf/blob/98d87bd7e82b98775d9e4c62235132caa54233ab/src/iperf_util.c#L103-L125
func makeCookie() [COOKIESIZE]byte {
	const rndchars = "abcdefghijklmnopqrstuvwxyz234567"

	var cookie [COOKIESIZE]byte
	for i := 0; i < COOKIESIZE-1; i++ {
		cookie[i] = rndchars[rand.Intn(len(rndchars))]
	}
	cookie[COOKIESIZE-1] = 0
	return cookie
}

var (
	addr    = flag.String("address", "localhost:5201", "endpoint, default localhost:5201")
	timeout = flag.Duration("timeout", 5*time.Second, "connection, read, write, idle timeout")
)

func main() {
	// parse inputs
	flag.Parse()
	if addr == nil {
		fmt.Fprintln(os.Stderr, "invalid address")
		os.Exit(-2)
	}
	if timeout == nil || *timeout == 0 {
		timeout = new(time.Duration)
		*timeout = 5 * time.Second
	}

	// dial server
	conn, err := net.DialTimeout("tcp", *addr, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not dial: %v\n", err.Error())
		os.Exit(-2)
	}

	// send the cookie
	cookie := makeCookie()
	conn.SetDeadline(time.Now().Add(*timeout))
	_, err = conn.Write(cookie[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not write: %v\n", err.Error())
		os.Exit(-2)
	}

	// read reply from server
	conn.SetDeadline(time.Now().Add(*timeout))
	var reply bytes.Buffer
	read, err := io.CopyN(&reply, conn, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read: %v\n", err.Error())
		os.Exit(-2)
	}
	if read <= 0 {
		fmt.Fprintf(os.Stderr, "got no reply from server, %+v\n", reply)
		os.Exit(-2)
	}

	// examine reply
	switch reply.Bytes()[0] {
	case PARAM_EXCHANGE:
		fmt.Fprintf(os.Stdout, "server ready for params\n")
	case ACCESS_DENIED:
		fmt.Fprintf(os.Stdout, "server is busy running a test\n")
	default:
		fmt.Fprintf(os.Stderr, "unexpected state\n")
		os.Exit(1)
	}
}
