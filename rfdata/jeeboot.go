package rfdata

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["JeeBoot"] = func() flow.Worker { return &JeeBoot{} }
}

// This takes JeeBoot requests and returns the desired information as reply.
type JeeBoot struct {
	flow.Work
	In  flow.Input
	Out flow.Output

	dev    string
	config Config
	fw     map[uint16]Firmware
}

// Start decoding JeeBoot packets.
func (w *JeeBoot) Run() {
	for m := range w.In {
		if req, ok := m.([]byte); ok {
			fmt.Println("JB request", len(req))
			reply := w.respondToRequest(req)
			if reply != nil {
				cmd := convertReplyToCmd(reply)
				fmt.Println("JB reply", cmd)
				w.Out.Send(reply)
			}
		}
	}
}

func convertReplyToCmd(reply interface{}) string {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, reply)
	flow.Check(err)
	cmd := strings.Replace(fmt.Sprintf("%v", buf.Bytes()), " ", ",", -1)
	return cmd[1:len(cmd)-1] + ",0s"
}

type HwIdStruct struct{ Board, Group, Node, SwId float64 }
type SwIdStruct struct{ File string }

type Config struct {
	HwId map[string]HwIdStruct // map 16-byte hwID to assigned pairing info
	SwId map[string]SwIdStruct // map each swId to a filename
}

func (c *Config) LookupHwId(hwId []byte) (board, group, node uint8) {
	key := hex.EncodeToString(hwId)
	if info, ok := c.HwId[key]; ok {
		board = uint8(info.Board)
		group = uint8(info.Group)
		node = uint8(info.Node)
	}
	return
}

func (c *Config) LookupSwId(group, node uint8) uint16 {
	for _, h := range c.HwId {
		if group == uint8(h.Group) && node == uint8(h.Node) {
			return uint16(h.SwId)
		}
	}
	return 0
}

// func loadConfig() (config Config) {
// 	// TODO: this sort of dynamic decoding is still very tedious
//
// 	hkeys, err := client.Call("db-keys", "/jeeboot/hwid/")
// 	flow.Check(err)
// 	config.HwId = make(map[string]HwIdStruct)
// 	for _, k := range hkeys.([]interface{}) {
// 		v, err := client.Call("db-get", "/jeeboot/hwid/"+k.(string))
// 		flow.Check(err)
// 		var hs HwIdStruct
// 		err = json.Unmarshal([]byte(v.(string)), &hs)
// 		flow.Check(err)
// 		config.HwId[k.(string)] = hs
// 	}
//
// 	fkeys, err := client.Call("db-keys", "/jeeboot/swid/")
// 	flow.Check(err)
// 	config.SwId = make(map[string]SwIdStruct)
// 	for _, k := range fkeys.([]interface{}) {
// 		v, err := client.Call("db-get", "/jeeboot/swid/"+k.(string))
// 		flow.Check(err)
// 		var ss SwIdStruct
// 		err = json.Unmarshal([]byte(v.(string)), &ss)
// 		flow.Check(err)
// 		config.SwId[k.(string)] = ss
// 	}
//
// 	fmt.Printf("CONFIG %d hw %d fw\n", len(config.HwId), len(config.SwId))
// 	return
// }

type Firmware struct {
	name string
	crc  uint16
	data []byte
}

// func loadAllFirmware(config Config) map[uint16]Firmware {
// 	fw := make(map[uint16]Firmware)
// 	for key, name := range config.SwId {
// 		swId, err := strconv.Atoi(key)
// 		flow.Check(err)
// 		fw[uint16(swId)] = readFirmware(name.File)
// 	}
// 	return fw
// }
//
// func readFirmware(name string) Firmware {
// 	buf := readIntelHexFile(name)
// 	data := padToBinaryMultiple(buf, 64)
// 	fmt.Printf("data %d -> %d bytes\n", buf.Len(), len(data))
//
// 	return Firmware{name, calculateCrc(data), data}
// }

type PairingRequest struct {
	Variant uint8     // variant of remote node, 1..250 freely available
	Board   uint8     // type of remote node, 100..250 freely available
	Group   uint8     // current network group, 1..250 or 0 if unpaired
	NodeId  uint8     // current node ID, 1..30 or 0 if unpaired
	Check   uint16    // crc checksum over the current shared key
	HwId    [16]uint8 // unique hardware ID or 0's if not available
}

type PairingAssign struct {
	Variant uint8     // variant of remote node, 1..250 freely available
	Board   uint8     // type of remote node, 100..250 freely available
	HwId    [16]uint8 // freshly assigned hardware ID for boards which need it
}

type PairingReply struct {
	Variant uint8     // variant of remote node, 1..250 freely available
	Board   uint8     // type of remote node, 100..250 freely available
	Group   uint8     // assigned network group, 1..250
	NodeId  uint8     // assigned node ID, 1..30
	ShKey   [16]uint8 // shared key or 0's if not used
}

type UpgradeRequest struct {
	Variant uint8  // variant of remote node, 1..250 freely available
	Board   uint8  // type of remote node, 100..250 freely available
	SwId    uint16 // current software ID 0 if unknown
	SwSize  uint16 // current software download size, in units of 16 bytes
	SwCheck uint16 // current crc checksum over entire download
}

type UpgradeReply struct {
	Variant uint8  // variant of remote node, 1..250 freely available
	Board   uint8  // type of remote node, 100..250 freely available
	SwId    uint16 // assigned software ID
	SwSize  uint16 // software download size, in units of 16 bytes
	SwCheck uint16 // crc checksum over entire download
}

type DownloadRequest struct {
	SwId    uint16 // current software ID
	SwIndex uint16 // current download index, as multiple of payload size
}

type DownloadReply struct {
	SwIdXor uint16    // current software ID xor current download index
	Data    [64]uint8 // download payload
}

func (w *JeeBoot) respondToRequest(req []byte) interface{} {
	// fmt.Printf("%s %X %d\n", w.dev, req, len(req))
	switch len(req) - 1 {

	case 22:
		var preq PairingRequest
		hdr := unpackReq(req, &preq)
		// if HwId is all zeroes, we need to issue a new random value
		if preq.HwId == [16]byte{} {
			reply := PairingAssign{Board: preq.Board}
			copy(reply.HwId[:], newRandomId())
			fmt.Printf("assigning fresh hardware ID %x for board %d hdr %08b\n",
				reply.HwId, preq.Board, hdr)
			return reply
		}
		board, group, node := w.config.LookupHwId(preq.HwId[:])
		if board == preq.Board && group != 0 && node != 0 {
			fmt.Printf("pair %x board %d hdr %08b\n", preq.HwId, board, hdr)
			reply := PairingReply{Board: board, Group: group, NodeId: node}
			return reply
		}
		fmt.Printf("pair %x board %d - no entry\n", preq.HwId, board)

	case 8:
		var ureq UpgradeRequest
		hdr := unpackReq(req, &ureq)
		group, node := uint8(212), hdr&0x1F // FIXME hard-coded for now
		// UpgradeRequest can be used as reply as well, it has the same fields
		reply := &ureq
		reply.SwId = w.config.LookupSwId(group, node)
		fw := w.fw[reply.SwId]
		reply.SwSize = uint16(len(fw.data) >> 4)
		reply.SwCheck = fw.crc
		fmt.Printf("upgrade %v hdr %08b\n", reply, hdr)
		return reply

	case 4:
		var dreq DownloadRequest
		hdr := unpackReq(req, &dreq)
		fw := w.fw[dreq.SwId]
		offset := 64 * dreq.SwIndex // FIXME hard-coded
		reply := DownloadReply{SwIdXor: dreq.SwId ^ dreq.SwIndex}
		fmt.Println("len", len(fw.data), "offset", offset, offset+64)
		for i, v := range fw.data[offset : offset+64] {
			reply.Data[i] = v ^ uint8(211*i)
		}
		fmt.Printf("download hdr %08b\n", hdr)
		return reply

	default:
		fmt.Printf("bad req? %d b = %d\n", len(req), req)
	}

	return nil
}

func unpackReq(data []byte, req interface{}) (h uint8) {
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &h)
	flow.Check(err)
	err = binary.Read(reader, binary.LittleEndian, req)
	flow.Check(err)
	fmt.Printf("%08b %X\n", h, req)
	return
}

func newRandomId() []byte {
	// use the uuid package (overkill?) to come up with 16 random bytes
	r, _ := hex.DecodeString(strings.Replace(uuid.New(), "-", "", -1))
	return r
}
