package modulos

import (
	"tlesio/systema"

	"github.com/sirupsen/logrus"
)

const (
	ECDSA_SECP256R1_SHA256 = 0x0403
	ECDSA_SECP384R1_SHA384 = 0x0503
	ECDSA_SECP521R1_SHA512 = 0x0603
	ED25519                = 0x0807
	ED448                  = 0x0808
	RSA_PSS_PSS_SHA256     = 0x0809
	RSA_PSS_PSS_SHA384     = 0x080A
	RSA_PSS_PSS_SHA512     = 0x080B
	RSA_PKCS1_SHA256       = 0x0401
	RSA_PKCS1_SHA384       = 0x0501
	RSA_PKCS1_SHA512       = 0x0601
	RSA_PSS_RSAE_SHA256    = 0x0804
	RSA_PSS_RSAE_SHA384    = 0x0805
	RSA_PSS_RSAE_SHA512    = 0x0806
)

var SignatureHashAlgorithms = map[uint16]string{
	ECDSA_SECP256R1_SHA256: "ecdsa_secp256r1_sha256",
	ECDSA_SECP384R1_SHA384: "ecdsa_secp384r1_sha384",
	ECDSA_SECP521R1_SHA512: "ecdsa_secp521r1_sha512",
	ED25519:                "ed25519",
	ED448:                  "ed448",
	RSA_PSS_PSS_SHA256:     "rsa_pss_pss_sha256",
	RSA_PSS_PSS_SHA384:     "rsa_pss_pss_sha384",
	RSA_PSS_PSS_SHA512:     "rsa_pss_pss_sha512",
	RSA_PSS_RSAE_SHA256:    "rsa_pss_rsae_sha256",
	RSA_PSS_RSAE_SHA384:    "rsa_pss_rsae_sha384",
	RSA_PSS_RSAE_SHA512:    "rsa_pss_rsae_sha512",
	RSA_PKCS1_SHA256:       "rsa_pkcs1_sha256",
	RSA_PKCS1_SHA384:       "rsa_pkcs1_sha384",
	RSA_PKCS1_SHA512:       "rsa_pkcs1_sha512",
}

type ModSignAlgo interface {
	Name() string
	LoadData([]byte) (*SignAlgoData, error)
}

type xModSignAlgo struct {
	lg *logrus.Logger
}

type SignAlgoData struct {
	Len   uint16
	Algos []uint16
}

func NewModSignAlgo(lg *logrus.Logger) (ModSignAlgo, error) {

	if lg == nil {
		return nil, systema.ErrNilLogger
	}

	lg.Info("Module loaded: ", "Signature_Algorithms")
	return &xModSignAlgo{lg}, nil
}

func (x xModSignAlgo) Name() string {
	return "Signature_Algorithms"
}

// Assuming data is in correct format
func (x xModSignAlgo) LoadData(data []byte) (*SignAlgoData, error) {

	var offset uint16 = 2
	var newData SignAlgoData

	newData.Len = uint16(data[0])<<8 | uint16(data[1])/2
	if len(data) < int(newData.Len) {
		return nil, systema.ErrInvalidData
	}

	newData.Algos = make([]uint16, 0)
	for i := 0; i < int(newData.Len); i++ {
		newData.Algos = append(newData.Algos,
			uint16(data[offset])<<8|uint16(data[offset+1]))
		offset += 2
	}

	return &newData, nil
}
