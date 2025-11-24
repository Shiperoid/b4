package nfq

import (
	"crypto/rand"
	"encoding/binary"
	"net"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// GREASE values (RFC 8701)
var greaseValues = []uint16{
	0x0a0a, 0x1a1a, 0x2a2a, 0x3a3a,
	0x4a4a, 0x5a5a, 0x6a6a, 0x7a7a,
	0x8a8a, 0x9a9a, 0xaaaa, 0xbaba,
	0xcaca, 0xdada, 0xeaea, 0xfafa,
}

// TLS Extension types
const (
	extServerName           = 0x0000
	extMaxFragmentLength    = 0x0001
	extStatusRequest        = 0x0005
	extSupportedGroups      = 0x000a
	extSignatureAlgorithms  = 0x000d
	extUseSRTP              = 0x000e
	extHeartbeat            = 0x000f
	extALPN                 = 0x0010
	extSCT                  = 0x0012
	extPadding              = 0x0015
	extExtendedMasterSecret = 0x0017
	extSessionTicket        = 0x0023
	extSupportedVersions    = 0x002b
	extPSKExchangeModes     = 0x002d
	extKeyShare             = 0x0033
)

// MutateClientHello applies mutations to TLS ClientHello
func (w *Worker) MutateClientHello(cfg *config.SetConfig, packet []byte, dst net.IP) []byte {
	if cfg.Faking.SNIMutation.Mode == "off" {
		return packet
	}

	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen

	if len(packet) <= payloadStart+5 {
		return packet
	}

	// Check if this is TLS ClientHello
	payload := packet[payloadStart:]
	if payload[0] != 0x16 || payload[1] != 0x03 {
		return packet
	}

	// Parse TLS record
	recordLen := int(binary.BigEndian.Uint16(payload[3:5]))
	if len(payload) < 5+recordLen {
		return packet
	}

	// Check for ClientHello (type 0x01)
	if payload[5] != 0x01 {
		return packet
	}

	switch cfg.Faking.SNIMutation.Mode {
	case "duplicate":
		return w.duplicateSNI(packet, cfg)
	case "grease":
		return w.addGREASE(packet, cfg)
	case "padding":
		return w.addPadding(packet, cfg)
	case "reorder":
		return w.reorderExtensions(packet, cfg)
	case "full":
		return w.fullMutation(packet, cfg)
	default:
		return w.fullMutation(packet, cfg)
	}
}

// duplicateSNI adds multiple SNI extensions
func (w *Worker) duplicateSNI(packet []byte, cfg *config.SetConfig) []byte {
	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen

	// Find extensions offset
	extOffset := w.findExtensionsOffset(packet[payloadStart:])
	if extOffset < 0 {
		return packet
	}

	// Build fake SNI extensions
	fakeSNIs := make([]byte, 0, 1024)

	for _, sni := range cfg.Faking.SNIMutation.FakeSNIs {
		if sni == "" {
			continue
		}

		sniExt := make([]byte, 9+len(sni))
		binary.BigEndian.PutUint16(sniExt[0:2], extServerName)      // Extension type
		binary.BigEndian.PutUint16(sniExt[2:4], uint16(5+len(sni))) // Extension length
		binary.BigEndian.PutUint16(sniExt[4:6], uint16(3+len(sni))) // Server name list length
		sniExt[6] = 0                                               // Name type: host_name
		binary.BigEndian.PutUint16(sniExt[7:9], uint16(len(sni)))   // Name length
		copy(sniExt[9:], sni)

		fakeSNIs = append(fakeSNIs, sniExt...)
	}

	// Insert fake SNIs into packet
	return w.insertExtensions(packet, fakeSNIs)
}

// addGREASE adds GREASE extensions
func (w *Worker) addGREASE(packet []byte, cfg *config.SetConfig) []byte {
	grease := make([]byte, 0, cfg.Faking.SNIMutation.GreaseCount*8)

	for i := 0; i < cfg.Faking.SNIMutation.GreaseCount; i++ {
		// Pick random GREASE value
		greaseVal := greaseValues[i%len(greaseValues)]

		ext := make([]byte, 8)
		binary.BigEndian.PutUint16(ext[0:2], greaseVal) // GREASE extension type
		binary.BigEndian.PutUint16(ext[2:4], 4)         // Length

		// Random GREASE data
		rand.Read(ext[4:8])

		grease = append(grease, ext...)
	}

	return w.insertExtensions(packet, grease)
}

// addPadding adds padding extension
func (w *Worker) addPadding(packet []byte, cfg *config.SetConfig) []byte {
	paddingSize := cfg.Faking.SNIMutation.PaddingSize
	if paddingSize < 16 {
		paddingSize = 16
	}
	if paddingSize > 4096 {
		paddingSize = 4096
	}

	padding := make([]byte, 4+paddingSize)
	binary.BigEndian.PutUint16(padding[0:2], extPadding)          // Extension type
	binary.BigEndian.PutUint16(padding[2:4], uint16(paddingSize)) // Length
	// Leave padding data as zeros (standard practice)

	return w.insertExtensions(packet, padding)
}

// reorderExtensions randomly reorders TLS extensions
func (w *Worker) reorderExtensions(packet []byte, cfg *config.SetConfig) []byte {
	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen

	extOffset := w.findExtensionsOffset(packet[payloadStart:])
	if extOffset < 0 {
		return packet
	}

	// Parse existing extensions
	extensions := w.parseExtensions(packet[payloadStart+extOffset:])
	if len(extensions) < 2 {
		return packet // Nothing to reorder
	}

	// Shuffle extensions (keep SNI first for compatibility)
	var sniExt []byte
	var otherExts [][]byte

	for _, ext := range extensions {
		if len(ext) >= 2 && binary.BigEndian.Uint16(ext[0:2]) == extServerName {
			sniExt = ext
		} else {
			otherExts = append(otherExts, ext)
		}
	}

	// Random shuffle other extensions
	for i := len(otherExts) - 1; i > 0; i-- {
		j := int(randomUint32() % uint32(i+1))
		otherExts[i], otherExts[j] = otherExts[j], otherExts[i]
	}

	// Rebuild extensions
	newExts := make([]byte, 0, 4096)
	if sniExt != nil {
		newExts = append(newExts, sniExt...)
	}
	for _, ext := range otherExts {
		newExts = append(newExts, ext...)
	}

	return w.replaceExtensions(packet, newExts)
}

// fullMutation applies all mutations
func (w *Worker) fullMutation(packet []byte, cfg *config.SetConfig) []byte {
	mutated := packet

	// 1. Add duplicate SNIs
	mutated = w.duplicateSNI(mutated, cfg)

	// 2. Add GREASE
	mutated = w.addGREASE(mutated, cfg)

	// 3. Add fake ALPN with many protocols
	mutated = w.addFakeALPN(mutated)

	// 4. Add unknown extensions
	mutated = w.addUnknownExtensions(mutated, cfg.Faking.SNIMutation.FakeExtCount)

	// 5. Reorder extensions
	mutated = w.reorderExtensions(mutated, cfg)

	// 6. Add padding last to fill MTU
	mutated = w.addPadding(mutated, cfg)

	return mutated
}

// Helper: Add fake ALPN extension with many protocols
func (w *Worker) addFakeALPN(packet []byte) []byte {
	protocols := []string{
		"http/1.0", "http/1.1", "h2", "h3",
		"spdy/3", "spdy/3.1",
		"quic", "hq", "doq",
		"xmpp", "mqtt", "amqp",
		"grpc", "websocket",
	}

	alpnData := make([]byte, 0, 256)
	for _, proto := range protocols {
		alpnData = append(alpnData, byte(len(proto)))
		alpnData = append(alpnData, proto...)
	}

	alpn := make([]byte, 6+len(alpnData))
	binary.BigEndian.PutUint16(alpn[0:2], extALPN)                 // Extension type
	binary.BigEndian.PutUint16(alpn[2:4], uint16(2+len(alpnData))) // Extension length
	binary.BigEndian.PutUint16(alpn[4:6], uint16(len(alpnData)))   // ALPN list length
	copy(alpn[6:], alpnData)

	return w.insertExtensions(packet, alpn)
}

// Helper: Add unknown/reserved extension types
func (w *Worker) addUnknownExtensions(packet []byte, count int) []byte {
	unknown := make([]byte, 0, count*8)

	// Use reserved/unassigned extension types
	extTypes := []uint16{0x00ff, 0x1234, 0x5678, 0x9abc, 0xfe00, 0xffff}

	for i := 0; i < count && i < len(extTypes); i++ {
		ext := make([]byte, 8)
		binary.BigEndian.PutUint16(ext[0:2], extTypes[i])
		binary.BigEndian.PutUint16(ext[2:4], 4) // Length
		rand.Read(ext[4:8])                     // Random data

		unknown = append(unknown, ext...)
	}

	return w.insertExtensions(packet, unknown)
}

// Helper: Find extensions offset in ClientHello
func (w *Worker) findExtensionsOffset(payload []byte) int {
	if len(payload) < 43 {
		return -1
	}

	// Skip TLS header (5 bytes)
	// Skip Handshake header (4 bytes)
	// Skip Version (2 bytes)
	// Skip Random (32 bytes)
	pos := 43

	// Session ID
	if pos >= len(payload) {
		return -1
	}
	sidLen := int(payload[pos])
	pos += 1 + sidLen

	// Cipher suites
	if pos+2 > len(payload) {
		return -1
	}
	csLen := int(binary.BigEndian.Uint16(payload[pos : pos+2]))
	pos += 2 + csLen

	// Compression methods
	if pos >= len(payload) {
		return -1
	}
	compLen := int(payload[pos])
	pos += 1 + compLen

	// Extensions start here
	if pos+2 > len(payload) {
		return -1
	}

	return pos
}

// Helper: Parse extensions from buffer
func (w *Worker) parseExtensions(data []byte) [][]byte {
	if len(data) < 2 {
		return nil
	}

	extLen := int(binary.BigEndian.Uint16(data[0:2]))
	if len(data) < 2+extLen {
		return nil
	}

	extensions := [][]byte{}
	pos := 2

	for pos+4 <= 2+extLen {
		// Extension type (2 bytes)
		// Extension length (2 bytes)
		el := int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		if pos+4+el > 2+extLen {
			break
		}

		ext := make([]byte, 4+el)
		copy(ext, data[pos:pos+4+el])
		extensions = append(extensions, ext)

		pos += 4 + el
	}

	return extensions
}

// Helper: Insert extensions into packet
func (w *Worker) insertExtensions(packet []byte, newExts []byte) []byte {
	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen

	extOffset := w.findExtensionsOffset(packet[payloadStart:])
	if extOffset < 0 {
		return packet
	}

	// Get current extensions
	extPos := payloadStart + extOffset
	if len(packet) < extPos+2 {
		return packet
	}

	currentExtLen := int(binary.BigEndian.Uint16(packet[extPos : extPos+2]))

	// Build new packet with inserted extensions
	newPacket := make([]byte, len(packet)+len(newExts))

	// Copy everything before extensions
	copy(newPacket, packet[:extPos])

	// Write new extensions length
	newExtLen := currentExtLen + len(newExts)
	binary.BigEndian.PutUint16(newPacket[extPos:extPos+2], uint16(newExtLen))

	// Copy original extensions
	copy(newPacket[extPos+2:], packet[extPos+2:extPos+2+currentExtLen])

	// Add new extensions
	copy(newPacket[extPos+2+currentExtLen:], newExts)

	// Copy everything after extensions
	copy(newPacket[extPos+2+currentExtLen+len(newExts):], packet[extPos+2+currentExtLen:])

	// Update lengths
	w.updatePacketLengths(newPacket)

	return newPacket
}

// Helper: Replace all extensions
func (w *Worker) replaceExtensions(packet []byte, newExts []byte) []byte {
	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen

	extOffset := w.findExtensionsOffset(packet[payloadStart:])
	if extOffset < 0 {
		return packet
	}

	extPos := payloadStart + extOffset
	currentExtLen := int(binary.BigEndian.Uint16(packet[extPos : extPos+2]))

	// Build new packet
	sizeDiff := len(newExts) - currentExtLen
	newPacket := make([]byte, len(packet)+sizeDiff)

	// Copy everything before extensions
	copy(newPacket, packet[:extPos])

	// Write new extensions length
	binary.BigEndian.PutUint16(newPacket[extPos:extPos+2], uint16(len(newExts)))

	// Write new extensions
	copy(newPacket[extPos+2:], newExts)

	// Copy everything after old extensions
	copy(newPacket[extPos+2+len(newExts):], packet[extPos+2+currentExtLen:])

	// Update lengths
	w.updatePacketLengths(newPacket)

	return newPacket
}

// Helper: Update all packet lengths after mutation
func (w *Worker) updatePacketLengths(packet []byte) {
	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen
	payloadLen := len(packet) - payloadStart

	// Update IP total length
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(packet)))

	// Update TLS record length
	if payloadLen >= 5 {
		binary.BigEndian.PutUint16(packet[payloadStart+3:payloadStart+5], uint16(payloadLen-5))
	}

	// Update ClientHello length
	if payloadLen >= 9 {
		helloLen := payloadLen - 9
		packet[payloadStart+6] = byte(helloLen >> 16)
		packet[payloadStart+7] = byte(helloLen >> 8)
		packet[payloadStart+8] = byte(helloLen)
	}

	// Fix checksums
	sock.FixIPv4Checksum(packet[:ipHdrLen])
	sock.FixTCPChecksum(packet)
}

func randomUint32() uint32 {
	var b [4]byte
	rand.Read(b[:])
	return binary.BigEndian.Uint32(b[:])
}
