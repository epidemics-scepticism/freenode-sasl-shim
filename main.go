package main

import (
	"crypto/tls"
	"log"
	"net"
)

func dialFreenode() (net.Conn, error) {
	conn, e := dialSocks5("127.0.0.1:9050", "freenodeok2gncmy.onion:6697")
	if e != nil {
		log.Fatalln("ERROR", e)
	}
	cfg := &tls.Config{
		ServerName:   "zettel.freenode.net",
	}
	sasl, e := tls.LoadX509KeyPair("sasl.crt", "sasl.key")
	if e != nil {
		conn.Close()
		return conn, e
	}
	cfg.Certificates = []tls.Certificate{sasl}
	tconn := tls.Client(conn, cfg)
	if e := tconn.Handshake(); e != nil {
		conn.Close()
		return conn, e
	}
	conn = net.Conn(tconn)
	state := tconn.ConnectionState()
	for k, cert := range state.PeerCertificates {
		log.Println("INFO",
			"Chain #",
			k,
			"Subject",
			cert.Subject.CommonName,
			"Issuer",
			cert.Issuer.CommonName,
		)
	}
	_, e = conn.Write([]byte("CAP REQ :sasl\r\nAUTHENTICATE EXTERNAL\r\nAUTHENTICATE +\r\nCAP END\r\n"))
	if e != nil {
		log.Println("ERROR", e)
		conn.Close()
		return conn, e
	}
	return conn, nil
}

func netToChan(c chan []byte, n net.Conn) {
	b := make([]byte, 4096)
	for {
		if i, e := n.Read(b[:]); e != nil {
			close(c)
			return
		} else {
			c <- b[:i]
		}
	}
}

func toNet(b []byte, n net.Conn) error {
	for i := 0; i < len(b); {
		if t, e := n.Write(b[i:]); e != nil {
			return e
		} else {
			i += t
		}
	}
	return nil
}

func ferry(client, server net.Conn) {
	defer client.Close()
	defer server.Close()
	defer log.Println("WARN", "Connection closed")
	fromClient := make(chan []byte)
	fromServer := make(chan []byte)
	go netToChan(fromClient, client)
	go netToChan(fromServer, server)
	for {
		select {
		case b := <-fromClient:
			if e := toNet(b, server); e != nil {
				log.Println("ERROR", e)
				return
			}
		case b := <-fromServer:
			if e := toNet(b, client); e != nil {
				log.Println("ERROR", e)
				return
			}
		}
	}
}

func main() {
	listener, e := net.Listen("tcp4", "127.0.0.1:17649")
	if e != nil {
		log.Fatalln("ERROR", e)
	}
	for {
		client, e := listener.Accept()
		if e != nil {
			log.Println("ERROR", e)
		}
		server, e := dialFreenode()
		if e != nil {
			log.Println("ERROR", e)
		}
		go ferry(client, server)
	}
}
