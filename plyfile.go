package plyfile

/*
#cgo LDFLAGS: -L${SRCDIR}/lib -lplyfile
#include "lib/ply.h"
*/
import "C"

import (
  "os"
  "unsafe"
  "bytes"
  //"encoding/gob"
  "encoding/binary"
  "fmt"
)

const (
  PLY_ASCII = 1        /* ascii PLY file */
  PLY_BINARY_BE = 2        /* binary PLY file, big endian */
  PLY_BINARY_LE = 3        /* binary PLY file, little endian */

  PLY_OKAY = 0           /* ply routine worked okay */
  PLY_ERROR = -1           /* error in ply routine */

  /* scalar data types supported by PLY format */
  PLY_START_TYPE = 0
  PLY_CHAR = 1
  PLY_SHORT = 2
  PLY_INT = 3
  PLY_UCHAR = 4
  PLY_USHORT = 5
  PLY_UINT = 6
  PLY_FLOAT = 7
  PLY_DOUBLE = 8
  PLY_END_TYPE = 9

  PLY_SCALAR = 0
  PLY_LIST = 1
)

type CPlyProperty C.struct_PlyProperty
type PlyProperty struct {
  name string               /* property name */
  external_type int         /* file's data type */
  internal_type int                    /* program's data type */
  offset int                           /* offset bytes of prop in a struct */

  is_list int                          /* 1 = list, 0 = scalar */
  count_external int                  /* file's count type */
  count_internal int                   /* program's count type */
  count_offset int                 /* offset byte for list count */
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
func (prop *PlyProperty) FromC(cprop CPlyProperty) {
  prop.name = C.GoString(cprop.name)
}


type PlyElement struct {
  name string                   /* element name */
  num int                      /* number of elements in this object */
  size int                     /* size of element (bytes) or -1 if variable */
  nprops int                   /* number of properties for this element */
  props []PlyProperty          /* list of properties in the file */
  store_prop string             /* flags: property wanted by user? */
  other_offset int             /* offset to un-asked-for props, or -1 if none*/
  other_size int               /* size of other_props structure */
}

type PlyOtherProp struct {   /* describes other properties in an element */
  name string                   /* element name */
  size int                     /* size of other_props */
  nprops int                   /* number of properties in other_props */
  props []PlyProperty          /* list of properties in other_props */
}

type PlyFile struct {        /* description of PLY file */
  FILE *os.File                     /* file pointer */
  file_type int                /* ascii or binary */
  version float64                /* version number of file */
  nelems int                   /* number of elements of object */
  elems []PlyElement           /* list of elements */
  num_comments int             /* number of comments */
  comments []string;              /* list of comments */
  num_obj_info int             /* number of items of object information */
  obj_info []string              /* list of object info items */
  which_elem PlyElement       /* which element we're currently writing */
}

type CPlyFile *C.struct_PlyFile

func PlyOpenForWriting(filename string, nelems int, elem_names []string, file_type int, version *float32) CPlyFile {

  c_elem_names := make([]*C.char, nelems)
  for i := 0; i < nelems; i++ {
    c_elem_names[i] = C.CString(elem_names[i])
  }

  plyfile := C.ply_open_for_writing(C.CString(filename), C.int(nelems), &c_elem_names[0], C.int(file_type), (*C.float)(version))

  return plyfile
}

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

func PlyClose(plyfile CPlyFile) {
  C.ply_close(plyfile)
}

func pointerToInt(ptr uintptr) []byte {
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

func copyByteSliceToArray(barray *[8]byte, bslice []byte) {
  for i := 0; i < len(bslice); i++ {
    barray[i] = bslice[i]
  }
}
