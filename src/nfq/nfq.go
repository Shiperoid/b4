package nfq

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/log"
	"github.com/daniellavrushin/b4/sni"
	"github.com/daniellavrushin/b4/sock"
	"github.com/florianl/go-nfqueue"
)

func (w *Worker) Start() error {
	s, err := sock.NewSenderWithMark(int(w.cfg.Mark))
	if err != nil {
		return err
	}
	w.sock = s
	w.frag = &sock.Fragmenter{}

	c := nfqueue.Config{
		NfQueue:      w.qnum,
		MaxPacketLen: 0xffff,
		MaxQueueLen:  4096,
		Copymode:     nfqueue.NfQnlCopyPacket,
	}
	q, err := nfqueue.Open(&c)
	if err != nil {
		return err
	}
	w.q = q

	w.wg.Add(1)
	go w.gc()

	go func() {
		pid := os.Getpid()
		log.Infof("NFQ bound pid=%d queue=%d", pid, w.qnum)
		_ = q.RegisterWithErrorFunc(w.ctx, func(a nfqueue.Attribute) int {
			if a.PacketID == nil || a.Payload == nil || len(*a.Payload) == 0 {
				return 0
			}
			id := *a.PacketID
			raw := *a.Payload

			v := raw[0] >> 4
			if v != 4 && v != 6 {
				_ = q.SetVerdict(id, nfqueue.NfAccept)
				return 0
			}
			var proto uint8
			var src, dst net.IP
			var ihl int
			if v == 4 {
				if len(raw) < 20 {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				ihl = int(raw[0]&0x0f) * 4
				if len(raw) < ihl {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				proto = raw[9]
				src = net.IP(raw[12:16])
				dst = net.IP(raw[16:20])
			} else {
				if len(raw) < 40 {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				ihl = 40
				proto = raw[6]
				src = net.IP(raw[8:24])
				dst = net.IP(raw[24:40])
			}

			if proto == 6 && len(raw) >= ihl+20 {
				tcp := raw[ihl:]
				if len(tcp) < 20 {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				datOff := int((tcp[12]>>4)&0x0f) * 4
				if len(tcp) < datOff {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				payload := tcp[datOff:]
				sport := binary.BigEndian.Uint16(tcp[0:2])
				dport := binary.BigEndian.Uint16(tcp[2:4])
				if dport == 443 && len(payload) > 0 {
					k := fmt.Sprintf("%s:%d>%s:%d", src.String(), sport, dst.String(), dport)
					host, ok := w.feed(k, payload)
					if ok && w.matcher.Match(host) {
						log.Infof("TCP: %s %s:%d -> %s:%d", host, src.String(), sport, dst.String(), dport)
						go w.dropAndInjectTCP(raw, dst)
						_ = q.SetVerdict(id, nfqueue.NfDrop)
						return 0
					}
				}
			}

			if proto == 17 && len(raw) >= ihl+8 {
				udp := raw[ihl:]
				if len(udp) >= 8 {
					payload := udp[8:]
					sport := binary.BigEndian.Uint16(udp[0:2])
					dport := binary.BigEndian.Uint16(udp[2:4])
					if dport == 443 {
						if host, ok := sni.ParseQUICClientHelloSNI(payload); ok && w.matcher.Match(host) {
							log.Infof("UDP: %s %s:%d -> %s:%d", host, src.String(), sport, dst.String(), dport)
							go w.dropAndInjectQUIC(raw, dst)
							_ = q.SetVerdict(id, nfqueue.NfDrop)
							return 0
						}
					}
				}
			}

			_ = q.SetVerdict(id, nfqueue.NfAccept)
			return 0
		}, func(err error) int {
			log.Errorf("nfq: %v", err)
			return 0
		})
	}()

	return nil
}

func (w *Worker) dropAndInjectQUIC(raw []byte, dst net.IP) {
	fake, ok := sock.BuildFakeUDPFromOriginal(raw, 1200, 8)
	if ok {
		_ = w.sock.SendIPv4(fake, dst)
		time.Sleep(10 * time.Millisecond)
	}
	frags, ok := sock.IPv4FragmentUDP(raw, 24)
	if !ok {
		return
	}
	for i, f := range frags {
		_ = w.sock.SendIPv4(f, dst)
		if i == 0 {
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (w *Worker) dropAndInjectTCP(raw []byte, dst net.IP) {
	if len(raw) < 40 || raw[0]>>4 != 4 {
		_ = w.sock.SendIPv4(raw, dst)
		return
	}

	ipHdrLen := int((raw[0] & 0x0F) * 4)
	tcpHdrLen := int((raw[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen

	if len(raw) <= payloadStart {
		_ = w.sock.SendIPv4(raw, dst)
		return
	}

	// Send fake SNI packets BEFORE the real fragments
	if w.cfg.FakeSNI {
		for i := 0; i < w.cfg.FakeSNISeqLength; i++ {
			fake := w.buildFakeSNI(raw)
			if fake != nil {
				_ = w.sock.SendIPv4(fake, dst)
			}
		}
	}

	// Find split position - default to position 1 in payload
	splitPos := 1
	if w.cfg.FragSNIPosition > 0 {
		splitPos = w.cfg.FragSNIPosition
	}

	// Fragment based on strategy
	switch w.cfg.FragmentStrategy {
	case "tcp":
		w.sendTCPFragments(raw, payloadStart+splitPos, dst)
	case "ip":
		w.sendIPFragments(raw, payloadStart+splitPos, dst)
	default:
		_ = w.sock.SendIPv4(raw, dst)
	}
}

func (w *Worker) feed(key string, chunk []byte) (string, bool) {
	w.mu.Lock()
	st := w.flows[key]
	if st == nil {
		st = &flowState{buf: nil, last: time.Now()}
		w.flows[key] = st
	}
	if len(st.buf) < w.limit {
		need := w.limit - len(st.buf)
		if len(chunk) < need {
			st.buf = append(st.buf, chunk...)
		} else {
			st.buf = append(st.buf, chunk[:need]...)
		}
	}
	st.last = time.Now()
	buf := append([]byte(nil), st.buf...)
	w.mu.Unlock()
	host, ok := sni.ParseTLSClientHelloSNI(buf)
	if ok && host != "" {
		w.mu.Lock()
		delete(w.flows, key)
		w.mu.Unlock()
		return host, true
	}
	return "", false
}

func (w *Worker) buildFakeSNI(original []byte) []byte {
	ipHdrLen := int((original[0] & 0x0F) * 4)
	tcpHdrLen := int((original[ipHdrLen+12] >> 4) * 4)

	// Use the default fake SNI payload from youtubeUnblock
	fakePayload := sock.DefaultFakeSNI
	if w.cfg.FakeSNIType == config.FakePayloadRandom {
		// Generate random payload
		fakePayload = make([]byte, 1200)
		for i := range fakePayload {
			fakePayload[i] = byte(rand.Intn(256))
		}
	}

	fake := make([]byte, ipHdrLen+tcpHdrLen+len(fakePayload))

	// Copy headers
	copy(fake, original[:ipHdrLen+tcpHdrLen])
	copy(fake[ipHdrLen+tcpHdrLen:], fakePayload)

	// Update IP header
	binary.BigEndian.PutUint16(fake[2:4], uint16(len(fake))) // Total length

	// Apply faking strategy
	switch w.cfg.FakeStrategy {
	case "ttl":
		fake[8] = w.cfg.FakeTTL
	case "pastseq":
		seq := binary.BigEndian.Uint32(fake[ipHdrLen+4 : ipHdrLen+8])
		binary.BigEndian.PutUint32(fake[ipHdrLen+4:ipHdrLen+8],
			seq-uint32(len(fakePayload)))
	case "randseq":
		seq := binary.BigEndian.Uint32(fake[ipHdrLen+4 : ipHdrLen+8])
		binary.BigEndian.PutUint32(fake[ipHdrLen+4:ipHdrLen+8],
			seq-uint32(w.cfg.FakeSeqOffset)+uint32(len(fakePayload)))
	case "tcp_check":
		// Will break checksum later
	case "md5sum":
		// Add TCP MD5 option (requires extending TCP header)
		w.addTCPMD5Option(fake, ipHdrLen)
	}

	// Fix checksums
	sock.FixIPv4Checksum(fake[:ipHdrLen])
	sock.FixTCPChecksum(fake)

	// Break checksum if needed
	if w.cfg.FakeStrategy == "tcp_check" {
		fake[ipHdrLen+16] += 1
	}

	return fake
}

func (w *Worker) addTCPMD5Option(packet []byte, ipHdrLen int) []byte {
	tcpOffset := ipHdrLen
	tcpHdrLen := int((packet[tcpOffset+12] >> 4) * 4)

	// TCP MD5 option requires 20 bytes (kind=19, len=18, sig=16)
	const MD5_OPT_LEN = 20

	// Check if we need to extend the TCP header
	optLen := tcpHdrLen - 20 // Current options length
	needed := MD5_OPT_LEN - optLen

	if needed > 0 {
		// Need to extend the packet
		newPacket := make([]byte, len(packet)+needed)

		// Copy IP header and TCP header
		copy(newPacket, packet[:ipHdrLen+tcpHdrLen])

		// Copy payload after making room
		copy(newPacket[ipHdrLen+tcpHdrLen+needed:], packet[ipHdrLen+tcpHdrLen:])

		// Update to new packet
		packet = newPacket
		tcpHdrLen += needed

		// Update TCP data offset
		packet[tcpOffset+12] = byte((tcpHdrLen/4)<<4) | (packet[tcpOffset+12] & 0x0F)

		// Update IP total length
		totalLen := binary.BigEndian.Uint16(packet[2:4]) + uint16(needed)
		binary.BigEndian.PutUint16(packet[2:4], totalLen)
	}

	// Add MD5 option at the end of TCP options
	optStart := ipHdrLen + 20 // Start of TCP options

	// MD5 signature option (RFC 2385)
	packet[optStart] = 19   // Kind
	packet[optStart+1] = 18 // Length

	// Zero out signature bytes
	for i := 0; i < 16; i++ {
		packet[optStart+2+i] = 0
	}

	// Fill remaining with NOPs
	for i := optStart + 18; i < ipHdrLen+tcpHdrLen; i++ {
		packet[i] = 0x01 // NOP
	}

	return packet
}

func (w *Worker) sendTCPFragments(packet []byte, splitPos int, dst net.IP) {
	if splitPos <= 0 || splitPos >= len(packet) {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	hdrTotal := ipHdrLen + tcpHdrLen

	if splitPos <= hdrTotal {
		splitPos = hdrTotal + 1
	}

	// Create two segments
	seg1Len := splitPos
	seg1 := make([]byte, seg1Len)
	copy(seg1, packet[:seg1Len])

	seg2Len := len(packet) - splitPos + hdrTotal
	seg2 := make([]byte, seg2Len)

	// Copy headers to segment 2
	copy(seg2, packet[:hdrTotal])
	// Copy remaining payload
	copy(seg2[hdrTotal:], packet[splitPos:])

	// Fix segment 1
	binary.BigEndian.PutUint16(seg1[2:4], uint16(seg1Len))
	sock.FixIPv4Checksum(seg1[:ipHdrLen])
	sock.FixTCPChecksum(seg1)

	// Fix segment 2
	binary.BigEndian.PutUint16(seg2[2:4], uint16(seg2Len))
	// Adjust sequence number
	seq := binary.BigEndian.Uint32(seg2[ipHdrLen+4 : ipHdrLen+8])
	binary.BigEndian.PutUint32(seg2[ipHdrLen+4:ipHdrLen+8],
		seq+uint32(splitPos-hdrTotal))
	sock.FixIPv4Checksum(seg2[:ipHdrLen])
	sock.FixTCPChecksum(seg2)

	// Send in order based on config
	if w.cfg.FragSNIReverse {
		_ = w.sock.SendIPv4(seg2, dst)
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(seg1, dst)
	} else {
		_ = w.sock.SendIPv4(seg1, dst)
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(seg2, dst)
	}
}

func (w *Worker) sendIPFragments(packet []byte, splitPos int, dst net.IP) {
	if splitPos <= 0 || splitPos >= len(packet) {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	ipHdrLen := int((packet[0] & 0x0F) * 4)

	// Align to 8-byte boundary for IP fragmentation
	splitPos = (splitPos + 7) &^ 7
	if splitPos >= len(packet) {
		splitPos = len(packet) - 8
	}

	// Fragment 1
	frag1 := make([]byte, splitPos)
	copy(frag1, packet[:splitPos])

	// Set MF flag
	frag1[6] |= 0x20
	binary.BigEndian.PutUint16(frag1[2:4], uint16(splitPos))
	sock.FixIPv4Checksum(frag1[:ipHdrLen])

	// Fragment 2
	frag2Len := ipHdrLen + len(packet) - splitPos
	frag2 := make([]byte, frag2Len)
	copy(frag2, packet[:ipHdrLen])
	copy(frag2[ipHdrLen:], packet[splitPos:])

	// Set fragment offset
	fragOff := uint16(splitPos-ipHdrLen) / 8
	binary.BigEndian.PutUint16(frag2[6:8], fragOff)
	binary.BigEndian.PutUint16(frag2[2:4], uint16(frag2Len))
	sock.FixIPv4Checksum(frag2[:ipHdrLen])

	// Send fragments
	if w.cfg.FragSNIReverse {
		_ = w.sock.SendIPv4(frag2, dst)
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(frag1, dst)
	} else {
		_ = w.sock.SendIPv4(frag1, dst)
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(frag2, dst)
	}
}

func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	if w.q != nil {
		_ = w.q.Close()
	}
	if w.sock != nil {
		w.sock.Close()
	}
}

func (w *Worker) gc() {
	defer w.wg.Done()
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case now := <-t.C:
			w.mu.Lock()
			for k, st := range w.flows {
				if now.Sub(st.last) > w.ttl {
					delete(w.flows, k)
				}
			}
			w.mu.Unlock()
		}
	}
}
