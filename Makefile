CC := mips-linux-gnu-gcc
QEMU := qemu-mips-static
CFLAGS := -ggdb
MIPS_LD := /usr/mips-linux-gnu/lib

INFILE ?= v165_410_STD.all
OUTFILE ?= v165_410_STD_decompressed.bin

%.o: %.S
	$(CC) $(CFLAGS) -o $@ -c $<

main: main.c memcpy.o decompress.o
	$(CC) $(CFLAGS) -o $@ $^

.PHONY: run
run: main
	LD_LIBRARY_PATH=$(MIPS_LD) $(QEMU) $(MIPS_LD)/ld.so.1 ./main $(INFILE) > $(OUTFILE)

.PHONY: clean
clean:
	rm -f main memcpy.o decompress.o
