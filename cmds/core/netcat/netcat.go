// Copyright 2012-2023 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// netcat creates arbitrary TCP and UDP connections and listens and sends arbitrary data.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/u-root/u-root/pkg/uroot/util"
)

const usage = "netcat [go-style network address]"

var errMissingHostnameOrPort = fmt.Errorf("missing hostname or port")

type params struct {
	network string
	listen  bool
	verbose bool
	host    string
	port    string
}

func parseParms() (params, error) {
	netType := flag.String("net", "tcp", "What net type to use, e.g. tcp, unix, etc.")
	listen := flag.Bool("l", false, "Listen for connections.")
	verbose := flag.Bool("v", false, "Verbose output.")
	flag.Parse()

	if len(flag.Args()) != 2 {
		return params{}, errMissingHostnameOrPort
	}

	return params{
		network: *netType,
		listen:  *listen,
		verbose: *verbose,
		host:    flag.Args()[0],
		port:    flag.Args()[1],
	}, nil
}

type cmd struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	params
}

func command(stdin io.Reader, stdout io.Writer, stderr io.Writer, p params) *cmd {
	return &cmd{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		params: p,
	}
}

func init() {
	flag.Usage = util.Usage(flag.Usage, usage)
}

func main() {
	p, err := parseParms()
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}
	if err := command(os.Stdin, os.Stdout, os.Stderr, p).run(); err != nil {
		log.Fatalf("nc: %v", err)
	}
}

func (c *cmd) run() error {
	var conn net.Conn
	var err error

	addr := net.JoinHostPort(c.host, c.port)

	if c.listen {
		ln, err := net.Listen(c.network, addr)
		if err != nil {
			return err
		}
		if c.verbose {
			fmt.Fprintln(c.stderr, "Listening on", ln.Addr())
		}

		conn, err = ln.Accept()
		if err != nil {
			return err
		}
	} else {
		if conn, err = net.Dial(c.network, addr); err != nil {
			return err
		}
	}
	if c.verbose {
		fmt.Fprintln(c.stderr, "Connected to", conn.RemoteAddr())
	}

	go func() {
		if _, err := io.Copy(conn, c.stdin); err != nil {
			fmt.Fprintln(c.stderr, err)
		}
	}()
	if _, err = io.Copy(c.stdout, conn); err != nil {
		fmt.Fprintln(c.stderr, err)
	}
	if c.verbose {
		fmt.Fprintln(c.stderr, "Disconnected")
	}

	return nil
}
