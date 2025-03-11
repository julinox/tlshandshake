package handshake

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

import (
	"encoding/binary"
	"fmt"
	"strings"
	"tlesio/systema"
	"tlesio/tlssl"
	ex "tlesio/tlssl/extensions"
	"tlesio/tlssl/suite"
)

var (
	offsetVersion         uint32 = 2
	offsetRandom          uint32 = 32
	offsetSessionIdLen    uint32 = 1
	offsetCipherSuitesLen uint32 = 2
)

type MsgHello struct {
	Version      [2]byte
	Random       [32]byte
	SessionId    []byte
	CipherSuites []uint16
	Extensions   map[uint16]interface{} //ExtensionType -> ExtensionData
}

type xClientHello struct {
	stateBasicInfo
	tCtx *tlssl.TLSContext
}

func NewClientHello(actx *AllContexts) ClientHello {

	var newX xClientHello

	if actx == nil || actx.Tctx == nil || actx.Hctx == nil {
		return nil
	}

	newX.ctx = actx.Hctx
	newX.tCtx = actx.Tctx
	return &newX
}

func (x *xClientHello) Name() string {
	return "_ClientHello_"
}

func (x *xClientHello) Next() (int, error) {
	return x.nextState, x.Handle()
}

func (x *xClientHello) Handle() error {

	var err error
	var newMsg MsgHello
	var offset, aux uint32

	offset = 0
	x.tCtx.Lg.Tracef("Running state: %v", x.Name())
	x.tCtx.Lg.Debugf("Running state: %v", x.Name())
	buff := x.ctx.GetBuffer(CLIENTHELLO)
	if buff == nil {
		return fmt.Errorf("nil ClientHello buffer")
	}

	cliHelloBuf := buff[tlssl.TLS_HEADER_SIZE+tlssl.TLS_HANDSHAKE_SIZE:]
	if len(cliHelloBuf) < 38 {
		return fmt.Errorf("ClientHello buffer is too small")
	}

	offset += x.version(cliHelloBuf[offset:], &newMsg)
	offset += x.random(cliHelloBuf[offset:], &newMsg)
	aux, err = x.sessionId(cliHelloBuf[offset:], &newMsg)
	if err != nil {
		return err
	}

	offset += aux
	aux, err = x.cSuites(cliHelloBuf[offset:], &newMsg)
	if err != nil {
		return err
	}

	offset += aux
	// Skip parsing Compression Methods
	if len(cliHelloBuf) < int(offset+1) {
		return fmt.Errorf("buffer too small in Compression Methods len")
	}

	compressionMethodsLen := uint32(cliHelloBuf[offset])
	offset += 1 + compressionMethodsLen
	newMsg.Extensions = make(map[uint16]interface{})
	offset += x.extensions(cliHelloBuf[offset:], &newMsg)

	if int(offset) != len(cliHelloBuf) {
		return fmt.Errorf("ClientHello message parse doesnt match offset")
	}

	x.ctx.AppendOrder(CLIENTHELLO)
	x.ctx.SetMsgHello(&newMsg)
	x.ctx.SetBuffer(CLIENTRANDOM, newMsg.Random[:])
	x.nextState = SERVERHELLO
	return nil
}

func (x *xClientHello) version(buff []byte, msg *MsgHello) uint32 {

	msg.Version = [2]byte{buff[0], buff[1]}
	x.tCtx.Lg.Trace("Field[Version]: ",
		systema.PrettyPrintBytes(msg.Version[:]))
	return offsetVersion
}

func (x *xClientHello) random(buff []byte, msg *MsgHello) uint32 {

	copy(msg.Random[:], buff)
	x.tCtx.Lg.Tracef("Field[Random]: %x", msg.Random)
	return offsetRandom
}

func (x *xClientHello) sessionId(buff []byte, msg *MsgHello) (uint32, error) {

	sessionIdLen := uint32(buff[0])
	if sessionIdLen == 0 {
		return offsetSessionIdLen, nil
	}

	if sessionIdLen > 32 {
		return 0, fmt.Errorf("invalid session ID length")
	}

	msg.SessionId = make([]byte, sessionIdLen)
	copy(msg.SessionId, buff[1:sessionIdLen+1])
	x.tCtx.Lg.Tracef("Field[SessionID]: %x", msg.SessionId)
	return offsetSessionIdLen + sessionIdLen, nil
}

// Cipher Suites
func (x *xClientHello) cSuites(buffer []byte, msg *MsgHello) (uint32, error) {

	offset := uint32(offsetCipherSuitesLen)
	fieldLen := binary.BigEndian.Uint16(buffer[:2])
	if len(buffer) < int(fieldLen) {
		return 0, fmt.Errorf("CipherSuites field is too small")
	}

	if fieldLen%2 != 0 {
		return 0, fmt.Errorf("CipherSuites field is not a multiple of 2")
	}

	fl := fieldLen / 2
	msg.CipherSuites = make([]uint16, fl)
	for i := uint16(0); i < fl; i++ {
		msg.CipherSuites[i] = binary.BigEndian.Uint16(
			buffer[offset : offset+2])
		offset += 2
	}

	x.tCtx.Lg.Trace("Field[CipherSuites]: ",
		algosToName(0xFFFF, msg.CipherSuites))
	return offset, nil
}

// Parse and store only supported extensions data
func (x *xClientHello) extensions(buffer []byte, msg *MsgHello) uint32 {

	if len(buffer) < 2 {
		return 0
	}

	offset := 0
	extLen := binary.BigEndian.Uint16(buffer[:2])
	if len(buffer) < int(extLen) {
		return 0
	}

	offset += 2
	for offset < int(extLen) {
		extID := binary.BigEndian.Uint16(buffer[offset : offset+2])
		extLen := binary.BigEndian.Uint16(buffer[offset+2 : offset+4])
		offset += 2 + 2
		ext := x.tCtx.Exts.Get(extID)
		if ext != nil {
			data, err := ext.LoadData(
				buffer[offset:offset+int(extLen)], int(extLen))
			if err != nil {
				x.tCtx.Lg.Errorf("data load(%v): %v",
					ex.ExtensionName[extID], err)
			} else {
				msg.Extensions[extID] = data
				x.tCtx.Lg.Trace(fmt.Sprintf("Field[Extension %v]: %v",
					ex.ExtensionName[extID],
					ext.PrintRaw(buffer[offset:offset+int(extLen)])))
			}
		}

		offset += int(extLen)
	}

	return uint32(offset)
}

func algoToName(varr, algo uint16) string {

	switch varr {
	case 0xFFFF:
		return fmt.Sprintf("%s(0x%04X)", suite.CipherSuiteNames[algo], algo)

	case 0x000D:
		return fmt.Sprintf("%s(0x%04X)", ex.SignHashAlgorithms[algo], algo)
	}

	return "unknown_algorithm_name_or_type"
}

func algosToName(varr uint16, algos []uint16) string {

	var names []string

	for _, v := range algos {
		names = append(names, algoToName(varr, v))
	}

	return fmt.Sprintf("\n%s", strings.Join(names, "\n"))
}
