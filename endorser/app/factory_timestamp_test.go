/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: LGPL-3.0-or-later
*/

package app

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-x-evm/endorser/config"
	"github.com/hyperledger/fabric-x-evm/endorser/core"
)

func TestApplyTimestampBounds(t *testing.T) {
	end, err := core.New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Defaults (zero config fields).
	applyTimestampBounds(end, config.Endorser{})
	// Explicit overrides.
	applyTimestampBounds(end, config.Endorser{
		MaxTimestampFuture: 7 * time.Second,
		MaxTimestampPast:   11 * time.Second,
	})
}
