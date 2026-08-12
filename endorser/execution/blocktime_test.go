/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: LGPL-3.0-or-later
*/

package execution

import (
	"context"
	"math/big"
	"testing"

	fxcommon "github.com/hyperledger/fabric-x-evm/common"
	"github.com/hyperledger/fabric-x-sdk/state"
	_ "modernc.org/sqlite"
)

func TestNewExecutor_UsesProvidedBlockTime(t *testing.T) {
	backend, err := state.NewWriteDB(Channel, "file:exec_blocktime?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	kvs := &testVersionedDBSnapshotter{db: backend}
	reader, err := kvs.NewSnapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	stateDB, err := NewStateDB(context.Background(), reader, Namespace, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	cfg := EVMConfig{ChainConfig: fxcommon.BuildChainConfig(4011)}
	want := uint64(1_700_000_123)
	ex, err := NewExecutor(stateDB, reader, big.NewInt(0), want, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if ex.BlockCtx.Time != want {
		t.Fatalf("BlockCtx.Time = %d, want %d", ex.BlockCtx.Time, want)
	}
}

func TestNewExecutor_ZeroBlockTimeFallsBackToDefault(t *testing.T) {
	backend, err := state.NewWriteDB(Channel, "file:exec_blocktime0?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	kvs := &testVersionedDBSnapshotter{db: backend}
	reader, err := kvs.NewSnapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	stateDB, err := NewStateDB(context.Background(), reader, Namespace, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	cfg := EVMConfig{ChainConfig: fxcommon.BuildChainConfig(4011)}
	ex, err := NewExecutor(stateDB, reader, nil, 0, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if ex.BlockCtx.Time != DefaultBlockTime {
		t.Fatalf("BlockCtx.Time = %d, want DefaultBlockTime %d", ex.BlockCtx.Time, DefaultBlockTime)
	}
}
