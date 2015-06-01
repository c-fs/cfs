package erasure

import _ "github.com/cfs/Jerasure"
// #include "jerasure.h"
// #include "cauchy.h"
// #cgo CXXFLAGS: -std=c++11
// #cgo CFLAGS: -I../../Jerasure/internal/include
// #cgo CFLAGS: -I../../gf-complete/internal/include
// #cgo darwin LDFLAGS: -Wl,-undefined -Wl,dynamic_lookup
// #cgo !darwin LDFLAGS: -Wl,-unresolved-symbols=ignore-all
import "C"
import (
	"log"
	"reflect"
	"unsafe"
)

// void jerasure_bitmatrix_encode(int k, int m, int w, int *bitmatrix,char **data_ptrs, char **coding_ptrs, int size, int packetsize);
func EncodeBitMatrix(k, m, w int, bitmatrix []int, data, coding [][]byte, stripesize, packetsize int) {
	cdata := make([]*C.char, len(data))
	for i := range cdata {
		cdata[i] = (*C.char)(unsafe.Pointer(&data[i][0]))
	}
	ccoding := make([]*C.char, len(coding))
	for i := range ccoding {
		ccoding[i] = (*C.char)(unsafe.Pointer(&coding[i][0]))
	}
	cints := make([]C.int, len(bitmatrix))
	for i := range bitmatrix {
		cints[i] = C.int(bitmatrix[i])
	}

	C.jerasure_bitmatrix_encode(C.int(k), C.int(m), C.int(w), (*C.int)(unsafe.Pointer(&cints[0])),
		(**C.char)(unsafe.Pointer(&cdata[0])), (**C.char)(unsafe.Pointer(&ccoding[0])), C.int(stripesize), C.int(packetsize))
}

// int jerasure_make_decoding_bitmatrix(int k, int m, int w, int *matrix, int *erased, int *decoding_matrix, int *dm_ids)
func DecodeBitMatrix(k, m, w int, matrix, erased, decodematrix, data []int) {
	cmatrix := make([]C.int, len(matrix))
	for i := range matrix {
		cmatrix[i] = C.int(matrix[i])
	}
	cerased := make([]C.int, len(erased))
	for i := range erased {
		cerased[i] = C.int(erased[i])
	}
	cdecodematrix := make([]C.int, len(decodematrix))
	cdata := make([]C.int, len(data))

	C.jerasure_make_decoding_bitmatrix(C.int(k), C.int(m), C.int(w), (*C.int)(unsafe.Pointer(&cmatrix[0])),
		(*C.int)(unsafe.Pointer(&cerased[0])), (*C.int)(unsafe.Pointer(&cdecodematrix[0])), (*C.int)(unsafe.Pointer(&cdata[0])))

	for i := range decodematrix {
		decodematrix[i] = int(cdecodematrix[i])
	}
	for i := range data {
		data[i] = int(cdata[i])
	}
}

// int jerasure_invert_bitmatrix(int *mat, int *inv, int rows);
func InvertBitMatrix(matrix, invert []int, rows int) {
	cintsin := make([]C.int, len(matrix))
	for i := range matrix {
		cintsin[i] = C.int(matrix[i])
	}
	cintsout := make([]C.int, len(matrix))
	C.jerasure_invert_bitmatrix((*C.int)(unsafe.Pointer(&cintsin[0])), (*C.int)(unsafe.Pointer(&cintsout[0])), C.int(rows))

	for i := range cintsout {
		invert[i] = int(cintsout[i])
	}
}

// int *cauchy_original_coding_matrix(int k, int m, int w);
// int *jerasure_matrix_to_bitmatrix(k, m, w, matrix);
func GenerateCRSBitMatrix(k, m, w int) []int {
	gom := make([]int, 0)

	matrix := C.cauchy_original_coding_matrix(C.int(k), C.int(m), C.int(w))
	if matrix == nil {
		log.Fatal("erasure: cannot make matrix")
	}
	bitmatrix := C.jerasure_matrix_to_bitmatrix(C.int(k), C.int(m), C.int(w), matrix)
	if bitmatrix == nil {
		log.Fatal("erasure: cannot make matrix")
	}
	length := int(k * m * w * w)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(bitmatrix)),
		Len:  length,
		Cap:  length,
	}
	matrixSlice := *(*[]C.int)(unsafe.Pointer(&hdr))

	for _, e := range matrixSlice {
		gom = append(gom, int(e))
	}

	return gom
}
