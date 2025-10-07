package nfq

import (
	"context"
	"encoding/binary"
	"net"
	"sync"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/log"
	"github.com/daniellavrushin/b4/sni"
	"github.com/florianl/go-nfqueue"
	"golang.org/x/sys/unix"
)

type Worker struct {
	cfg    *config.Config
	ctx    context.Context
	cancel context.CancelFunc
	qs     []*nfqueue.Nfqueue
	wg     sync.WaitGroup
}

func NewWorker(cfg *config.Config) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{cfg: cfg, ctx: ctx, cancel: cancel}
}

func (w *Worker) startOne(family uint8) error {
	flags := nfqueue.NfQaCfgFlagFailOpen
	if w.cfg.UseGSO {
		flags |= nfqueue.NfQaCfgFlagGSO
	}
	if w.cfg.UseConntrack {
		flags |= nfqueue.NfQaCfgFlagConntrack
	}
	c := nfqueue.Config{
		NfQueue:      uint16(w.cfg.QueueStartNum),
		MaxPacketLen: 0xffff,
		Copymode:     nfqueue.NfQnlCopyPacket,
		Flags:        uint32(flags),
		AfFamily:     family,
	}
	q, err := nfqueue.Open(&c)
	if err != nil {
		return err
	}
	w.qs = append(w.qs, q)
	w.wg.Add(1)
	go func(qx *nfqueue.Nfqueue) {
		defer w.wg.Done()
		_ = qx.RegisterWithErrorFunc(w.ctx, func(a nfqueue.Attribute) int {
			if a.PacketID == nil {
				return 0
			}
			id := *a.PacketID
			if a.Mark != nil && *a.Mark == uint32(w.cfg.Mark) {
				_ = qx.SetVerdict(id, nfqueue.NfAccept)
				return 0
			}
			if a.Payload == nil || len(*a.Payload) == 0 {
				_ = qx.SetVerdict(id, nfqueue.NfAccept)
				return 0
			}
			raw := *a.Payload
			v := raw[0] >> 4
			switch v {
			case 4:
				if len(raw) < 20 {
					_ = qx.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				ihl := int(raw[0]&0x0f) * 4
				if len(raw) < ihl {
					_ = qx.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				proto := raw[9]
				src := net.IP(raw[12:16])
				dst := net.IP(raw[16:20])
				if proto == 6 && len(raw) >= ihl+20 {
					tcp := raw[ihl:]
					datOff := int((tcp[12]>>4)&0x0f) * 4
					if len(tcp) >= datOff {
						payload := tcp[datOff:]
						dport := binary.BigEndian.Uint16(tcp[2:4])
						if dport == 443 {
							if host, ok := sni.ParseTLSClientHelloSNI(payload); ok && host != "" {
								log.Infof("NFQ TLS SNI v4: %s %s:%d -> %s:%d", host, src.String(), binary.BigEndian.Uint16(tcp[0:2]), dst.String(), dport)
							}
						}
					}
				} else if proto == 17 && len(raw) >= ihl+8 {
					udp := raw[ihl:]
					dport := binary.BigEndian.Uint16(udp[2:4])
					if dport == 443 && len(udp) >= 8 {
						payload := udp[8:]
						if host, ok := sni.ParseQUICClientHelloSNI(payload); ok && host != "" {
							log.Infof("NFQ QUIC SNI v4: %s %s:%d -> %s:%d", host, src.String(), binary.BigEndian.Uint16(udp[0:2]), dst.String(), dport)
						}
					}
				}
			case 6:
				if len(raw) < 40 {
					_ = qx.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				next := raw[6]
				src := net.IP(raw[8:24])
				dst := net.IP(raw[24:40])
				l4 := raw[40:]
				if next == 6 && len(l4) >= 20 {
					datOff := int((l4[12]>>4)&0x0f) * 4
					if len(l4) >= datOff {
						payload := l4[datOff:]
						dport := binary.BigEndian.Uint16(l4[2:4])
						if dport == 443 {
							if host, ok := sni.ParseTLSClientHelloSNI(payload); ok && host != "" {
								log.Infof("NFQ TLS SNI v6: %s [%s]:%d -> [%s]:%d", host, src.String(), binary.BigEndian.Uint16(l4[0:2]), dst.String(), dport)
							}
						}
					}
				} else if next == 17 && len(l4) >= 8 {
					dport := binary.BigEndian.Uint16(l4[2:4])
					if dport == 443 {
						payload := l4[8:]
						if host, ok := sni.ParseQUICClientHelloSNI(payload); ok && host != "" {
							log.Infof("NFQ QUIC SNI v6: %s [%s]:%d -> [%s]:%d", host, src.String(), binary.BigEndian.Uint16(l4[0:2]), dst.String(), dport)
						}
					}
				}
			}
			_ = qx.SetVerdict(id, nfqueue.NfAccept)
			return 0
		}, func(err error) int { return 0 })
		<-w.ctx.Done()
	}(q)
	return nil
}

func (w *Worker) Start() error {
	if err := w.startOne(unix.AF_INET); err != nil {
		return err
	}
	if err := w.startOne(unix.AF_INET6); err != nil {
		return err
	}
	return nil
}

func (w *Worker) Stop() {
	w.cancel()
	for _, q := range w.qs {
		_ = q.Close()
	}
	w.wg.Wait()
}
