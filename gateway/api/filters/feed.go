/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: LGPL-3.0-or-later
*/

package filters

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/hyperledger/fabric-lib-go/common/flogging"
	"github.com/hyperledger/fabric-x-sdk/blocks"
)

const internalBuffer = 32

var feedLogger = flogging.MustGetLogger("gateway.api.filters")

// blockSink receives committed blocks off the synchronizer hot path.
type blockSink interface {
	onBlock(b blocks.Block)
}

// BlockFeed implements blocks.BlockHandler. Handle never blocks on slow RPC
// consumers: it does a non-blocking send into an internal buffer and returns.
// A dedicated drain goroutine fans blocks out to the registered sink.
type BlockFeed struct {
	in   chan blocks.Block
	quit chan struct{}
	done chan struct{}

	sink atomic.Value // blockSink

	dropMu   sync.Mutex
	dropping bool
}

// NewBlockFeed starts the drain goroutine. Call Close on shutdown.
func NewBlockFeed() *BlockFeed {
	f := &BlockFeed{
		in:   make(chan blocks.Block, internalBuffer),
		quit: make(chan struct{}),
		done: make(chan struct{}),
	}
	go f.drain()
	return f
}

// SetSink registers the consumer that receives blocks from the drain loop.
// Typically the FilterAPI. Safe to call once during wiring before traffic.
func (f *BlockFeed) SetSink(s blockSink) {
	f.sink.Store(s)
}

// Handle enqueues the block without waiting for filter delivery.
func (f *BlockFeed) Handle(_ context.Context, b blocks.Block) error {
	select {
	case f.in <- b:
		f.clearDrop()
	default:
		f.noteDrop()
	}
	return nil
}

// Close stops the drain goroutine and waits for it to exit.
func (f *BlockFeed) Close() {
	select {
	case <-f.quit:
		return
	default:
		close(f.quit)
	}
	<-f.done
}

// Done is closed after Close finishes draining.
func (f *BlockFeed) Done() <-chan struct{} {
	return f.done
}

func (f *BlockFeed) drain() {
	defer close(f.done)
	for {
		select {
		case <-f.quit:
			// Drain anything already queued so Close does not race a late Handle.
			for {
				select {
				case b := <-f.in:
					f.deliver(b)
				default:
					return
				}
			}
		case b := <-f.in:
			f.deliver(b)
		}
	}
}

func (f *BlockFeed) deliver(b blocks.Block) {
	v := f.sink.Load()
	if v == nil {
		return
	}
	v.(blockSink).onBlock(b)
}

func (f *BlockFeed) noteDrop() {
	f.dropMu.Lock()
	defer f.dropMu.Unlock()
	if f.dropping {
		return
	}
	f.dropping = true
	feedLogger.Warnf("block filter feed falling behind; dropping notifications until consumers catch up")
}

func (f *BlockFeed) clearDrop() {
	f.dropMu.Lock()
	defer f.dropMu.Unlock()
	f.dropping = false
}
