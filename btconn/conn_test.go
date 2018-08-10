package btconn

import (
	"net"
	"testing"

	"github.com/cenkalti/rain/mse"
)

var addr = &net.TCPAddr{
	IP:   net.IPv4(127, 0, 0, 1),
	Port: 5000,
}

var (
	ext1     = [8]byte{0x0A}
	ext2     = [8]byte{0x0B}
	id1      = [20]byte{0x0C}
	id2      = [20]byte{0x0D}
	infoHash = [20]byte{0x0E}
	sKeyHash = mse.HashSKey(infoHash[:])
)

func TestUnencrypted(t *testing.T) {
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	done := make(chan struct{})
	var gerr error
	go func() {
		defer close(done)
		conn, cipher, ext, id, err2 := Dial(addr, false, false, ext1, infoHash, id1)
		if err2 != nil {
			gerr = err2
			return
		}
		if conn == nil {
			t.Errorf("conn: %s", conn)
		}
		if cipher != 0 {
			t.Errorf("cipher: %d", cipher)
		}
		if ext != ext2 {
			t.Errorf("ext: %s", ext)
		}
		if id != id2 {
			t.Errorf("id: %s", id)
		}
	}()
	conn, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	_, cipher, ext, id, ih, err := Accept(conn, nil, false, func(ih [20]byte) bool { return ih == infoHash }, ext2, id2)
	if err != nil {
		t.Fatal(err)
	}
	<-done
	if gerr != nil {
		t.Fatal(err)
	}
	if cipher != 0 {
		t.Errorf("cipher: %d", cipher)
	}
	if ext != ext1 {
		t.Errorf("ext: %s", ext)
	}
	if ih != infoHash {
		t.Errorf("ih: %s", ih)
	}
	if id != id1 {
		t.Errorf("id: %s", id)
	}
}

func TestEncrypted(t *testing.T) {
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	done := make(chan struct{})
	var gerr error
	go func() {
		defer close(done)
		conn, cipher, ext, id, err2 := Dial(addr, true, false, ext1, infoHash, id1)
		if err2 != nil {
			gerr = err2
			return
		}
		if conn == nil {
			t.Errorf("conn: %s", conn)
		}
		if cipher != mse.RC4 {
			t.Errorf("cipher: %d", cipher)
		}
		if ext != ext2 {
			t.Errorf("ext: %s", ext)
		}
		if id != id2 {
			t.Errorf("id: %s", id)
		}
		_, err2 = conn.Write([]byte("hello out"))
		if err2 != nil {
			t.Fail()
		}
		b := make([]byte, 10)
		n, err2 := conn.Read(b)
		if err2 != nil {
			t.Error(err2)
		}
		if n != 8 {
			t.Fail()
		}
		if string(b[:8]) != "hello in" {
			t.Fail()
		}
	}()
	conn, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	encConn, cipher, ext, id, ih, err := Accept(
		conn,
		func(h [20]byte) (sKey []byte) {
			if h == sKeyHash {
				return infoHash[:]
			}
			return nil
		},
		false,
		func(ih [20]byte) bool { return ih == infoHash },
		ext2, id2)
	if err != nil {
		t.Fatal(err)
	}
	if cipher != mse.RC4 {
		t.Errorf("cipher: %d", cipher)
	}
	if ext != ext1 {
		t.Errorf("ext: %s", ext)
	}
	if ih != infoHash {
		t.Errorf("ih: %s", ih)
	}
	if id != id1 {
		t.Errorf("id: %s", id)
	}
	b := make([]byte, 10)
	n, err := encConn.Read(b)
	if err != nil {
		t.Error(err)
	}
	if n != 9 {
		t.Fail()
	}
	if string(b[:9]) != "hello out" {
		t.Fail()
	}
	_, err = encConn.Write([]byte("hello in"))
	if err != nil {
		t.Fail()
	}
	<-done
	if gerr != nil {
		t.Fatal(err)
	}
}