#define _GNU_SOURCE
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <errno.h>
#include <unistd.h>
#include <string.h>
#include <stdlib.h>
#include <stdint.h>
#include <arpa/inet.h>

extern int my_memcpy(char *dst, char *src, int len);
extern int decompress(char *src, char *dst, int srclen, int dstlen, void *wrkmem);

#define print(fmt, args...) do { fprintf(stderr, fmt "\n", ##args); fflush(stderr); } while(0)
#define fail(fmt, args...) do { print(fmt, ##args); _exit(1); } while(0)
#define pfail(fmt, args...) do { fail(fmt ": %s", ##args, strerror(errno)); } while(0)

const char magic[] = { 0xaa, 0x1d, 0x7f, 0x50 };

char *readall(const char *filename, int *size) {
    struct stat st;
    if (stat(filename, &st) < 0)
        pfail("Unable to stat %s", filename);

    *size = st.st_size;
    char *buf = malloc(*size);
    if (!buf)
        pfail("Unable to allocate %d bytes", *size);

    int fd = open(filename, O_RDONLY);
    if (fd < 0)
        pfail("Unable to open %s", filename);

    int pos = 0;
    int n;
    do {
        n = read(fd, buf+pos, *size-pos);
        if (n < 0)
            pfail("Unable to read from %s", filename);

        pos += n;
    } while (n > 0 && pos < *size);

    return buf;
}

uint32_t uint_at(const unsigned char *buf) {
    return (buf[3] << 24) |
           (buf[2] << 16) |
           (buf[1] << 8)  |
           buf[0];
}

int main(int argc, char **argv) {
    if (argc < 2)
        fail("Usage: %s <filename>", argv[0]);

    int size;
    char *in = readall(argv[1], &size);

    print("read %d bytes", size);

    char *p = memmem(in, size, magic, sizeof(magic));
    if (!p)
        fail("Header not found");
    print("Header found at 0x%x", p-in);
    p += sizeof(magic);

    char buf[0x100000];
    while (p < in+size) {
        uint32_t cs = uint_at(p);
        print("chunk size: 0x%x", cs);
        p += sizeof(uint32_t);
        int len = decompress(p, buf, cs, sizeof(buf), 0);
        if (len < 0) {
            print("Unable to decompress data at offset 0x%x, ret = %d", p-in, len);
            p = memmem(p, size-(p-in), magic, sizeof(magic));
            if (!p) {
                print("No further header found, exiting");
                break;
            } else {
                print("New header found at offset 0x%x", p-in);
                p += sizeof(magic);
                continue;
            }
        }

        print("Inflated %d bytes at 0x%x to %d bytes", cs, p-in, len);

        do {
            int wlen = write(STDOUT_FILENO, buf, len);
            if (wlen < 0)
                pfail("Unable to write %d bytes to stdout", len);
            len -= wlen;
        } while (len > 0);

        p += cs;
    }

    return 0;
}
