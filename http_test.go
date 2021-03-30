package webutils

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"testing"
)

func TestRequest_Bytes_makeBodyReader_ReturnsReader(t *testing.T) {
	data := []byte("0123")
	r := Request{Body: data}

	reader, err := r.makeBodyReader()

	require.NoError(t, err)
	require.Implements(t, (*io.Reader)(nil), reader)
	gotData, _ := ioutil.ReadAll(reader)
	require.Equal(t, data, gotData)
}

func TestRequest_Struct_makeBodyReader_ReturnsReader(t *testing.T) {
	data := struct {
		Field string
	}{Field: "0123"}
	r := Request{Body: data}

	reader, err := r.makeBodyReader()

	require.NoError(t, err)
	require.Implements(t, (*io.Reader)(nil), reader)
	gotData, _ := ioutil.ReadAll(reader)
	require.Equal(t, []byte("{\"Field\":\"0123\"}"), gotData)
}

func TestRequest_Reader_makeBodyReader_ReturnsSameReader(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	r := Request{Body: reader}

	gotReader, err := r.makeBodyReader()

	require.NoError(t, err)
	require.Equal(t, reader, gotReader)
}
