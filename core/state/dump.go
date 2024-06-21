// Copyright 2014 The go-ethereum Authors
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
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/foreverbit/biternal/common"
	"github.com/foreverbit/biternal/common/hexutil"
	"github.com/foreverbit/biternal/core/types"
	"github.com/foreverbit/biternal/log"
	"github.com/foreverbit/biternal/rlp"
	"github.com/foreverbit/biternal/trie"
)

// DumpConfig is a set of options to control what portions of the statewill be
// iterated and collected.
type DumpConfig struct {
	SkipCode          bool
	SkipStorage       bool
	OnlyWithAddresses bool
	Start             []byte
	Max               uint64
}

// DumpCollector interface which the state trie calls during iteration
type DumpCollector interface {
	// OnRoot is called with the state root
	OnRoot(common.Hash)
	// OnAccount is called once for each account in the trie
	OnAccount(common.Address, DumpAccount)
	// OnPod is called once for each pod in the trie
	OnPod(*big.Int, DumpPod)
}

// DumpAccount represents an account in the state.
type DumpAccount struct {
	Balance   string                 `json:"balance"`
	Nonce     uint64                 `json:"nonce"`
	Root      hexutil.Bytes          `json:"root"`
	CodeHash  hexutil.Bytes          `json:"codeHash"`
	Code      hexutil.Bytes          `json:"code,omitempty"`
	Storage   map[common.Hash]string `json:"storage,omitempty"`
	Address   *common.Address        `json:"address,omitempty"` // Address only present in iterative (line-by-line) mode
	SecureKey hexutil.Bytes          `json:"key,omitempty"`     // If we don't have address, we can output the key
}

// DumpPod represents a pod in the state.
type DumpPod struct {
	GasLimit        uint64           `json:"gasLimit"`
	CurrentGasLimit uint64           `json:"currentGasLimit"`
	Passengers      []common.Address `json:"passengers"`
	Block           *big.Int         `json:"block,omitempty"`
	SecureKey       hexutil.Bytes    `json:"key,omitempty"`
}

// Dump represents the full dump in a collected format, as one large map.
type Dump struct {
	Root     string                         `json:"root"`
	Accounts map[common.Address]DumpAccount `json:"accounts"`
	Pods     map[*big.Int]DumpPod           `json:"pods"`
}

// OnRoot implements DumpCollector interface
func (d *Dump) OnRoot(root common.Hash) {
	d.Root = fmt.Sprintf("%x", root)
}

// OnAccount implements DumpCollector interface
func (d *Dump) OnAccount(addr common.Address, account DumpAccount) {
	d.Accounts[addr] = account
}

func (d *Dump) OnPod(block *big.Int, pod DumpPod) {
	d.Pods[block] = pod
}

// IteratorDump is an implementation for iterating over data.
type IteratorDump struct {
	Root     string                         `json:"root"`
	Accounts map[common.Address]DumpAccount `json:"accounts"`
	Pods     map[*big.Int]DumpPod           `json:"pods"`
	Next     []byte                         `json:"next,omitempty"` // nil if no more objects
}

// OnRoot implements DumpCollector interface
func (d *IteratorDump) OnRoot(root common.Hash) {
	d.Root = fmt.Sprintf("%x", root)
}

// OnAccount implements DumpCollector interface
func (d *IteratorDump) OnAccount(addr common.Address, account DumpAccount) {
	d.Accounts[addr] = account
}

// OnPod implements DumpCollector interface
func (d *IteratorDump) OnPod(block *big.Int, pod DumpPod) {
	d.Pods[block] = pod
}

// iterativeDump is a DumpCollector-implementation which dumps output line-by-line iteratively.
type iterativeDump struct {
	*json.Encoder
}

// OnAccount implements DumpCollector interface
func (d iterativeDump) OnAccount(addr common.Address, account DumpAccount) {
	dumpAccount := &DumpAccount{
		Balance:   account.Balance,
		Nonce:     account.Nonce,
		Root:      account.Root,
		CodeHash:  account.CodeHash,
		Code:      account.Code,
		Storage:   account.Storage,
		SecureKey: account.SecureKey,
		Address:   nil,
	}
	if addr != (common.Address{}) {
		dumpAccount.Address = &addr
	}
	d.Encode(dumpAccount)
}

func (d iterativeDump) OnPod(block *big.Int, pod DumpPod) {
	dumpPod := &DumpPod{
		GasLimit:        pod.GasLimit,
		CurrentGasLimit: pod.CurrentGasLimit,
		Passengers:      pod.Passengers,
		Block:           block,
		SecureKey:       pod.SecureKey,
	}
	d.Encode(dumpPod)
}

// OnRoot implements DumpCollector interface
func (d iterativeDump) OnRoot(root common.Hash) {
	d.Encode(struct {
		Root common.Hash `json:"root"`
	}{root})
}

// DumpToCollector iterates the state according to the given options and inserts
// the items into a collector for aggregation or serialization.
func (s *StateDB) DumpToCollector(c DumpCollector, conf *DumpConfig) (nextKey []byte) {
	// Sanitize the input to allow nil configs
	if conf == nil {
		conf = new(DumpConfig)
	}
	var (
		missingPreimages int
		accounts         uint64
		pods             uint64
		start            = time.Now()
		logged           = time.Now()
	)
	log.Info("Trie dumping started", "root", s.trie.Hash())
	c.OnRoot(s.trie.Hash())

	it := trie.NewIterator(s.trie.NodeIterator(conf.Start))
	for it.Next() {
		value := it.Value
		stateType := stateTypeFromPrefix(value[0])
		value = value[1:]

		switch stateType {
		case AccountState:
			var data types.StateAccount
			if err := rlp.DecodeBytes(value, &data); err != nil {
				panic(err)
			}
			account := DumpAccount{
				Balance:   data.Balance.String(),
				Nonce:     data.Nonce,
				Root:      data.Root[:],
				CodeHash:  data.CodeHash,
				SecureKey: it.Key,
			}
			addrBytes := s.trie.GetKey(it.Key)
			if addrBytes == nil {
				// Preimage missing
				missingPreimages++
				if conf.OnlyWithAddresses {
					continue
				}
			}
			// TODO: what if addrBytes is nil?
			addr := common.BytesToAddress(addrBytes[1:])
			obj := newAccountObject(s, addr, data)
			if !conf.SkipCode {
				account.Code = obj.Code(s.db)
			}
			if !conf.SkipStorage {
				account.Storage = make(map[common.Hash]string)
				storageIt := trie.NewIterator(obj.getTrie(s.db).NodeIterator(nil))
				for storageIt.Next() {
					_, content, _, err := rlp.Split(storageIt.Value)
					if err != nil {
						log.Error("Failed to decode the value returned by iterator", "error", err)
						continue
					}
					account.Storage[common.BytesToHash(s.trie.GetKey(storageIt.Key))] = common.Bytes2Hex(content)
				}
			}
			c.OnAccount(addr, account)
			accounts++
			break
		case PodState:
			var data types.StatePod
			if err := rlp.DecodeBytes(value, &data); err != nil {
				panic(err)
			}
			pod := DumpPod{
				GasLimit:        data.GasLimit,
				CurrentGasLimit: data.CurrentGasLimit,
				Passengers:      data.Passengers,
				SecureKey:       it.Key,
			}
			blockBytes := s.trie.GetKey(it.Key)
			if blockBytes == nil {
				// Preimage missing
				missingPreimages++
			}
			// TODO: what if blockBytes is nil?
			block := new(big.Int).SetBytes(blockBytes[1:])
			c.OnPod(block, pod)
			pods++
			break
		default:
			panic("unknown state type in dump trie")
		}

		if time.Since(logged) > 8*time.Second {
			log.Info("Trie dumping in progress", "at", it.Key, "accounts", accounts, "pods", pods,
				"elapsed", common.PrettyDuration(time.Since(start)))
			logged = time.Now()
		}
		// TODO include pod max config
		if conf.Max > 0 && accounts >= conf.Max {
			if it.Next() {
				nextKey = it.Key
			}
			break
		}
	}
	if missingPreimages > 0 {
		log.Warn("Dump incomplete due to missing preimages", "missing", missingPreimages)
	}
	log.Info("Trie dumping complete", "accounts", accounts, "pods", pods,
		"elapsed", common.PrettyDuration(time.Since(start)))

	return nextKey
}

// RawDump returns the entire state a single large object
func (s *StateDB) RawDump(opts *DumpConfig) Dump {
	dump := &Dump{
		Accounts: make(map[common.Address]DumpAccount),
		Pods:     make(map[*big.Int]DumpPod),
	}
	s.DumpToCollector(dump, opts)
	return *dump
}

// Dump returns a JSON string representing the entire state as a single json-object
func (s *StateDB) Dump(opts *DumpConfig) []byte {
	dump := s.RawDump(opts)
	result, err := json.MarshalIndent(dump, "", "    ")
	if err != nil {
		fmt.Println("Dump err", err)
	}
	return result
}

// IterativeDump dumps out accounts/pods as json-objects, delimited by linebreaks on stdout
func (s *StateDB) IterativeDump(opts *DumpConfig, output *json.Encoder) {
	s.DumpToCollector(iterativeDump{output}, opts)
}

// IteratorDump dumps out a batch of accounts/pod starts with the given start key
func (s *StateDB) IteratorDump(opts *DumpConfig) IteratorDump {
	iterator := &IteratorDump{
		Accounts: make(map[common.Address]DumpAccount),
		Pods:     make(map[*big.Int]DumpPod),
	}
	iterator.Next = s.DumpToCollector(iterator, opts)
	return *iterator
}
