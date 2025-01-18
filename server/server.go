package handshake

import (
	"net"
	"tlesio/tlssl"

	clog "github.com/julinox/consolelogrus"
	"github.com/sirupsen/logrus"
)

const (
	port         = ":8443"
	responseBody = "Hello, TLS!"
)

type zzl struct {
	lg     *logrus.Logger
	tessio tlssl.TLS12
}

func RealServidor() {

	var ssl zzl
	var err error

	ssl.lg = clog.InitNewLogger(&clog.CustomFormatter{Tag: "SERVER"})
	if err != nil {
		ssl.lg.Error("Error creating TLS Control: ", err)
		return
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		ssl.lg.Error(err)
		return
	}

	defer listener.Close()
	ssl.lg.Info("Listening on PORT ", port)
	_, err = tlssl.NewTLSDefault()
	if err != nil {
		ssl.lg.Error("Error creating TLS Control: ", err)
		return
	}

	/*for {
		conn, err := listener.Accept()
		if err != nil {
			ssl.lg.Error("Error accepting connection:", err)
			continue
		}

		go ssl.handleConnection(conn)
	}*/
}

func (ssl *zzl) handleConnection(conn net.Conn) {

	defer conn.Close()
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		ssl.lg.Error("Error reading data:", err)
		return
	}

	if n <= 5 {
		ssl.lg.Warning("Very little Data")
		return
	}

	tlssl.TLSMe(buffer[:n], nil)
}
