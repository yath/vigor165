// Dumps memory via UART. Hangs often.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/jacobsa/go-serial/serial"
)

var (
	serialPortFlag = flag.String("serial_port", "/dev/ttyUSB0", "Serial port to use.")
	baudRateFlag   = flag.Uint("baud_rate", 115200, "Baud rate to use.")

	startAddrFlag = flag.Uint("start_addr", 0x60000000, "Start address to dump.")
	lengthFlag    = flag.Uint("length", 0, "Number of bytes to dump, 0 for infinite.")
	blockSizeFlag = flag.Uint("block_size", 0x100, "Number of bytes to dump at once. Higher values risk watchdog timeouts.")
)

func openSerialPort(portName string, baudRate uint) (io.ReadWriteCloser, error) {
	oo := serial.OpenOptions{
		PortName:              *serialPortFlag,
		BaudRate:              *baudRateFlag,
		DataBits:              8,
		StopBits:              1,
		ParityMode:            serial.PARITY_NONE,
		RTSCTSFlowControl:     false,
		MinimumReadSize:       1,
		InterCharacterTimeout: 1, // 1/10s
	}
	log.Printf("Opening serial port with options %+v", oo)

	s, err := serial.Open(oo)
	if err != nil {
		return nil, fmt.Errorf("can't open serial port: %v", err)
	}

	return s, nil
}

type state int

const (
	initialState state = iota
	mainMenuState
	dumpingState
)

func main() {
	flag.Parse()

	sp, err := openSerialPort(*serialPortFlag, *baudRateFlag)
	if err != nil {
		log.Fatalf("Can't open serial port: %v", err)
	}

	fmt.Fprintf(sp, "\n") // Trigger some output.

	currAddr := *startAddrFlag
	st := mainMenuState
	s := bufio.NewScanner(sp)
F:
	for s.Scan() {
		if err := s.Err(); err != nil {
			log.Fatalf("Reading from serial port: %v", err)
		}

		line := s.Text()
		if line != "" {
			log.Printf("received: %q", line)
		}

	S:
		switch st {
		case mainMenuState:
			log.Printf("In main menu, issuing dump request for 0x%08x", currAddr)
			fmt.Fprintf(sp, "m%x\n%x\n\n", currAddr, *blockSizeFlag)
			currAddr += *blockSizeFlag
			st = dumpingState

		case dumpingState:
			if strings.Index(line, " Main Menu ") != -1 {
				st = mainMenuState
				break S
			}

			if line != "" {
				fmt.Println(line)
			}
		}

		if *lengthFlag > 0 && currAddr >= (*startAddrFlag+*lengthFlag) {
			break F
		}
	}
	log.Print("EOF")
}
