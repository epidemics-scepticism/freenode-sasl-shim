package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/textproto"
)

func dialFreenode() (*textproto.Conn, error) {
	socksConn, e := dialSocks5("127.0.0.1:9050", "freenodeok2gncmy.onion:6697")
	if e != nil {
		socksConn.Close()
		return nil, e
	}
	cfg := &tls.Config{
		ServerName: "zettel.freenode.net",
	}
	sasl, e := tls.LoadX509KeyPair("sasl.crt", "sasl.key")
	if e != nil {
		socksConn.Close()
		return nil, e
	}
	cfg.Certificates = []tls.Certificate{sasl}
	tlsConn := tls.Client(socksConn, cfg)
	if e := tlsConn.Handshake(); e != nil {
		socksConn.Close()
		return nil, e
	}
	conn := textproto.NewConn(tlsConn)
	state := tlsConn.ConnectionState()
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
	if conn.PrintfLine("%s", "CAP REQ :sasl") != nil ||
		conn.PrintfLine("%s", "AUTHENTICATE EXTERNAL") != nil ||
		conn.PrintfLine("%s", "AUTHENTICATE +") != nil ||
		conn.PrintfLine("%s", "CAP END") != nil {
		log.Println("ERROR", e)
		conn.Close()
		return nil, e
	}
	return conn, nil
}

func netToChan(c chan string, n *textproto.Conn) {
	for {
		if s, e := n.ReadLine(); e != nil {
			close(c)
			return
		} else {
			c <- s
		}
	}
}

func toNet(s string, n *textproto.Conn) error {
	if e := n.PrintfLine("%s", s); e != nil {
		return e
	}
	return nil
}

func ferry(client, server *textproto.Conn) {
	defer client.Close()
	defer server.Close()
	defer log.Println("WARN", "Connection closed")
	fromClient := make(chan string)
	fromServer := make(chan string)
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
	listener, e := net.Listen("tcp", "127.0.0.1:17649")
	if e != nil {
		log.Fatalln("ERROR", e)
	}
	log.Println("INFO", "Listening on", listener.Addr().String())
	for {
		tcpClient, e := listener.Accept()
		if e != nil {
			log.Println("ERROR", e)
			continue
		}
		log.Println("INFO", "Connection from", tcpClient.RemoteAddr().String())
		client := textproto.NewConn(tcpClient)
		server, e := dialFreenode()
		if e != nil {
			log.Println("ERROR", e)
			client.Close()
			continue
		}
		go ferry(client, server)
	}
}
