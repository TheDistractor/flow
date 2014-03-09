package rfdata

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"os"
	"strings"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["ReadTextFile"] = func() flow.Worker { return &ReadTextFile{} }
	flow.Registry["IntelHexToBin"] = func() flow.Worker { return &IntelHexToBin{} }
	flow.Registry["BinaryFill"] = func() flow.Worker { return &BinaryFill{} }
	flow.Registry["CalcCrc"] = func() flow.Worker { return &CalcCrc{} }
}

// ReadTextFile takes strings and replaces them by the lines of that file.
// Inserts <open> and <close> tags before doing so. Registers as "ReadTextFile".
type ReadTextFile struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start picking up strings and injecting the text lines instead.
func (w *ReadTextFile) Run() {
	for m := range w.In {
		if name, ok := m.(string); ok {
			file, err := os.Open(name)
			if err != nil {
				panic(err)
			}
			scanner := bufio.NewScanner(file)
			w.Out.Send(flow.Tag{"<open>", name})
			for scanner.Scan() {
				w.Out.Send(scanner.Text())
			}
			w.Out.Send(flow.Tag{"<close>", name})
		} else {
			w.Out.Send(m)
		}
	}
}

// IntelHexToBin takes lines of text and converts it to one large []byte value.
// Inserts an <addr> tag before the data. Registers as "IntelHexToBin".
type IntelHexToBin struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start reading ":..." lines. Anything else cause the data to be flushed out.
func (w *IntelHexToBin) Run() {
	var buf bytes.Buffer
	for m := range w.In {
		if t, ok := m.(string); ok {
			if strings.HasPrefix(t, ":") {
				b, err := hex.DecodeString(t[1:])
				if err != nil {
					panic(err)
				}
				// TODO: probably doesn't handle hex files over 64 KB
				if b[3] == 0 {
					if buf.Len() == 0 {
						addr := int(b[1]) + int(b[2])<<8
						w.Out.Send(flow.Tag{"<addr>", addr})
					}
					buf.Write(b[4 : 4+b[0]])
				}
			}
		} else {
			if buf.Len() > 0 {
				w.Out.Send(buf.Bytes())
				buf.Reset()
			}
			w.Out.Send(m)
		}
	}
	if buf.Len() > 0 {
		w.Out.Send(buf.Bytes())
	}
}

// Take binary data and make sure it is filled to a specified multiple.
// Registers as "BinaryFill".
type BinaryFill struct {
	flow.Work
	In  flow.Input
	Len flow.Input
	Out flow.Output
}

// Start looking for []byte values, everything else is passed through unchanged.
func (w *BinaryFill) Run() {
	if n, ok := <-w.Len; ok {
		for m := range w.In {
			if data, ok := m.([]byte); ok {
				for n.(int) > 0 && len(data)%n.(int) != 0 {
					data = append(data, 0xFF)
				}
				m = data
			}
			w.Out.Send(m)
		}
	}
}

var crcTable = []uint16{
	0x0000, 0xCC01, 0xD801, 0x1400, 0xF001, 0x3C00, 0x2800, 0xE401,
	0xA001, 0x6C00, 0x7800, 0xB401, 0x5000, 0x9C01, 0x8801, 0x4400,
}

func calculateCrc(buf []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, data := range buf {
		crc = crc>>4 ^ crcTable[crc&0x0F] ^ crcTable[data&0x0F]
		crc = crc>>4 ^ crcTable[crc&0x0F] ^ crcTable[data>>4]
	}
	return crc
}

// CalcCrc takes []byte values and adds its CRC as <crc16> tag after it.
// Registers as "CalcCrc".
type CalcCrc struct {
	flow.Work
	In  flow.Input
	Len flow.Input
	Out flow.Output
}

// Start looking for []byte values, everything else is passed through unchanged.
func (w *CalcCrc) Run() {
	for m := range w.In {
		if data, ok := m.([]byte); ok {
			w.Out.Send(m)
			m = flow.Tag{"<crc16>", calculateCrc(data)}
		}
		w.Out.Send(m)
	}
}
