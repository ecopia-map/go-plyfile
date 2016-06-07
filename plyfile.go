package plyfile

/*
#cgo LDFLAGS: -L${SRCDIR}/lib -lplyfile
#include <string.h>
#include "lib/ply.h"
*/
import "C"

import (
	"bytes"
	"unsafe"
	"encoding/binary"
	"fmt"
	"reflect"
)

// PLY definitions
const (
	PLY_ASCII     = 1 /* ascii PLY file */
	PLY_BINARY_BE = 2 /* binary PLY file, big endian */
	PLY_BINARY_LE = 3 /* binary PLY file, little endian */

	PLY_OKAY  = 0  /* ply routine worked okay */
	PLY_ERROR = -1 /* error in ply routine */

	/* scalar data types supported by PLY format */
	PLY_START_TYPE = 0
	PLY_CHAR       = 1
	PLY_SHORT      = 2
	PLY_INT        = 3
	PLY_UCHAR      = 4
	PLY_USHORT     = 5
	PLY_UINT       = 6
	PLY_FLOAT      = 7
	PLY_DOUBLE     = 8
	PLY_END_TYPE   = 9

	PLY_SCALAR = 0
	PLY_LIST   = 1
)

type CPlyProperty C.struct_PlyProperty
type PlyProperty struct {
	name          string /* property name */
	external_type int    /* file's data type */
	internal_type int    /* program's data type */
	offset        int    /* offset bytes of prop in a struct */

	is_list        int /* 1 = list, 0 = scalar */
	count_external int /* file's count type */
	count_internal int /* program's count type */
	count_offset   int /* offset byte for list count */
}

/* ToC converts a PlyProperty go structure to a PlyProperty C structure for passing to C functions */
func (prop *PlyProperty) ToC() CPlyProperty {
	var cprop CPlyProperty
	cprop.name = C.CString(prop.name)
	cprop.external_type = C.int(prop.external_type)
	cprop.internal_type = C.int(prop.internal_type)
	cprop.offset = C.int(prop.offset)
	cprop.is_list = C.int(prop.is_list)
	cprop.count_external = C.int(prop.count_external)
	cprop.count_internal = C.int(prop.count_internal)
	cprop.count_offset = C.int(prop.count_offset)
	return cprop
}

/* FromC converts a PlyProperty C structure (passed from a C function to the Go program) to a Go structure */
func (prop *PlyProperty) FromC(cprop CPlyProperty) {
	prop.name = C.GoString(cprop.name)
	prop.external_type = int(cprop.external_type)
	prop.internal_type = int(cprop.internal_type)
	prop.offset = int(cprop.offset)

	prop.is_list = int(cprop.is_list)
	prop.count_external = int(cprop.count_external)
	prop.count_internal = int(cprop.count_internal)
	prop.count_offset = int(cprop.count_offset)
}

type CPlyFile *C.struct_PlyFile
type CPlyElement *C.struct_PlyElement

/* PlyOpenForWriting creates a new PLY file (called filename) and writes in header information, specified by the other parameters. The returned PlyFile object is used to access header information and data stored in the PLY file.  */
func PlyOpenForWriting(filename string, nelems int, elem_names []string, file_type int, version *float32) CPlyFile {

	c_elem_names := make([]*C.char, nelems)
	for i := 0; i < nelems; i++ {
		c_elem_names[i] = C.CString(elem_names[i])
	}

	plyfile := C.ply_open_for_writing(C.CString(filename), C.int(nelems), &c_elem_names[0], C.int(file_type), (*C.float)(version))

	return plyfile
}

/* PlyOpenForReading opens a PLY file (specified by filename) and reads in the header information. The returned PlyFile object is used to access header information and data stored in the PLY file. */
func PlyOpenForReading(filename string) (CPlyFile, []string) {

	plyfile := C.ply_open_and_read_header(C.CString(filename))

	nelems := int(plyfile.nelems)

	elements := make([]CPlyElement, nelems)
	elem_names := make([]string, nelems)

	for i := 0; i < nelems; i++ {
		elements[i] = C.ply_get_element_by_index(plyfile, C.int(i))
		elem_names[i] = C.GoString(elements[i].name)
	}

	return plyfile, elem_names
}

/* PlyClose closes the open plyfile, specified by the CPlyFile object. Note that the PLY file memory is tracked by C, not by Go, and calling this function is necessary to free memory associated with the open PLY file. */
func PlyClose(plyfile CPlyFile) {
	C.ply_close(plyfile)
}

/* Writing Functions */

func PlyElementCount(plyfile CPlyFile, element_name string, nelems int) {
	C.ply_element_count(plyfile, C.CString(element_name), C.int(nelems))
}

func PlyDescribeProperty(plyfile CPlyFile, element_name string, prop PlyProperty) {
	propertyptr := prop.ToC()
	C.ply_describe_property(plyfile, C.CString(element_name), &propertyptr)
}

func PlyPutComment(plyfile CPlyFile, comment string) {
	C.ply_put_comment(plyfile, C.CString(comment))
}

func PlyPutObjInfo(plyfile CPlyFile, obj_info string) {
	C.ply_put_obj_info(plyfile, C.CString(obj_info))
}

func PlyHeaderComplete(plyfile CPlyFile) {
	C.ply_header_complete(plyfile)
}

func PlyPutElementSetup(plyfile CPlyFile, element_name string) {
	C.ply_put_element_setup(plyfile, C.CString(element_name))
}

func PlyPutElement(plyfile CPlyFile, element interface{}) {
	// write the passed in element to a buffer
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, element)
	if err != nil {
		panic(err)
	}
	element_bytes := buf.Bytes()

	// pass a pointer to the buffer
	C.ply_put_element(plyfile, unsafe.Pointer(&element_bytes[0]))
}

/* Reading Functions */
func PlyGetElementDescription(plyfile CPlyFile, element_name string) ([]PlyProperty, int, int) {
	var nelems int
	var nprops int

	cnelems := C.int(nelems)
	cnprops := C.int(nprops)

	cplist_ptr := C.ply_get_element_description(plyfile, C.CString(element_name), &cnelems, &cnprops)

	nprops = int(cnprops)

	// convert cplist_ptr to a go slice of pointers
	cplist_ptr_go := (*[1 << 30]*CPlyProperty)(unsafe.Pointer(cplist_ptr))[:nprops]

	// iterate through the slice of pointers, converting from CPlyProperty to PlyProperty
	plist := make([]PlyProperty, nprops)
	for i := 0; i < nprops; i++ {
		tmp := *cplist_ptr_go[i]
		plist[i].FromC(tmp)
	}

	return plist, int(cnelems), int(cnprops)
}

func PlyGetProperty(plyfile CPlyFile, elem_name string, prop PlyProperty) {
	cprop := prop.ToC()
	C.ply_get_property(plyfile, C.CString(elem_name), &cprop)
}

func PlyGetElement(plyfile CPlyFile, element interface{}, size uintptr) {

	var ptr C.char

	C.ply_get_element(plyfile, unsafe.Pointer(&ptr))
	//fmt.Println(&ptr)

	// convert the pointer into a number, so we can do pointer arithmetic
	ptrval := uintptr(unsafe.Pointer(&ptr))

	// convert the *C.char array into a byte slice
	var byteSlice = make([]byte, size)
	for i := 0; i < len(byteSlice); i++ {
		byteSlice[i] = byte(*(*C.char)(unsafe.Pointer(ptrval)))
		ptrval++
	}

	// get the pointer to the memory location of the input element
	elem_ptr := reflect.ValueOf(element).Elem().Addr().Interface()

	// copy the byte slice into the memory of the input element
	r := bytes.NewReader(byteSlice)
	err := binary.Read(r, binary.LittleEndian, elem_ptr)
	if err != nil {
		panic(err)
	}
}

func PlyGetComments(plyfile CPlyFile) []string {
	var cptr **C.char
	var cnum_comments C.int
	cptr = C.ply_get_comments(plyfile, &cnum_comments)

	num_comments := int(cnum_comments)

	// convert cptr to a go slice of pointers
	cstring_list := (*[1 << 30]*C.char)(unsafe.Pointer(cptr))[:num_comments]

	comments := make([]string, num_comments)

	for i := 0; i < num_comments; i++ {
		comments[i] = C.GoString(cstring_list[i])
	}

	return comments
}

func PlyGetObjInfo(plyfile CPlyFile) []string {
	var cptr **C.char
	var cnum_obj_info C.int
	cptr = C.ply_get_obj_info(plyfile, &cnum_obj_info)

	num_obj_info := int(cnum_obj_info)

	// convert cptr to a go slice of pointers
	cstring_list := (*[1 << 30]*C.char)(unsafe.Pointer(cptr))[:num_obj_info]

	obj_info := make([]string, num_obj_info)

	for i := 0; i < num_obj_info; i++ {
		obj_info[i] = C.GoString(cstring_list[i])
	}

	return obj_info
}

/* Util Functions */

/* PointerToByteSlice takes a memory location and stores it in a byte slice, which is returned. Note that this function is typically very unsafe in Go programs. Use caution! */
func PointerToByteSlice(ptr uintptr) []byte {
	size := unsafe.Sizeof(ptr)
	buf := make([]byte, size)
	switch size {
	case 4:
		binary.LittleEndian.PutUint32(buf, uint32(ptr))
	case 8:
		binary.LittleEndian.PutUint64(buf, uint64(ptr))
	default:
		panic(fmt.Sprintf("Error: unknown ptr size: %v", size))
	}
	return buf
}

/* ConvertByteSliceToInt32 takes a byte slice containing a memory location and a number of integer elements and returns an int32 array made up of the contents of the memory pointed to by the byte slice (to a maximum of num_elems elements). */
func ConvertByteSliceToInt32(bslice []byte, num_elems int) (ret []int32) {
	// create a buffer containing the memory location of interest
	buf := bytes.NewBuffer(bslice)

	// transcribe the memory location from the byte slice to a pointer
	var tmp uint32
	err := binary.Read(buf, binary.LittleEndian, &tmp)
	if err != nil {
		panic(err)
	}
	ptr := uintptr(tmp)

	// read the memory at ptr into a new byte slice
	var tmpSlice = make([]byte, len(bslice))
	for i := 0; i < len(tmpSlice); i++ {
		tmpSlice[i] = byte(*(*C.char)(unsafe.Pointer(ptr)))
		ptr++
	}

	// create a return slice and read the new byte slice into it
	ret = make([]int32, num_elems)
	buf = bytes.NewBuffer(tmpSlice)
	err = binary.Read(buf, binary.LittleEndian, ret)
	if err != nil {
		panic(err)
	}

	return
}
