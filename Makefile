CFLAGS := -ggdb
MIPS_LD := /usr/mips-linux-gnu/lib

INFILE ?= v165_410_STD.all
OUTFILE ?= v165_410_STD_decompressed.bin

%.o: %.S
	mips-linux-gnu-gcc $(CFLAGS) -o $@ -c $<

main: main.c memcpy.o decompress.o
	mips-linux-gnu-gcc $(CFLAGS) -o $@ $^

.PHONY: run
run: main
	LD_LIBRARY_PATH=$(MIPS_LD) qemu-mips-static $(MIPS_LD)/ld.so.1 ./main $(INFILE) > $(OUTFILE)

.PHONY: clean
clean:
	rm -f main memcpy.o decompress.o
