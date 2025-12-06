package nfq

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// sendDisorderFragmentsV6 - IPv6 version: splits and sends in random order
func (w *Worker) sendDisorderFragmentsV6(cfg *config.SetConfig, packet []byte, dst net.IP) {
	const ipv6HdrLen = 40

	if len(packet) < ipv6HdrLen+20 {
		_ = w.sock.SendIPv6(packet, dst)
		return
	}

	tcpHdrLen := int((packet[ipv6HdrLen+12] >> 4) * 4)
	payloadStart := ipv6HdrLen + tcpHdrLen
	payloadLen := len(packet) - payloadStart

	if payloadLen < 10 {
		_ = w.sock.SendIPv6(packet, dst)
		return
	}

	payload := packet[payloadStart:]
	seq0 := binary.BigEndian.Uint32(packet[ipv6HdrLen+4 : ipv6HdrLen+8])

	var splits []int
	if sniStart, sniEnd, ok := locateSNI(payload); ok && sniEnd > sniStart {
		sniLen := sniEnd - sniStart
		splits = append(splits, sniStart)
		if sniLen > 6 {
			splits = append(splits, sniStart+sniLen/2)
		}
		splits = append(splits, sniEnd)
	} else {
		splits = []int{1, payloadLen / 2, payloadLen * 3 / 4}
	}

	validSplits := []int{0}
	for _, s := range splits {
		if s > 0 && s < payloadLen {
			validSplits = append(validSplits, s)
		}
	}
	validSplits = append(validSplits, payloadLen)

	type segment struct {
		data   []byte
		seqOff uint32
	}

	segments := make([]segment, 0, len(validSplits)-1)
	for i := 0; i < len(validSplits)-1; i++ {
		start := validSplits[i]
		end := validSplits[i+1]

		segLen := payloadStart + (end - start)
		seg := make([]byte, segLen)
		copy(seg[:payloadStart], packet[:payloadStart])
		copy(seg[payloadStart:], payload[start:end])

		binary.BigEndian.PutUint32(seg[ipv6HdrLen+4:ipv6HdrLen+8], seq0+uint32(start))
		binary.BigEndian.PutUint16(seg[4:6], uint16(segLen-ipv6HdrLen)) // IPv6 payload length

		if i < len(validSplits)-2 {
			seg[ipv6HdrLen+13] &^= 0x08 // Clear PSH
		}

		sock.FixTCPChecksumV6(seg)
		segments = append(segments, segment{data: seg, seqOff: uint32(start)})
	}

	// Thread-safe shuffle
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	if len(segments) > 2 {
		for i := len(segments) - 1; i > 0; i-- {
			j := r.Intn(i + 1)
			segments[i], segments[j] = segments[j], segments[i]
		}
	} else if len(segments) == 2 {
		segments[0], segments[1] = segments[1], segments[0]
	}

	seg2d := cfg.TCP.Seg2Delay
	for i, seg := range segments {
		_ = w.sock.SendIPv6(seg.data, dst)
		if i < len(segments)-1 {
			if seg2d > 0 {
				jitter := r.Intn(seg2d/2 + 1)
				time.Sleep(time.Duration(seg2d+jitter) * time.Millisecond)
			} else {
				time.Sleep(time.Duration(1000+r.Intn(2000)) * time.Microsecond)
			}
		}
	}
}
