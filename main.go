package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	if strings.HasPrefix(cmd, "l") {
		if err := listen(flag.Arg(1)); err != nil {
			log.Printf("Error listening: %v", err)
			os.Exit(2)
		}
	} else if strings.HasPrefix(cmd, "c") {
		if err := connect(flag.Arg(1)); err != nil {
			log.Printf("error connecting: %v", err)
			os.Exit(2)
		}
	} else {
		log.Printf("unknown command: %q", cmd)
		os.Exit(1)
	}
}

// Protocol:
// read 1 byte as n; read next n bytes; sleep; reply with a "1" (\x31); repeat
func listen(fn string) error {
	usdln, err := net.Listen("unix", fn)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Printf("closing listener due to interrupt...")
		if err := usdln.Close(); err != nil {
			log.Printf("close error: %v", err)
		}
	}()

	for {
		conn, err := usdln.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %w", err)
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing connection %s: (%T) %v", conn.RemoteAddr(), err, err)
		}
	}()

	buf := make([]byte, 256)
	for {
		if _, err := conn.Read(buf[:1]); err != nil {
			log.Printf("error reading initial byte from connection %s: (%T) %v", conn.RemoteAddr(), err, err)
			return
		}
		expected := int(buf[0])
		if n, err := conn.Read(buf[:expected]); err != nil {
			log.Printf("error reading message from connection %s after byte %d: (%T) %v", conn.RemoteAddr(), n, err, err)
			return
		}
		log.Printf("message >> %q", buf[:expected])

		time.Sleep(time.Second)
		if _, err := conn.Write([]byte{31}); err != nil {
			log.Printf("error writing message to connection %s: (%T) %v", conn.RemoteAddr(), err, err)
			return
		}
	}
}

func connect(fn string) error {
	conn, err := net.Dial("unix", fn)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Printf("closing connection due to interrupt...")
		if err := conn.Close(); err != nil {
			log.Printf("close error: %v", err)
		}
	}()

	buf := []byte{0}
	msg := []byte("Hello World!\n")
	for {
		if _, err := conn.Write([]byte{byte(len(msg))}); err != nil {
			log.Printf("error writing message header %q: (%T) %v", msg, err, err)
			return nil
		}
		if n, err := conn.Write(msg); err != nil {
			log.Printf("error writing message %q after %d bytes: (%T) %v", msg, n, err, err)
			return nil
		}
		if _, err := conn.Read(buf); err != nil {
			log.Printf("error reading ok response: (%T) %v", err, err)
			return nil
		}
		if buf[0] != 31 {
			log.Printf("unexpected response: %q %d %x", buf[0], buf[0], buf[0])
			return nil
		}
	}
}
