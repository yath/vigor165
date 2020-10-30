CC := mips-linux-gnu-gcc
QEMU := qemu-mips-static
CFLAGS := -ggdb
MIPS_LIB_PREFIX := /usr/mips-linux-gnu

INFILE ?= v165_410_STD.all
OUTPREFIX ?= out_

%.o: %.S
	$(CC) $(CFLAGS) -o $@ -c $<

main: main.c decompress.o
	$(CC) $(CFLAGS) -o $@ $^

.PHONY: run
run: main
	$(QEMU) -L $(MIPS_LIB_PREFIX) ./main $(INFILE) $(OUTPREFIX)

.PHONY: clean
clean:
	rm -f main decompress.o
