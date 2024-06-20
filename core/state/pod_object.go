// Copyright 2024 The Biternal Authors
// This file is part of the biternal library.
//
// The biternal library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The biternal library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the biternal library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"github.com/foreverbit/biternal/common"
	"github.com/foreverbit/biternal/core/types"
	"github.com/foreverbit/biternal/rlp"
	"github.com/foreverbit/biternal/trie"
	"math/big"
)

type podObject struct {
	block   *big.Int
	podHash common.Hash
	data    types.StatePod

	db    *StateDB
	dbErr error

	// When one pod is executed, it will be marked as executed
	executed bool
	deleted  bool
}

func newPodObject(db *StateDB, block *big.Int, data types.StatePod) *podObject {
	if data.GasLimit == 0 {
		// TODO: need one algorithm to calculate the gas limit in future block
		data.GasLimit = 1000000
	}
	return &podObject{
		block:   block,
		podHash: podKeyHash(block),
		data:    data,
		db:      db,
	}
}

/// Implement stateObject interface for podObject

func (o *podObject) Type() StateType {
	return PodState
}

func (o *podObject) Key() []byte {
	return podKey(o.block)
}

func (o *podObject) ValueBytes() ([]byte, error) {
	return rlp.EncodeToBytes(o.data)
}

func (o *podObject) KeyHash() common.Hash {
	return o.podHash
}

func (o *podObject) DeepCopy(db *StateDB) stateObject {
	// TODO
	return nil
}

func (o *podObject) Empty() bool {
	return o.block == nil && o.data.GasLimit == 0
}

func (o *podObject) Suicided() bool {
	return o.executed
}

func (o *podObject) Deleted() bool {
	return o.deleted
}

func (o *podObject) MarkDeleted() {
	o.deleted = true
}

func (o *podObject) SnapRLP() []byte {
	// TODO
	return nil
}

func (o *podObject) Commit(s *StateDB) (*trie.NodeSet, error) {
	// TODO
	return nil, nil
}
