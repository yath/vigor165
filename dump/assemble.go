// Assembles dump files into a binary and a bitmap of used blocks.
package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	baseOffset     = flag.Uint("base_offset", 0x60000000, "Base offset to subtract from dump addresses.")
	inputFileFlag  = flag.String("input_file", "dump-6000xxxx_6053xxxx.txt", "Input file to parse.")
	outputFileFlag = flag.String("output_file", "foo.bin", "Output filename.")
	mapFileFlag    = flag.String("map_file", "foo.map", "Bitmap of dirty bytes.")
)

func mustParseWord(s string) uint32 {
	ret, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		log.Panicf("Can't parse %q as 32-bit hex uint: %v", s, err)
	}

	return uint32(ret)
}

func nthbyte(w uint32, b uint) byte {
	return byte((w >> (b * 8)) & 0xff)
}

var chunkRE = regexp.MustCompile(`([0-9A-Fa-f]{8})\s*:\s*((?:\s+[0-9A-Fa-f]{8}){4})\s*`)

func parseLine(l string) (uint32, []byte) {
	sm := chunkRE.FindStringSubmatch(l)
	if len(sm) == 0 {
		if l != "" && !strings.HasPrefix(l, "DrayTek> sys mem ") {
			log.Printf("Unmatched line: %q", l)
		}
		return 0, nil
	}

	off := mustParseWord(sm[1])
	ret := make([]byte, 0, 16)
	for _, w := range strings.Fields(sm[2]) {
		wi := mustParseWord(w)
		ret = append(ret, nthbyte(wi, 3), nthbyte(wi, 2), nthbyte(wi, 1), nthbyte(wi, 0))
	}

	return off, ret
}

type memory struct {
	base        uint32
	dirty, data []byte
}

func (m *memory) allocate(maxRelAddr uint32) {
	if cap(m.dirty) < int((maxRelAddr/8)+1) {
		d := make([]byte, len(m.dirty), int(((maxRelAddr/8)+1)*2))
		copy(d, m.dirty)
		m.dirty = d
	}
	if len(m.dirty) < int(maxRelAddr/8)+1 {
		m.dirty = m.dirty[:(maxRelAddr/8)+1]
	}

	if cap(m.data) < int(maxRelAddr+1) {
		d := make([]byte, int((maxRelAddr+1)*2))
		copy(d, m.data)
		m.data = d
	}
	if len(m.data) < int(maxRelAddr+1) {
		m.data = m.data[:maxRelAddr+1]
	}
}

func (m *memory) set(off uint32, data []byte) {
	if off < m.base {
		log.Printf("Ignoring chunk with invalid offset %x", off)
		return
	}

	rel := off - m.base
	m.allocate(rel + uint32(len(data)))

	for i, b := range data {
		bo := rel + uint32(i)
		dby, dbi := bo/8, bo%8
		if m.dirty[dby]&(1<<dbi) != 0 {
			if m.data[bo] != b {
				log.Printf("Difference at offset %x: previously %02x, now %02x", off+uint32(i), m.data[bo], b)
			}
		}
		m.data[bo] = b
		m.dirty[dby] |= (1 << dbi)
	}
}

func main() {
	flag.Parse()
	if *inputFileFlag == "" {
		log.Fatal("--input_file is required.")
	}

	f, err := os.Open(*inputFileFlag)
	if err != nil {
		log.Fatalf("Can't open input file: %v", err)
	}

	m := &memory{base: uint32(*baseOffset)}

	log.Printf("Reading from %q, writing to %q", *inputFileFlag, *outputFileFlag)
	s := bufio.NewScanner(f)
	for s.Scan() {
		if err := s.Err(); err != nil {
			log.Fatalf("Can't read from input file: %v", err)
		}

		l := s.Text()
		off, data := parseLine(l)
		if len(data) == 0 {
			continue
		}

		m.set(off, data)
	}

	if *outputFileFlag != "" {
		if err := ioutil.WriteFile(*outputFileFlag, m.data, 0666); err != nil {
			log.Fatalf("Can't write 0x%x bytes: %v", len(m.data), err)
		}
	}

	if *mapFileFlag != "" {
		if err := ioutil.WriteFile(*mapFileFlag, m.dirty, 0666); err != nil {
			log.Fatalf("Can't write 0x%x bytes: %v", len(m.dirty), err)
		}
	}
}
