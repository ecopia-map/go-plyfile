package plyfile

/*
#cgo LDFLAGS: -L${SRCDIR}/lib -lplyfile
#include <string.h>
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
  "reflect"
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
  prop.external_type = int(cprop.external_type)
  prop.internal_type = int(cprop.internal_type)
  prop.offset = int(cprop.offset)

  prop.is_list = int(cprop.is_list)
  prop.count_external = int(cprop.count_external)
  prop.count_internal = int(cprop.count_internal)
  prop.count_offset = int(cprop.count_offset)
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
type CPlyElement *C.struct_PlyElement

func PlyOpenForWriting(filename string, nelems int, elem_names []string, file_type int, version *float32) CPlyFile {

  c_elem_names := make([]*C.char, nelems)
  for i := 0; i < nelems; i++ {
    c_elem_names[i] = C.CString(elem_names[i])
  }

  plyfile := C.ply_open_for_writing(C.CString(filename), C.int(nelems), &c_elem_names[0], C.int(file_type), (*C.float)(version))

  return plyfile
}

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
func PlyGetElementDescription(plyfile CPlyFile, element_name string) ([]PlyProperty, int) {
  var nelems int
  var nprops int

  cnelems := C.int(nelems)
  cnprops := C.int(nprops)

  cplist_ptr := C.ply_get_element_description(plyfile, C.CString(element_name), &cnelems, &cnprops)

  nprops = int(cnprops)

  // convert cplist_ptr to a go slice of pointers
  cplist_ptr_go := (*[1<<30]*CPlyProperty)(unsafe.Pointer(cplist_ptr))[:nprops]

  // iterate through the slice of pointers, converting from CPlyProperty to PlyProperty
  plist := make([]PlyProperty, nprops)
  for i := 0; i < nprops; i++ {
    tmp := *cplist_ptr_go[i]
    plist[i].FromC(tmp)
  }

  return plist, int(cnelems)
}

func PlyGetProperty(plyfile CPlyFile, elem_name string, prop PlyProperty) {
  cprop := prop.ToC()
  C.ply_get_property(plyfile, C.CString(elem_name), &cprop)
}

func PlyGetElement(plyfile CPlyFile, element interface {}, size uintptr) {

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

  /*
  // create a new temporary element
  elem_type := reflect.TypeOf(element)
  fmt.Println(elem_type)
  //elem_tmp := reflect.New(elem_type).Elem().Addr().Interface()
  elem_tmp := reflect.New(elem_type).Elem().Addr() //.Elem().Addr().Interface()
  fmt.Println(elem_tmp)
  fmt.Printf("type: %T\n", elem_tmp)
  */

  // copy byte slice into temporary element
  elem_ptr := reflect.ValueOf(element).Elem().Addr().Interface() // should be elem_ptr

  r := bytes.NewReader(byteSlice)
  err := binary.Read(r, binary.LittleEndian, elem_ptr)
  if err != nil {
    panic(err)
  }

  // set input element to be temporary element (perform deep copy)
  /*
  for i := 0; i < reflect.ValueOf(element).NumField(); i++ {
    reflect.ValueOf(element).Ptr().Field(i).Set( reflect.ValueOf(elem_tmp).Field(i) )
  }
  */
  //fmt.Printf("1", reflect.ValueOf(element).Interface())


  //reflect.ValueOf(element) = elem_tmp



  //var v reflect.TypeOf(element)
  //fmt.Printf("%T\n\n", v)

  /* OLDDD!!!!!!!!!!!!!! */
  /*
  element_value := reflect.ValueOf(&element).Elem()
  fmt.Println(element_value.CanSet())


  fmt.Println(element_value.Kind())
  fmt.Println("before", element_value)
  err := binary.Read(r, binary.LittleEndian, element_value)
  if err != nil {
    panic(err)
  }
  fmt.Println("after", element_value)

  fmt.Println(element)
  */
  /*
  t := reflect.ValueOf(element)
  fmt.Printf("type %T\n\n", t)
  fmt.Println("t is ", t)
  */

  /*
  // copy byte slice into


  fmt.Println(byteSlice)


  t := reflect.New(reflect.TypeOf(element)).Elem().Interface()
  fmt.Println(t)
  fmt.Printf("Type: %T\n\n", t)

  element = t
  */

  /*
  buf := new(bytes.Buffer)
  for i := 0; i < int(size); i++ {
    buf.WriteByte(ptr[i])
  }

  */
  //fmt.Println(size)

  ///fmt.Println(element)

}

/* misc functions */

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
