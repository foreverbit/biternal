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
	"math/big"
)

type StateType struct {
	Prefix byte // One prefix for key and {value}
}

var (
	AccountState = StateType{'a'}
	PodState     = StateType{'p'}
)

func stateTypeFromPrefix(prefix byte) StateType {
	switch prefix {
	case 'a':
		return AccountState
	case 'p':
		return PodState
	default:
		panic("unknown state type")
	}
}

// PreKey Key before hash
func (st StateType) PreKey(key []byte) []byte {
	return append([]byte{st.Prefix}, key...)
}

func (st StateType) Value(value []byte) []byte {
	return append([]byte{st.Prefix}, value...)
}

func accountKey(addr common.Address) []byte {
	return AccountState.PreKey(addr.Bytes())
}

func keyToAddress(key []byte) common.Address {
	return common.BytesToAddress(key[1:])
}

func podKey(block *big.Int) []byte {
	return PodState.PreKey(block.Bytes())
}

func keyToBlock(key []byte) *big.Int {
	return new(big.Int).SetBytes(key[1:])
}

type Sender struct {
	Addr common.Address
}

func (s *Sender) Address() common.Address {
	return s.Addr
}
