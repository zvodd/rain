package librain

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/cenkalti/rain/internal/shared"
)

func (p *peerConn) readHandShake1() (*shared.InfoHash, error) {
	var pstrLen byte
	err := binary.Read(p.conn, binary.BigEndian, &pstrLen)
	if err != nil {
		return nil, err
	}
	if pstrLen != bitTorrent10pstrLen {
		return nil, fmt.Errorf("invalid pstrlen: %d != %d", pstrLen, bitTorrent10pstrLen)
	}

	pstr := make([]byte, bitTorrent10pstrLen)
	_, err = io.ReadFull(p.conn, pstr)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(pstr, bitTorrent10pstr) != 0 {
		return nil, fmt.Errorf("invalid pstr: %q != %q", string(pstr), string(bitTorrent10pstr))
	}

	_, err = io.CopyN(ioutil.Discard, p.conn, 8) // reserved bytes are not used
	if err != nil {
		return nil, err
	}

	var infoHash shared.InfoHash
	_, err = io.ReadFull(p.conn, infoHash[:])
	if err != nil {
		return nil, err
	}

	return &infoHash, nil
}

func (p *peerConn) readHandShake2() (*shared.PeerID, error) {
	var id shared.PeerID
	_, err := io.ReadFull(p.conn, id[:])
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (p *peerConn) sendHandShake(ih shared.InfoHash, id shared.PeerID) error {
	return binary.Write(p.conn, binary.BigEndian, newPeerHandShake(ih, id))
}

type peerHandShake struct {
	Pstrlen  byte
	Pstr     [bitTorrent10pstrLen]byte
	_        [8]byte
	InfoHash shared.InfoHash
	PeerID   shared.PeerID
}

func newPeerHandShake(ih shared.InfoHash, id shared.PeerID) *peerHandShake {
	h := &peerHandShake{
		Pstrlen:  bitTorrent10pstrLen,
		InfoHash: ih,
		PeerID:   id,
	}
	copy(h.Pstr[:], bitTorrent10pstr)
	return h
}