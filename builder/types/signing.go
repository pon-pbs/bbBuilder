package bbTypes

import (
	"github.com/ethereum/go-ethereum/builder/bls"
	commonTypes "github.com/bsn-eng/pon-golang-types/common"
)

type (
	Domain      [32]byte
	DomainType  [4]byte
	ForkVersion [4]byte
)

var (
	DomainBuilder Domain

	DomainTypeBeaconProposer = DomainType{0x00, 0x00, 0x00, 0x00}
	DomainTypeAppBuilder     = DomainType{0x00, 0x00, 0x00, 0x01}
)

func init() {
	DomainBuilder = ComputeDomain(DomainTypeAppBuilder, ForkVersion{}, commonTypes.Root{})
}

type SigningData struct {
	Root   commonTypes.Root   `ssz-size:"32"`
	Domain Domain `ssz-size:"32"`
}

type ForkData struct {
	CurrentVersion        ForkVersion `ssz-size:"4"`
	GenesisValidatorsRoot commonTypes.Root        `ssz-size:"32"`
}

type HashTreeRoot interface {
	HashTreeRoot() ([32]byte, error)
}

func ComputeDomain(dt DomainType, forkVersion ForkVersion, genesisValidatorsRoot commonTypes.Root) [32]byte {
	forkDataRoot, _ := (&ForkData{
		CurrentVersion:        forkVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}).HashTreeRoot()

	var domain [32]byte
	copy(domain[0:4], dt[:])
	copy(domain[4:], forkDataRoot[0:28])

	return domain
}

func ComputeSigningRoot(obj HashTreeRoot, d Domain) ([32]byte, error) {
	var zero [32]byte
	root, err := obj.HashTreeRoot()
	if err != nil {
		return zero, err
	}
	signingData := SigningData{root, d}
	msg, err := signingData.HashTreeRoot()
	if err != nil {
		return zero, err
	}
	return msg, nil
}

func SignMessage(obj HashTreeRoot, d Domain, sk *bls.SecretKey) (commonTypes.Signature, error) {
	root, err := ComputeSigningRoot(obj, d)
	if err != nil {
		return commonTypes.Signature{}, err
	}

	signatureBytes := bls.Sign(sk, root[:]).Compress()

	var signature commonTypes.Signature
	err = signature.FromSlice(signatureBytes)
	if err != nil {
		return [96]byte{}, err
	}

	return signature, nil
}

func VerifySignature(obj HashTreeRoot, d Domain, pkBytes, sigBytes []byte) (bool, error) {
	msg, err := ComputeSigningRoot(obj, d)
	if err != nil {
		return false, err
	}

	return bls.VerifySignatureBytes(msg[:], sigBytes, pkBytes)
}
