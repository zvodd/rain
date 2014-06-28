package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io/ioutil"
	"os"

	"code.google.com/p/bencode-go"

	"github.com/cenkalti/rain/internal/shared"
)

type Torrent struct {
	Info         Info       `bencode:"info"`
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	CreationDate int64      `bencode:"creation date"`
	Comment      string     `bencode:"comment"`
	CreatedBy    string     `bencode:"created by"`
	Encoding     string     `bencode:"encoding"`
}

type Info struct {
	PieceLength uint32 `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Private     byte   `bencode:"private"`
	Name        string `bencode:"name"`
	// Single File Mode
	Length int64  `bencode:"length"`
	Md5sum string `bencode:"md5sum"`
	// Multiple File mode
	Files []fileDict `bencode:"files"`

	// These fields do not exist in torrent file.
	// They are calculated when a Torrent is created with New.
	Hash        shared.InfoHash
	TotalLength int64
	NumPieces   uint32
}

type fileDict struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"`
	Md5sum string   `bencode:"md5sum"`
}

func New(path string) (*Torrent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)

	decoded, err := bencode.Decode(reader)
	if err != nil {
		return nil, err
	}

	torrentMap, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid torrent file")
	}

	infoMap, ok := torrentMap["info"]
	if !ok {
		return nil, errors.New("invalid torrent file")
	}

	t := new(Torrent)

	// Unmarshal bencoded bytes into the struct
	reader.Seek(0, os.SEEK_SET)
	err = bencode.Unmarshal(reader, t)
	if err != nil {
		return nil, err
	}

	// Calculate InfoHash
	hash := sha1.New()
	bencode.Marshal(hash, infoMap)
	copy(t.Info.Hash[:], hash.Sum(nil))

	// Calculate TotalLength
	if !t.Info.MultiFile() {
		t.Info.TotalLength = t.Info.Length
	} else {
		for _, f := range t.Info.Files {
			t.Info.TotalLength += f.Length
		}
	}

	// Calculate NumPieces
	t.Info.NumPieces = uint32(len(t.Info.Pieces)) / sha1.Size

	return t, nil
}

func (i *Info) HashOfPiece(index uint32) [sha1.Size]byte {
	if index >= i.NumPieces {
		panic("piece index out of range")
	}
	var hash [sha1.Size]byte
	start := index * sha1.Size
	end := start + sha1.Size
	copy(hash[:], []byte(i.Pieces[start:end]))
	return hash
}

// GetFiles returns the files in torrent as a slice, even if there is a single file.
func (i *Info) GetFiles() []fileDict {
	if i.MultiFile() {
		return i.Files
	} else {
		return []fileDict{fileDict{i.Length, []string{i.Name}, i.Md5sum}}
	}
}

// MultiFile returns true if the torrent contains more than one file.
func (i *Info) MultiFile() bool {
	return len(i.Files) != 0
}