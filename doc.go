/*
Copyright 2016 Alex Baden

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package plyfile provides functions for reading and writing PLY files. The package uses the C plyfile library, originally developed by Greg Turk and released in February 1994. All Go code is provided under the Apache 2.0 license. Greg Turk's code has a separate license (see lib folder).

Disclaimer

There are probably cleaner, more go-centric ways to write this code. However, this project accomplished several goals. First and foremost, it was a great way to learn more about the relationship between Go programs and C programs when using cgo. The plyfile C library uses a lot of dynamic memory allocation and other C tricks which are not exactly Go-friendly. Most of the Go code copies C memory to Go memory, which is not particularly efficient. Future versions of this library will likely convert some of the original C code to native Go code. But, for a first Go project, this was a great one; a lot of learning happened!

Acknowledgements

A very big thanks is owed to Greg Turk for releasing his original plyfile code. Preserving the flexibility of his original package was a major goal for this project, and none of it would be possible without his well written and well documented C library and accompanying test file.

Installation

Run
  go get github.com/alexbaden/go-plyfile

The install will fail with the following error:
  # github.com/alexbaden/go-plyfile
  /usr/bin/ld: cannot find -lplyfile
  collect2: error: ld returned 1 exit status# github.com/alexbaden/go-plyfile
  /usr/bin/ld: cannot find -lplyfile
  collect2: error: ld returned 1 exit status

That's fine, you just need to compile the C code with the included makefile. Head over to the github.com/alexbaden/go-plyfile/lib
directory:
  cd $GOPATH/src/github.com/alexbaden/go-plyfile/lib
Compile the C code using make.

Now you can run go build and go install from the parent directory:
  $GOPATH/src/github.com/alexbaden/go-plyfile

From there, you should be good to go. But, just to make sure, run go test and verify the two tests (Write and Read) pass.

Basics

The structure of Turk's original code is preserved. The plyfile package makes heavy use of cgo to interface with Turk's code and use his original functions. Note that this means some memory is not tracked by the Go garbage collector. It is critical to properly close PLY files using the PlyClose function!

A comparison of ply_test.go and lib/plytest.c shows nearly identical function calls. Thus, porting a C program that uses Turk's plyfile library should be relatively straightforward. The ply_test.go file shows the process of both writing and reading a PLY file, including how to open a file, describe properties, and write and read data. The basics of writing and reading are also described below.

Writing PLY Files

A full example of writing PLY files is available in ply_test.go. Briefly, the following steps are required:

First, some basic variables must be declared and defined. Namely, a string slice of element names (to define each element to be written to the PLY file). The type of PLY file (ASCII or Binary) must also be specified. In our test example, we generate some data, and set the PLY properties. Properties could be declared in the function body, but using a setting function allows us to re-declare properties in both the Writing and Reading function with little work.

Second, we open the PLY file for writing using the aptly named PlyOpenForWriting function. This function allocates C memory, creates a plyfile object (which we use to track our PLY file through writing), and returns the PLY file version (which will likely always be 1.0). Next, we describe element count, properties, and (optionally) add object info and/or comments. After all this, we can call PlyHeaderComplete, which finalizes the header and writes it to disk.

Third, we setup each element and write the elements to disk. Once elements are written, we can close the PLY file using PlyClose, which frees all allocated C memory and flushes the file to disk (also closing the open file handler). That's it!

Reading PLY Files

A full example of reading PLY files is also available in ply_test.go. Briefly, the following steps are required:

First, we declare our PLY properties and open the PLY file for reading using PlyOpenForReading. PlyOpenForReading returns a CPlyFile object, for tracking our PLY file in future functions, and a list of element names.

Second, we iterate through the returned element names. For each element name, we get the number of elements and number of properties as well as a slice of PLYProperty objects. Next, we allocate Go memory to store the incoming element data, and read the elements. We repeat for each element.

Finally, comments and object info are read from the PLY file, and returned as a slice of strings using the respective reading function.

A note about elements with list properties

The currently element with list property implementation (see Face in ply_test.go) likely needs to be adjusted. The Verts [16]byte array is used to store a pointer to the vertex_indices, and stores a pointer to the vertex_indices on return. Using a 32-bit or 64-bit integer may be better, and will possibly be changed in a future release. However, the basic idea is as follows:

When using a list, we expect to have a dynamic number elements for a given property. The C code expects a pointer to the list of elements, but Go won't package a struct containing pointers. So, we cheat and write the memory location into a byte slice, then pass that byte slice to the C code. As long as we're careful not to garbage collect the memory containing the list of elements while the C code is running, there are no issues.

When reading the list back in, we get a byte slice containing the memory location of the list data we are interested in. ConvertByteSliceToInt32 allows us to write the data stored in C memory, pointed to by the memory address written in the byte slice, to a Go slice. Again, if we're careful not to garbage collect or free memory before reading and converting to a Go object, this method works fine. Since most PLY reading or writing happens in a single function, memory issues should occur infrequently. However, this implementation is a weak point of the current program, and will likely need to be revised.

*/
package plyfile
