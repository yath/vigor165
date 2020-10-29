// Dumps memory via the telnet CLI in a format similar to the UART console.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

var (
	dstAddr     = flag.String("dst_addr", "192.168.1.1:23", "Telnet host:port to connect to")
	username    = flag.String("username", "admin", "Telnet username")
	password    = flag.String("password", "admin", "Telnet password")
	startAddr   = flag.Uint("start_addr", 0x60000000, "Start address to dump")
	blockSize   = flag.Uint("block_size", 0x100, "Nuber of bytes to dump at once")
	length      = flag.Uint("length", 0x80000ff, "Number of bytes to dump in total")
	logIO       = flag.Bool("log_io", false, "Log telnet I/O")
	statusEvery = flag.Duration("status_every", 10*time.Second, "Status report frequency")
)

type loggingReadWriter struct {
	b io.ReadWriter
}

func (l *loggingReadWriter) Read(p []byte) (n int, err error) {
	n, err = l.b.Read(p)
	log.Printf("Read %d bytes (err = %v): %q", n, err, string(p[:n]))
	return n, err
}

func (l *loggingReadWriter) Write(p []byte) (n int, err error) {
	n, err = l.b.Write(p)
	log.Printf("Sent %d bytes (err = %v): %q", n, err, string(p[:n]))
	return n, err
}

const (
	usernamePrompt       = "Account:"
	passwordPrompt       = "Password:"
	cliPrompt            = "DrayTek>"
	debugModeCommand     = "sys admin drayteker"
	dumpMemCommandFormat = "sys mem %08x"
)

func expect(r io.Reader, s string) ([]byte, error) {
	var scnbuf []byte
	for {
		rdbuf := make([]byte, 1024)
		n, err := r.Read(rdbuf)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			scnbuf = append(scnbuf, rdbuf[:n]...)
			if bytes.Index(scnbuf, []byte(s)) != -1 {
				return scnbuf, nil
			}
		}
	}
}

func expectAndSend(rw io.ReadWriter, expct, snd string) ([]byte, error) {
	ret, err := expect(rw, expct)
	if err != nil {
		return nil, fmt.Errorf("while waiting for %q: %w", expct, err)
	}

	if _, err := fmt.Fprintln(rw, snd); err != nil {
		return nil, fmt.Errorf("can't send response to %q: %w", expct, err)
	}

	return ret, nil
}

func regexpForNBytes(n int) (*regexp.Regexp, error) {
	re := `(?m)^([[:xdigit:]]{8})[[:space:]]+`
	for i := 0; i < n; i++ {
		re = re + `([[:xdigit:]]{2})[[:space:]-]+`
	}
	re = re + `.*$`
	return regexp.Compile(re)
}

func main() {
	flag.Parse()

	var conn io.ReadWriter
	conn, err := net.Dial("tcp", *dstAddr)
	if err != nil {
		log.Fatalf("Can't connect to telnet interface: %v", err)
	}

	if *logIO {
		conn = &loggingReadWriter{conn}
	}

	if _, err := expectAndSend(conn, usernamePrompt, *username); err != nil {
		log.Fatalf("Can't send username: %v", err)
	}

	if _, err := expectAndSend(conn, passwordPrompt, *password); err != nil {
		log.Fatalf("Can't send password: %v", err)
	}

	if _, err := expectAndSend(conn, cliPrompt, debugModeCommand); err != nil {
		log.Fatalf("Can't enable debug mode: %v", err)
	}

	lineRE, err := regexpForNBytes(16)
	if err != nil {
		log.Panicf("Can't compile regex to match dump line: %v", err)
	}
	replaceStr := `$1: $2$3$4$5 $6$7$8$9 $10$11$12$13 $14$15$16$17`
	log.Printf("Replacing output lines matching %v with %q", lineRE, replaceStr)

	addr := *startAddr
	nextStatusAt := time.Now()
	for addr < *startAddr+*length {
		if time.Now().After(nextStatusAt) {
			done := float64(addr - *startAddr)
			donePct := done / float64(*length) * 100.0
			log.Printf("Dumping address 0x%08x (%.2f%%/%.2fMiB of %.2fMiB requested)", addr, donePct, done/1024.0/1024.0, float64(*length)/1024.0/1024.0)
			nextStatusAt = time.Now().Add(*statusEvery)
		}

		data, err := expectAndSend(conn, cliPrompt, fmt.Sprintf(dumpMemCommandFormat, addr))
		if err != nil {
			log.Fatalf("Can't send request to dump address 0x%08x: %v", err)
		}

		s := strings.ReplaceAll(string(data), "\r", "")
		s = lineRE.ReplaceAllString(s, replaceStr)
		fmt.Print(s)

		addr += *blockSize
	}
}
