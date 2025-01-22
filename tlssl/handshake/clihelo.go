package handshake

import (
	"encoding/binary"
	"fmt"
	"tlesio/systema"
	tx "tlesio/tlssl/extensions"

	"github.com/sirupsen/logrus"
)

// ClientHello message structure in TLS 1.2
// | Field            | Size (bytes) | Description                                              |
// |------------------|--------------|----------------------------------------------------------|
// | Version          | 2 bytes      | TLS version (e.g., 0x0303 for TLS 1.2)                   |
// | Random           | 32 bytes     | Random number generated by the client                    |
// | Session ID       | Variable     | Client session ID, can be empty                          |
// | Cipher Suites    | Variable     | List of cipher suites supported by the client            |
// | Compression Methods | Variable   | Compression methods supported by the client             |
// | Extensions       | Variable     | Extensions (optional)                                    |
// |--------------------------------------------------------------------------------------------|
// Compresion Methods will be ignored

var (
	offsetVersion         uint32 = 2
	offsetRandom          uint32 = 32
	offsetSessionIdLen    uint32 = 1
	offsetCipherSuitesLen uint32 = 2
)

type CliHello interface {
	Handle([]byte) (*MsgCliHello, error)
}

type MsgCliHello struct {
	version      [2]byte
	random       [32]byte
	sessionId    []byte
	cipherSuites []uint16
	//extensions   []
}

type xOr struct {
	helloMsg MsgCliHello
	lg       *logrus.Logger
	exts     tx.TLSExtension
}

func NewCliHello(lg *logrus.Logger, exts tx.TLSExtension) CliHello {

	if lg == nil || exts == nil {
		return nil
	}

	return &xOr{
		lg:   lg,
		exts: exts,
	}
}

func (rox *xOr) Handle(buffer []byte) (*MsgCliHello, error) {

	var err error
	var aux uint32
	var offset uint32 = 0

	if buffer == nil || len(buffer) < 38 {
		return nil, fmt.Errorf("ClientHello buffer is nil or too small")
	}

	offset += rox.parseVersion(buffer)
	offset += rox.parseRandom(buffer[offset:])
	aux, err = rox.parseSessionID(buffer[offset:])
	if err != nil {
		return nil, err
	}

	offset += aux
	aux, err = rox.parseCipherSuites(buffer[offset:])
	if err != nil {
		return nil, err
	}

	offset += aux
	// Skip parsing Compression Methods
	if len(buffer) < int(offset+1) {
		return nil, fmt.Errorf("buffer too small in Compression Methods len")
	}

	compressionMethodsLen := uint32(buffer[offset])
	offset += 1 + compressionMethodsLen
	rox.parseExtensions(buffer[offset:])
	return &MsgCliHello{}, nil
}

func (rox *xOr) parseVersion(buffer []byte) uint32 {

	rox.helloMsg.version = [2]byte{buffer[0], buffer[1]}
	rox.lg.Trace("Field[Version]: ",
		systema.PrettyPrintBytes(rox.helloMsg.version[:]))
	return offsetVersion
}

func (rox *xOr) parseRandom(buffer []byte) uint32 {

	copy(rox.helloMsg.random[:], buffer[:offsetRandom])
	rox.lg.Trace("Field[Random]: ",
		systema.PrettyPrintBytes(rox.helloMsg.random[:]))
	return offsetRandom
}

func (rox *xOr) parseSessionID(buffer []byte) (uint32, error) {

	if len(buffer) < 1 {
		return 0, fmt.Errorf("sessionID field is too small")
	}

	fieldLen := uint32(buffer[0])
	offset := uint32(offsetSessionIdLen)
	if len(buffer) < int(offset+fieldLen) {
		return 0, fmt.Errorf("sessionID field is too small")
	}

	rox.helloMsg.sessionId = make([]byte, fieldLen)
	copy(rox.helloMsg.sessionId, buffer[offset:offset+fieldLen])
	rox.lg.Trace("Field[SessionID]: ",
		systema.PrettyPrintBytes(rox.helloMsg.sessionId))
	return offset + fieldLen, nil
}

func (rox *xOr) parseCipherSuites(buffer []byte) (uint32, error) {

	offset := uint32(offsetCipherSuitesLen)
	fieldLen := binary.BigEndian.Uint16(buffer[:2])
	if len(buffer) < int(fieldLen) {
		return 0, fmt.Errorf("CipherSuites field is too small")
	}

	if fieldLen%2 != 0 {
		return 0, fmt.Errorf("CipherSuites field is not a multiple of 2")
	}

	fl := fieldLen / 2
	rox.helloMsg.cipherSuites = make([]uint16, fl)
	for i := uint16(0); i < fl; i++ {
		rox.helloMsg.cipherSuites[i] = binary.BigEndian.Uint16(buffer[offset : offset+2])
		offset += 2
	}

	rox.lg.Trace("Field[CipherSuites]: ",
		printCipherSuiteNames(rox.helloMsg.cipherSuites))
	return offset, nil
}

func (rox *xOr) parseExtensions(buffer []byte) {

	// Parsing only supported extensions
	if len(buffer) < 2 {
		return
	}

	offset := 0
	extLen := binary.BigEndian.Uint16(buffer[:2])
	if len(buffer) < int(extLen) {
		return
	}

	offset += 2
	for offset < int(extLen) {
		extt := binary.BigEndian.Uint16(buffer[offset : offset+2])
		exttLen := binary.BigEndian.Uint16(buffer[offset+2 : offset+4])
		// Points
		offset += 2 + 2

		fmt.Printf("Extension Type/Len: %v/%v\n", tx.ExtensionName[extt], exttLen)
		offset += int(exttLen)
	}
}
