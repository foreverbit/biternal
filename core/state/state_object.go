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
	"github.com/foreverbit/biternal/crypto"
	"github.com/foreverbit/biternal/trie"
	"math/big"
)

type stateObject interface {
	Type() StateType
	Key() []byte
	KeyHash() common.Hash
	ValueBytes() ([]byte, error)
	DeepCopy(s *StateDB) stateObject
	Empty() bool

	Suicided() bool

	Deleted() bool
	MarkDeleted()

	SnapRLP() []byte

	Commit(s *StateDB) (*trie.NodeSet, error)
}

func accountKeyHash(addr common.Address) common.Hash {
	return crypto.Keccak256Hash(accountKey(addr))
}

func podKeyHash(block *big.Int) common.Hash {
	return crypto.Keccak256Hash(podKey(block))
}

func getAccountObject(so stateObject) *accountObject {
	if so.Type() != AccountState {
		return nil
	}
	return so.(*accountObject)
}

func getPodObject(so stateObject) *podObject {
	if so.Type() != PodState {
		return nil
	}
	return so.(*podObject)
}
