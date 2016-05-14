all:
	$(CC) -g -c -o plyfile.o plyfile.c
	ar rcs libplyfile.so plyfile.o 
bin:
	$(CC) -g -o plytest plytest.c plyfile.c
clean:
	rm plytest.o
	rm plyfile.o
	rm plytest
