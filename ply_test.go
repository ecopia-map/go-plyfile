package plyfile

import (
  "fmt"
  "testing"
  "unsafe"
)

/* Exported Fields: All struct fields must be exported (capitalized) for use in the plyfile package! */
type Vertex struct {
  X, Y, Z float32
}

type Face struct {
  Intensity byte
  Nverts byte
  Verts [8]byte // maximum size array
}

type VertexIndices [4]int32

func GenerateVertexFaceData() (verts []Vertex, faces []Face, vertex_indices []VertexIndices) {
  verts = make([]Vertex, 8)
  faces = make([]Face, 6)

  verts[0] = Vertex{0.0, 0.0, 0.0}
  verts[1] = Vertex{1.0, 0.0, 0.0}
  verts[2] = Vertex{1.0, 1.0, 0.0}
  verts[3] = Vertex{0.0, 1.0, 0.0}
  verts[4] = Vertex{0.0, 0.0, 1.0}
  verts[5] = Vertex{1.0, 0.0, 1.0}
  verts[6] = Vertex{1.0, 1.0, 1.0}
  verts[7] = Vertex{0.0, 1.0, 1.0}

  vertex_indices = make([]VertexIndices, 6)
  vertex_indices[0] = VertexIndices{0, 1, 2, 3}
  vertex_indices[1] = VertexIndices{7, 6, 5, 4}
  vertex_indices[2] = VertexIndices{0, 4, 5, 1}
  vertex_indices[3] = VertexIndices{1, 5, 6, 2}
  vertex_indices[4] = VertexIndices{2, 6, 7, 3}
  vertex_indices[5] = VertexIndices{3, 7, 4, 0}

  nil_array := [8]byte{0,0,0,0,0,0,0,0}

  faces[0] = Face{'\001', 4, nil_array}
  faces[1] = Face{'\004', 4, nil_array}
  faces[2] = Face{'\010', 4, nil_array}
  faces[3] = Face{'\020', 4, nil_array}
  faces[4] = Face{'\144', 4, nil_array}
  faces[5] = Face{'\377', 4, nil_array}

  for i := 0; i < 6; i++ {
    copyByteSliceToArray(&faces[i].Verts, pointerToInt(uintptr(unsafe.Pointer(&vertex_indices[i]))))
  }

  return verts, faces, vertex_indices
}

func SetPlyProperties() (vert_props []PlyProperty, face_props []PlyProperty) {
  vert_props = make([]PlyProperty, 3)
  vert_props[0] = PlyProperty{"x", PLY_FLOAT, PLY_FLOAT, int(unsafe.Offsetof(Vertex{}.X)), 0, 0, 0, 0}
  vert_props[1] = PlyProperty{"y", PLY_FLOAT, PLY_FLOAT, int(unsafe.Offsetof(Vertex{}.Y)), 0, 0, 0, 0}
  vert_props[2] = PlyProperty{"z", PLY_FLOAT, PLY_FLOAT, int(unsafe.Offsetof(Vertex{}.Z)), 0, 0, 0, 0}

  face_props = make([]PlyProperty, 2)
  face_props[0] = PlyProperty{"intensity", PLY_UCHAR, PLY_UCHAR, int(unsafe.Offsetof(Face{}.Intensity)), 0, 0, 0, 0}
  face_props[1] = PlyProperty{"vertex_indices", PLY_INT, PLY_INT, int(unsafe.Offsetof(Face{}.Verts)), 1, PLY_UCHAR, PLY_UCHAR, int(unsafe.Offsetof(Face{}.Nverts))}

  return vert_props, face_props

}

func TestWritePly(t *testing.T) {
  elem_names := make([]string, 2)
  elem_names[0] = "vertex"
  elem_names[1] = "face"
  var nelems int
  nelems = 2
  var version float32

  fmt.Println("Writing PLY file 'test.ply'...")

  plyfile := PlyOpenForWriting("test.ply", nelems, elem_names, PLY_ASCII, &version)

  // Note that we don't need a variable for vertex_indices, but we do need to return vertex_indices. Otherwise, the garbage collector will remove them once GenerateVertexFaceData() returns.
  verts, faces, _ := GenerateVertexFaceData()
  vert_props, face_props := SetPlyProperties()

  // Describe vertex properties
  PlyElementCount(plyfile, "vertex", len(verts))
  PlyDescribeProperty(plyfile, "vertex", vert_props[0])
  PlyDescribeProperty(plyfile, "vertex", vert_props[1])
  PlyDescribeProperty(plyfile, "vertex", vert_props[2])

  // Describe face properties
  PlyElementCount(plyfile, "face", len(faces))
  PlyDescribeProperty(plyfile, "face", face_props[0])
  PlyDescribeProperty(plyfile, "face", face_props[1])

  // Add a comment and an object information field
  PlyPutComment(plyfile, "go author: Alex Baden, c author: Greg Turk");
  PlyPutObjInfo(plyfile, "random information");

  // Finish writing header
  PlyHeaderComplete(plyfile)

  // Setup and write vertex elements
  PlyPutElementSetup(plyfile, "vertex")
  for _, vertex := range verts {
    PlyPutElement(plyfile, vertex)
  }

  // Setup and write face elements
  PlyPutElementSetup(plyfile, "face")
  for _, face := range faces {
    PlyPutElement(plyfile, face)
  }

  // close the PLY file
  PlyClose(plyfile)

  fmt.Println("Wrote PLY file.")
}

func TestReadPLY(t *testing.T) {
  fmt.Println("Reading PLY file 'test.ply'...")

  // setup properties
  vert_props, face_props := SetPlyProperties()

  // open the PLY file for reading
  plyfile, elem_names := PlyOpenForReading("test.ply")

  // print what we found out about the file
  fmt.Printf("version: %f\n", plyfile.version)
  fmt.Printf("file_type: %d\n", plyfile.file_type)

  // read each element
  for _, name := range elem_names {

    // get element description
    plist, num_elems, num_props := PlyGetElementDescription(plyfile, name)

    // print the name of the element, for debugging
    fmt.Println("element", name, num_elems)

    if name == "vertex" {

      // create a list to store all vertices
      vlist := make([]Vertex, num_elems)

      /* set up for getting vertex elements
      specifically, we are ensuring the 3 desirable properties of a vertex (x,,z) are returned.
      */
      PlyGetProperty(plyfile, name, vert_props[0])
      PlyGetProperty(plyfile, name, vert_props[1])
      PlyGetProperty(plyfile, name, vert_props[2])

      // grab vertex elements
      for i := 0; i < num_elems; i++ {
        PlyGetElement(plyfile, &vlist[i], unsafe.Sizeof(Vertex{}))

        // print out vertex for debugging
        /* TODO UNCOMMENT ME!!!
        fmt.Printf("vertex: %g %g %g\n", vlist[i].X, vlist[i].Y, vlist[i].Z)
        */

      }
    } else if name == "face" {
      // create a list to hold all face elements
      flist := make([]Face, num_elems)

      /* set up for getting face elements (See above) */
      PlyGetProperty(plyfile, name, face_props[0])
      PlyGetProperty(plyfile, name, face_props[1])

      // grab face elements
      for i := 0; i < num_elems; i++ {
        PlyGetElement(plyfile, &flist[i], unsafe.Sizeof(Face{}))

        // print out faces for debugging
        /*
        fmt.Printf("face: %d, list = ", flist[i].Intensity)

        for j := 0; j < int(flist[i].Nverts); j++ {
          fmt.Printf("%d ", flist[i].Verts[j])
        }
        fmt.Printf("\n")
        */
        fmt.Println(flist[i])
      }


    }

    for i := 0; i < num_props; i++ {
      fmt.Println("property", plist[i].name)
    }

  }

  // comments TODO

  // object info TODO


  // close the PLY file
  PlyClose(plyfile)

}
