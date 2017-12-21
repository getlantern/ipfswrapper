package ipfswrapper

import (
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAndGet(t *testing.T) {
	node, estart := Start(path.Join(os.TempDir(), "ipfs_test_repo"), "")
	assert.NoError(t, estart, "should start IPFS service")
	defer node.Stop()
	s := strconv.FormatInt(rand.Int63(), 10)
	path, eadd := node.AddString(s)
	assert.NoError(t, eadd, "should have no error adding string")
	d, eget := node.GetString(path)
	assert.NoError(t, eget, "should have no error getting string")
	assert.Equal(t, d, s, "should get the same string")
	r, eget := node.Get(path)
	assert.EqualValues(t, len(s), r.Size(), "should have correct length")
	assert.EqualValues(t, 0, r.Offset(), "should have an offset of 0")
	b := []byte{0}
	_, eread := r.Read(b)
	assert.NoError(t, eread, "should have no error reading from DagReader")
	assert.Equal(t, s[0], b[0], "should have read the same content")

	p, eseek := r.Seek(1, io.SeekCurrent)
	assert.NoError(t, eseek, "should have no error seeking")
	assert.EqualValues(t, 2, p, "should have read the same content")
	assert.EqualValues(t, 2, r.Offset(), "should have read the same content")
	b, eread = ioutil.ReadAll(r)
	assert.NoError(t, eread, "should have no error reading the rest")
	assert.Equal(t, s[2:], string(b), "should have read the rest content")
	assert.NoError(t, r.Close(), "should have no error closing the reader")

}
