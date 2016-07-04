package rados

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"os"
	"fmt"
	"io"
	"crypto/sha1"
	"encoding/hex"

)

type IoCtxWrapper struct {
	oid string
	striper * StriperPool
	offset int
}


func NewIoCtxWrapper(oid string, striper * StriperPool) *IoCtxWrapper{
	return &IoCtxWrapper{oid, striper, 0}
}

func (wrapper *IoCtxWrapper) Write(d []byte) (int, error) {

	n, err := wrapper.striper.Write(wrapper.oid, d, uint64(wrapper.offset))

	if err != nil {
		return n, err
	} else {
		wrapper.offset +=  len(d)
	}
	return len(d), err

}

func TestUploadDownloadCheckFile(t * testing.T) {

    assert.Equal(t, 3, len(os.Args))

    conf_file_path := os.Args[1]
    test_file_path := os.Args[2]

    conn, _ := NewConn("admin")
    err := conn.ReadConfigFile(conf_file_path)
    assert.NoError(t, err)

    fmt.Println("connecting")
    err = conn.Connect()
    assert.NoError(t, err)


    poolname := GetUUID()
    err = conn.MakePool(poolname)
    assert.NoError(t, err)


    pool, err := conn.OpenPool(poolname)
    assert.NoError(t, err)

    ioctx, err := pool.CreateStriper()
    assert.NoError(t, err)

    //start to upload
    file, err := os.Open(test_file_path)
    assert.NoError(t, err)


    ioctx.SetLayoutStripeUnit(512<<10)
    ioctx.SetLayoutObjectSize(4<<20)
    ioctx.SetLayoutStripeCount(4)


    writer := NewIoCtxWrapper("testoid", &ioctx)


    buf  := make([]byte,4<<20)
    n, err := io.CopyBuffer(writer, file, buf)

    assert.NoError(t, err)

    fmt.Printf("uploaded %d,%v\n", n, err)


    file.Seek(0,0)
    //512K
    chunks := make([]byte,512<<10)
    var chunk_chain[]byte
    for {
	    h1 := sha1.New()
	    n, err := file.Read(chunks)
	    if n == 0 {
		    break;
	    }
	    if err !=nil && err != io.EOF {
		    assert.NoError(t, err)
	    }
	    h1.Write(chunks[:n])
	    chunk_chain = append(chunk_chain, h1.Sum(nil)...)
    }

    h1 := sha1.New()
    h1.Write(chunk_chain)
    local_hex  := hex.EncodeToString(h1.Sum(nil))
    fmt.Printf("local hash info %s\n", local_hex)

    file.Close()


    //gethashinfo

    sha1_buf, _, err := pool.GetStripeSHA1("testoid")
    assert.NoError(t, err)

    h := sha1.New()
    h.Write(sha1_buf)
    b := h.Sum(nil)
    remote_hex := hex.EncodeToString(b)
    fmt.Printf("remote hash info %s\n", remote_hex)

    assert.Equal(t, remote_hex, local_hex)


    //TODO assert to others

    err = conn.DeletePool(poolname)
    assert.NoError(t, err)

    conn.Shutdown()
}
