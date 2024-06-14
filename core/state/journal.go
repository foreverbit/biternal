// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"math/big"

	"github.com/foreverbit/biternal/common"
)

// journalEntry is a modification entry in the state change journal that can be
// reverted on demand.
type journalEntry interface {
	// revert undoes the changes introduced by this journal entry.
	revert(*StateDB)

	// dirtied returns the Ethereum object hash modified by this journal entry.
	dirtied() *common.Hash
}

// journal contains the list of state modifications applied since the last state
// commit. These are tracked to be able to be reverted in the case of an execution
// exception or request for reversal.
type journal struct {
	entries []journalEntry      // Current changes tracked by the journal
	dirties map[common.Hash]int // Dirty accounts and the number of changes
}

// newJournal creates a new initialized journal.
func newJournal() *journal {
	return &journal{
		dirties: make(map[common.Hash]int),
	}
}

// append inserts a new modification entry to the end of the change journal.
func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
	if keyHash := entry.dirtied(); keyHash != nil {
		j.dirties[*keyHash]++
	}
}

// revert undoes a batch of journalled modifications along with any reverted
// dirty handling too.
func (j *journal) revert(statedb *StateDB, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		// Undo the changes made by the operation
		j.entries[i].revert(statedb)

		// Drop any dirty tracking induced by the change
		if keyHash := j.entries[i].dirtied(); keyHash != nil {
			if j.dirties[*keyHash]--; j.dirties[*keyHash] == 0 {
				delete(j.dirties, *keyHash)
			}
		}
	}
	j.entries = j.entries[:snapshot]
}

// dirty explicitly sets an object to dirty, even if the change entries would
// otherwise suggest it as clean. This method is an ugly hack to handle the RIPEMD
// precompile consensus exception.
func (j *journal) dirty(keyHash common.Hash) {
	j.dirties[keyHash]++
}

// length returns the current number of entries in the journal.
func (j *journal) length() int {
	return len(j.entries)
}

// Common journal entries
type (
	createObjectChange struct {
		key []byte // key bytes, include prefix
	}
	resetObjectChange struct {
		prev         stateObject
		prevdestruct bool
	}
)

// Account related journal entries
type (
	// Changes to the account trie.
	suicideChange struct {
		account     *common.Address
		prev        bool // whether account had already suicided
		prevbalance *big.Int
	}

	// Changes to individual accounts.
	balanceChange struct {
		account *common.Address
		prev    *big.Int
	}
	nonceChange struct {
		account *common.Address
		prev    uint64
	}
	storageChange struct {
		account       *common.Address
		key, prevalue common.Hash
	}
	codeChange struct {
		account            *common.Address
		prevcode, prevhash []byte
	}

	// Changes to other state values.
	refundChange struct {
		prev uint64
	}
	addLogChange struct {
		txhash common.Hash
	}
	addPreimageChange struct {
		hash common.Hash
	}
	touchChange struct {
		account *common.Address
	}
	// Changes to the access list
	accessListAddAccountChange struct {
		address *common.Address
	}
	accessListAddSlotChange struct {
		address *common.Address
		slot    *common.Hash
	}
)

func (ch createObjectChange) revert(s *StateDB) {
	keyHash := s.hashKey(ch.key)

	delete(s.stateObjects, keyHash)
	delete(s.stateObjectsDirty, keyHash)
}

func (ch createObjectChange) dirtied() *common.Hash {
	stateType := stateTypeFromPrefix(ch.key[0])
	switch stateType {
	case AccountState:
		addr := common.BytesToAddress(ch.key[1:])
		keyHash := accountKeyHash(addr)
		return &keyHash
	case PodState:
		block := keyToBlock(ch.key[1:])
		keyHash := podKeyHash(block)
		return &keyHash
	default:
		return nil
	}
}

func (ch resetObjectChange) revert(s *StateDB) {
	s.setStateObject(ch.prev)
	if !ch.prevdestruct && s.snap != nil {
		delete(s.snapDestructs, ch.prev.KeyHash())
	}
}

func (ch resetObjectChange) dirtied() *common.Hash {
	return nil
}

func (ch suicideChange) revert(s *StateDB) {
	obj := s.getStateObject(accountKey(*ch.account))
	if obj != nil {
		ao := getAccountObject(obj)
		ao.suicided = ch.prev
		ao.setBalance(ch.prevbalance)
	}
}

func (ch suicideChange) dirtied() *common.Hash {
	keyHash := accountKeyHash(*ch.account)
	return &keyHash
}

var ripemd = common.HexToAddress("0000000000000000000000000000000000000003")

func (ch touchChange) revert(s *StateDB) {
}

func (ch touchChange) dirtied() *common.Hash {
	keyHash := accountKeyHash(*ch.account)
	return &keyHash
}

func (ch balanceChange) revert(s *StateDB) {
	obj := s.getStateObject(accountKey(*ch.account))
	getAccountObject(obj).setBalance(ch.prev)
}

func (ch balanceChange) dirtied() *common.Hash {
	keyHash := accountKeyHash(*ch.account)
	return &keyHash
}

func (ch nonceChange) revert(s *StateDB) {
	obj := s.getStateObject(accountKey(*ch.account))
	getAccountObject(obj).setNonce(ch.prev)
}

func (ch nonceChange) dirtied() *common.Hash {
	keyHash := accountKeyHash(*ch.account)
	return &keyHash
}

func (ch codeChange) revert(s *StateDB) {
	obj := s.getStateObject(accountKey(*ch.account))
	getAccountObject(obj).setCode(common.BytesToHash(ch.prevhash), ch.prevcode)
}

func (ch codeChange) dirtied() *common.Hash {
	keyHash := accountKeyHash(*ch.account)
	return &keyHash
}

func (ch storageChange) revert(s *StateDB) {
	obj := s.getStateObject(accountKey(*ch.account))
	getAccountObject(obj).setState(ch.key, ch.prevalue)
}

func (ch storageChange) dirtied() *common.Hash {
	keyHash := accountKeyHash(*ch.account)
	return &keyHash
}

func (ch refundChange) revert(s *StateDB) {
	s.refund = ch.prev
}

func (ch refundChange) dirtied() *common.Hash {
	return nil
}

func (ch addLogChange) revert(s *StateDB) {
	logs := s.logs[ch.txhash]
	if len(logs) == 1 {
		delete(s.logs, ch.txhash)
	} else {
		s.logs[ch.txhash] = logs[:len(logs)-1]
	}
	s.logSize--
}

func (ch addLogChange) dirtied() *common.Hash {
	return nil
}

func (ch addPreimageChange) revert(s *StateDB) {
	delete(s.preimages, ch.hash)
}

func (ch addPreimageChange) dirtied() *common.Hash {
	return nil
}

func (ch accessListAddAccountChange) revert(s *StateDB) {
	/*
		One important invariant here, is that whenever a (addr, slot) is added, if the
		addr is not already present, the add causes two journal entries:
		- one for the address,
		- one for the (address,slot)
		Therefore, when unrolling the change, we can always blindly delete the
		(addr) at this point, since no storage adds can remain when come upon
		a single (addr) change.
	*/
	s.accessList.DeleteAddress(*ch.address)
}

func (ch accessListAddAccountChange) dirtied() *common.Hash {
	return nil
}

func (ch accessListAddSlotChange) revert(s *StateDB) {
	s.accessList.DeleteSlot(*ch.address, *ch.slot)
}

func (ch accessListAddSlotChange) dirtied() *common.Hash {
	return nil
}
