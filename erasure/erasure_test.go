package erasure

import (
	"bytes"
	"crypto/rand"
	"testing"
)

const (
	k          = 4
	m          = 2
	w          = 8
	packetsize = 8
)

func TestCRSEncode(t *testing.T) {
	stripesize := packetsize * w

	data := make([][]byte, k)
	for i := range data {
		data[i] = make([]byte, stripesize)
		rand.Read(data[i])
	}

	coding := make([][]byte, m)
	for i := range coding {
		coding[i] = make([]byte, stripesize)
	}

	// coding matrix is a [w*m][w*k]
	// omit data matrix since it is systematic
	matrix := GenerateCRSBitMatrix(k, m, w)
	if len(matrix) != w*w*m*k {
		t.Fatalf("len(matrix) = %d, want %d", len(matrix), w*w*m*k)
	}

	EncodeBitMatrix(k, m, w, matrix, data, coding, stripesize, packetsize)

	var (
		datai        = make([]int, k)
		decodematrix = make([]int, w*w*k*k)
		erasured     = make([]int, m+k)
	)
	// erasure the first m chunk
	for i := 0; i < m; i++ {
		erasured[i] = 1
	}

	DecodeBitMatrix(k, m, w, matrix, erasured, decodematrix, datai)
	survivors := append(data[m:], coding...)

	ndata := make([][]byte, k)
	for i := 0; i < k; i++ {
		ndata[i] = make([]byte, stripesize)
	}

	// recover one by one
	for i := 0; i < k; i++ {
		EncodeBitMatrix(k, 1, w, decodematrix[k*w*w*i:k*w*w*(i+1)], survivors, ndata[i:], stripesize, packetsize)
		if !bytes.Equal(data[i], ndata[i]) {
			t.Fatalf("data is not equal")
		}
	}
}
