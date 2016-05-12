package plyfile

import (
  "fmt"
  "testing"
  "unsafe"
)

type Vertex struct {
  x, y, z float32
}

type Face struct {
  intensity byte
  nverts int32
  verts [4]int32
}

type VertexIndices [4]int32

func GenerateVertexFaceData() (verts []Vertex, faces []Face) {
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

  vertex_indices := make([]VertexIndices, 6)
  vertex_indices[0] = VertexIndices{0, 1, 2, 3}
  vertex_indices[1] = VertexIndices{7, 6, 5, 4}
  vertex_indices[2] = VertexIndices{0, 4, 5, 1}
  vertex_indices[3] = VertexIndices{1, 5, 6, 2}
  vertex_indices[4] = VertexIndices{2, 6, 7, 3}
  vertex_indices[5] = VertexIndices{3, 7, 4, 0}

  faces[0] = Face{'\001', 4, vertex_indices[0]}
  faces[1] = Face{'\004', 4, vertex_indices[1]}
  faces[2] = Face{'\010', 4, vertex_indices[2]}
  faces[3] = Face{'\020', 4, vertex_indices[3]}
  faces[4] = Face{'\144', 4, vertex_indices[4]}
  faces[5] = Face{'\377', 4, vertex_indices[5]}

  return verts, faces
}

func SetPlyProperties() (vert_props []PlyProperty, face_props []PlyProperty) {
  vert_props = make([]PlyProperty, 3)
  vert_props[0] = PlyProperty{"x", PLY_FLOAT, PLY_FLOAT, int(unsafe.Offsetof(Vertex{}.x)), 0, 0, 0, 0}
  vert_props[1] = PlyProperty{"y", PLY_FLOAT, PLY_FLOAT, int(unsafe.Offsetof(Vertex{}.y)), 0, 0, 0, 0}
  vert_props[2] = PlyProperty{"z", PLY_FLOAT, PLY_FLOAT, int(unsafe.Offsetof(Vertex{}.z)), 0, 0, 0, 0}

  face_props = make([]PlyProperty, 2)
  face_props[0] = PlyProperty{"intensity", PLY_UCHAR, PLY_UCHAR, int(unsafe.Offsetof(Face{}.intensity)), 0, 0, 0, 0}
  face_props[1] = PlyProperty{"vertex_indices", PLY_INT, PLY_INT, int(unsafe.Offsetof(Face{}.verts)), 1, PLY_UCHAR, PLY_UCHAR, int(unsafe.Offsetof(Face{}.nverts))}

  return vert_props, face_props

}

func TestWritePly(t *testing.T) {
  elem_names := make([]string, 2)
  elem_names[0] = "vertex"
  elem_names[1] = "face"
  var nelems int
  nelems = 2
  var version float32

  plyfile := PlyOpenForWriting("test.ply", nelems, elem_names, PLY_ASCII, &version)

  verts, faces := GenerateVertexFaceData()
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

  fmt.Println("Done")


}
