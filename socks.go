// Copyright 2012, Hailiang Wang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"net"
	"strconv"
)

func dialSocks5(proxy, targetAddr string) (conn net.Conn, err error) {
	// dial TCP
	conn, err = net.Dial("tcp", proxy)
	if err != nil {
		return
	}

	// version identifier/method selection request
	req := []byte{
		5, // version number
		1, // number of methods
		0, // method 0: no authentication (only anonymous access supported for now)
	}
	resp, err := sendReceive(conn, req)
	if err != nil {
		return
	} else if len(resp) != 2 {
		err = errors.New("Server does not respond properly.")
		return
	} else if resp[0] != 5 {
		err = errors.New("Server does not support Socks 5.")
		return
	} else if resp[1] != 0 { // no auth
		err = errors.New("socks method negotiation failed.")
		return
	}

	// detail request
	host, port, err := splitHostPort(targetAddr)
	req = []byte{
		5,               // version number
		1,               // connect command
		0,               // reserved, must be zero
		3,               // address type, 3 means domain name
		byte(len(host)), // address length
	}
	req = append(req, []byte(host)...)
	req = append(req, []byte{
		byte(port >> 8), // higher byte of destination port
		byte(port),      // lower byte of destination port (big endian)
	}...)
	resp, err = sendReceive(conn, req)
	if err != nil {
		return
	} else if len(resp) != 10 {
		err = errors.New("Server does not respond properly.")
	} else if resp[1] != 0 {
		err = errors.New("Can't complete SOCKS5 connection.")
	}

	return
}

func sendReceive(conn net.Conn, req []byte) (resp []byte, err error) {
	_, err = conn.Write(req)
	if err != nil {
		return
	}
	resp, err = readAll(conn)
	return
}

func readAll(conn net.Conn) (resp []byte, err error) {
	resp = make([]byte, 1024)
	n, err := conn.Read(resp)
	resp = resp[:n]
	return
}

func splitHostPort(addr string) (host string, port uint16, err error) {
	host, portStr, err := net.SplitHostPort(addr)
	portInt, err := strconv.ParseUint(portStr, 10, 16)
	port = uint16(portInt)
	return
}
