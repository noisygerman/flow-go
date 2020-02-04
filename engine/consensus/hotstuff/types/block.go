package types

import (
	"time"

	"github.com/dapperlabs/flow-go/model/flow"
)

type Block struct {
	// specified
	View        uint64
	QC          *QuorumCertificate
	PayloadHash []byte

	// configed
	Height  uint64
	ChainID string

	// autogenerated
	Timestamp time.Time
}

func (b *Block) ID() flow.Identifier {
	panic("TODO")
}

func NewBlock(view uint64, qc *QuorumCertificate, payloadHash []byte, height uint64, chainID string) *Block {

	t := time.Now()

	return &Block{
		View:        view,
		QC:          qc,
		PayloadHash: payloadHash,
		Height:      height,
		ChainID:     chainID,
		Timestamp:   t,
	}
}

func (b *Block) ToVote() *UnsignedVote {
	panic("TODO")
}
