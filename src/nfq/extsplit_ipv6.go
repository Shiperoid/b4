package nfq

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// sendExtSplitFragmentsV6 - IPv6 version: splits before SNI extension
func (w *Worker) sendExtSplitFragmentsV6(cfg *config.SetConfig, packet []byte, dst net.IP) {
	const ipv6HdrLen = 40

	if len(packet) < ipv6HdrLen+20 {
		_ = w.sock.SendIPv6(packet, dst)
		return
	}

	tcpHdrLen := int((packet[ipv6HdrLen+12] >> 4) * 4)
	payloadStart := ipv6HdrLen + tcpHdrLen
	payloadLen := len(packet) - payloadStart

	if payloadLen < 50 {
		_ = w.sock.SendIPv6(packet, dst)
		return
	}

	payload := packet[payloadStart:]
	splitPos := findPreSNIExtensionPoint(payload)

	if splitPos <= 5 || splitPos >= payloadLen-10 {
		w.sendTCPSegmentsv6(cfg, packet, dst)
		return
	}

	seq0 := binary.BigEndian.Uint32(packet[ipv6HdrLen+4 : ipv6HdrLen+8])

	// Segment 1: everything before SNI extension
	seg1Len := payloadStart + splitPos
	seg1 := make([]byte, seg1Len)
	copy(seg1[:payloadStart], packet[:payloadStart])
	copy(seg1[payloadStart:], payload[:splitPos])

	binary.BigEndian.PutUint16(seg1[4:6], uint16(seg1Len-ipv6HdrLen))
	seg1[ipv6HdrLen+13] &^= 0x08 // Clear PSH
	sock.FixTCPChecksumV6(seg1)

	// Segment 2: SNI extension onwards
	seg2Len := payloadStart + (payloadLen - splitPos)
	seg2 := make([]byte, seg2Len)
	copy(seg2[:payloadStart], packet[:payloadStart])
	copy(seg2[payloadStart:], payload[splitPos:])

	binary.BigEndian.PutUint32(seg2[ipv6HdrLen+4:ipv6HdrLen+8], seq0+uint32(splitPos))
	binary.BigEndian.PutUint16(seg2[4:6], uint16(seg2Len-ipv6HdrLen))
	sock.FixTCPChecksumV6(seg2)

	delay := cfg.TCP.Seg2Delay

	if cfg.Fragmentation.ReverseOrder {
		_ = w.sock.SendIPv6(seg2, dst)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		_ = w.sock.SendIPv6(seg1, dst)
	} else {
		_ = w.sock.SendIPv6(seg1, dst)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		_ = w.sock.SendIPv6(seg2, dst)
	}
}
