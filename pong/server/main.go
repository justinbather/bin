package main

import (
	"bufio"
	"net"
	"os"

	"github.com/justinbather/prettylog"
)

func main() {
	log := prettylog.New()

	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Errorf("Error starting server: %s", err)
		os.Exit(1)
	}
	defer l.Close()

	log.Info("Listening for connections on :8000")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Infof("Error accepting connection: %s", err)
		}

		log.Info("New Connection!")
		go handleConn(conn, log)
	}
}

func handleConn(conn net.Conn, log *prettylog.Logger) {
	defer conn.Close()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	for {
		var buf []byte
		_, err := r.Read(buf)
		if err != nil {
			log.Warn("Connection closed by client")
			return
		}

		log.Info("Client sent: ", buf)
		w.Write(buf)
	}

}
