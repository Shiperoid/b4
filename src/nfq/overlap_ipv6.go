package nfq

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// sendOverlapFragmentsV6 - IPv6 version: exploits TCP segment overlap behavior
func (w *Worker) sendOverlapFragmentsV6(cfg *config.SetConfig, packet []byte, dst net.IP) {
	const ipv6HdrLen = 40

	if len(packet) < ipv6HdrLen+20 {
		_ = w.sock.SendIPv6(packet, dst)
		return
	}

	tcpHdrLen := int((packet[ipv6HdrLen+12] >> 4) * 4)
	payloadStart := ipv6HdrLen + tcpHdrLen
	payloadLen := len(packet) - payloadStart

	if payloadLen < 20 {
		_ = w.sock.SendIPv6(packet, dst)
		return
	}

	payload := packet[payloadStart:]
	seq0 := binary.BigEndian.Uint32(packet[ipv6HdrLen+4 : ipv6HdrLen+8])

	sniStart, sniEnd, ok := locateSNI(payload)
	if !ok || sniEnd <= sniStart {
		w.sendTCPSegmentsv6(cfg, packet, dst)
		return
	}

	// Segment 1: From start to beyond SNI, with FAKE SNI in overlap region
	seg1End := sniEnd + 2
	if seg1End > payloadLen {
		seg1End = payloadLen
	}

	seg1Len := payloadStart + seg1End
	seg1 := make([]byte, seg1Len)
	copy(seg1[:payloadStart], packet[:payloadStart])
	copy(seg1[payloadStart:], payload[:seg1End])

	// Random garbage instead of predictable pattern
	sniLen := sniEnd - sniStart
	garbageSNI := make([]byte, sniLen)
	rand.Read(garbageSNI)

	copy(seg1[payloadStart+sniStart:payloadStart+sniEnd], garbageSNI)
	binary.BigEndian.PutUint16(seg1[4:6], uint16(seg1Len-ipv6HdrLen))
	seg1[ipv6HdrLen+13] &^= 0x08 // Clear PSH
	sock.FixTCPChecksumV6(seg1)

	// Segment 2: Starts BEFORE seg1 ends (overlap), contains real SNI
	overlapStart := sniStart - 8
	if overlapStart < 0 {
		overlapStart = 0
	}

	seg2Len := payloadStart + (payloadLen - overlapStart)
	seg2 := make([]byte, seg2Len)
	copy(seg2[:payloadStart], packet[:payloadStart])
	copy(seg2[payloadStart:], payload[overlapStart:])

	binary.BigEndian.PutUint32(seg2[ipv6HdrLen+4:ipv6HdrLen+8], seq0+uint32(overlapStart))
	binary.BigEndian.PutUint16(seg2[4:6], uint16(seg2Len-ipv6HdrLen))
	sock.FixTCPChecksumV6(seg2)

	delay := cfg.TCP.Seg2Delay

	_ = w.sock.SendIPv6(seg1, dst)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	_ = w.sock.SendIPv6(seg2, dst)
}
