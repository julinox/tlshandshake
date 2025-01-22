package tlssl

import (
	"encoding/binary"
	"fmt"
	"tlesio/systema"
)

// Format of Handshake messages in TLS 1.2:
// -------------------------------------------
// Each handshake message has the following structure:
//
// | Field             | Size (bytes) | Description                                              |
// |-------------------|--------------|----------------------------------------------------------|
// | HandshakeType     | 1 byte       | Type of handshake message. Example:                      |
// |                   |              | - 1: ClientHello                                         |
// |                   |              | - 2: ServerHello                                         |
// |                   |              | - 11: Certificate                                        |
// |                   |              | - 14: ServerHelloDone                                    |
// |                   |              | - 16: ClientKeyExchange                                  |
// |                   |              | - 20: Finished                                           |
// |-------------------|--------------|----------------------------------------------------------|
// | Length            | 3 bytes      | Len of the message (in bytes), excluding the type field. |
// |-------------------|--------------|----------------------------------------------------------|
// | Handshake Message | Variable     | The body of the message (varies according to the         |
// |                                  | handshake type).                                         |
// |---------------------------------------------------------------------------------------------|

// Example structure for a ClientHello:
// | Field            | Size (bytes) | Description                                              |
// |------------------|--------------|----------------------------------------------------------|
// | HandshakeType    | 1 byte       | 0x01 (ClientHello)                                       |
// | Length           | 3 bytes      | Len of the message (in bytes), excluding HandshakeType   |
// | Version          | 2 bytes      | TLS version (e.g., 0x0303 for TLS 1.2)                   |
// | Random           | 32 bytes     | Random number generated by the client                    |
// | Session ID       | Variable     | Client session ID, can be empty                          |
// | Cipher Suites    | Variable     | List of cipher suites supported by the client            |
// | Compression Methods | Variable   | Compression methods supported by the client             |
// |--------------------------------------------------------------------------------------------|

type HandshakeTypeType uint8

const (
	HandshakeTypeClientHelo        HandshakeTypeType = 1
	HandshakeTypeServerHelo        HandshakeTypeType = 2
	HandshakeTypeCertificate       HandshakeTypeType = 11
	HandshakeTypeServerHeloDone    HandshakeTypeType = 14
	HandshakeTypeClientKeyExchange HandshakeTypeType = 16
	HandshakeTypeFinished          HandshakeTypeType = 20
)

type handshakeMsg struct {
	rcvBuffSize   int
	length        uint32
	handshakeType HandshakeTypeType
}

func processHandshakeMsg(ctrl *tlsio, buffer []byte) error {

	var err error
	var newMsg handshakeMsg

	if ctrl == nil {
		return systema.ErrNilController
	}

	if ctrl.logg == nil {
		return systema.ErrNilLogger
	}

	if buffer == nil {
		return systema.ErrNilParams
	}

	if len(buffer) < 4 {
		return fmt.Errorf("handshake message size did not match 4 bytes")
	}

	newMsg.rcvBuffSize = len(buffer)
	newMsg.handshakeType = HandshakeTypeType(buffer[0])
	buffer[0] = 0
	newMsg.length = binary.BigEndian.Uint32(buffer[:4])
	ctrl.logg.Trace(&newMsg)
	switch newMsg.handshakeType {
	case HandshakeTypeClientHelo:
		_, err = ctrl.handShakeIf.cliHello.Handle(buffer[4:])
	}

	return err
}

func (h HandshakeTypeType) String() string {

	switch h {
	case HandshakeTypeClientHelo:
		return "Client Hello"
	case HandshakeTypeServerHelo:
		return "Server Hello"
	case HandshakeTypeCertificate:
		return "Certificate"
	case HandshakeTypeServerHeloDone:
		return "Server Hello Done"
	case HandshakeTypeClientKeyExchange:
		return "Client Key Exchange"
	case HandshakeTypeFinished:
		return "Finished"
	default:
		return "Unknow"
	}
}

func (hm *handshakeMsg) String() string {
	return fmt.Sprintf("HandshakeType: %v | Len: %v",
		hm.handshakeType, hm.length)
}
