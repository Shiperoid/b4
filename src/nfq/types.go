package nfq

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/daniellavrushin/b4/dhcp"
	"github.com/daniellavrushin/b4/sock"
	"github.com/florianl/go-nfqueue"
)

type Pool struct {
	Workers  []*Worker
	configMu sync.Mutex
	Dhcp     *dhcp.Manager
}

type Worker struct {
	packetsProcessed uint64
	cfg              atomic.Value
	qnum             uint16
	ctx              context.Context
	cancel           context.CancelFunc
	q                *nfqueue.Nfqueue
	wg               sync.WaitGroup
	matcher          atomic.Value
	sock             *sock.Sender
	ipToMac          atomic.Value
}
