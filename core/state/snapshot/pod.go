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

package snapshot

import (
	"github.com/foreverbit/biternal/common"
	"github.com/foreverbit/biternal/core/types"
	"github.com/foreverbit/biternal/rlp"
)

type Pod struct {
	GasLimit        uint64
	CurrentGasLimit uint64
	Passengers      [][]byte
}

func SlimPod(gasLimit uint64, currentGasLimit uint64, passengers []common.Address) Pod {
	slim := Pod{
		GasLimit:        gasLimit,
		CurrentGasLimit: currentGasLimit,
	}
	for _, p := range passengers {
		slim.Passengers = append(slim.Passengers, p[:])
	}
	return slim
}

func SlimPodRLP(gasLimit uint64, currentGasLimit uint64, passengers []common.Address) []byte {
	data, err := rlp.EncodeToBytes(SlimPod(gasLimit, currentGasLimit, passengers))
	if err != nil {
		panic(err)
	}
	return data
}

func FullPod(data []byte) (Pod, error) {
	var pod Pod
	if err := rlp.DecodeBytes(data, &pod); err != nil {
		return Pod{}, err
	}
	return pod, nil
}

func FullPodRLP(data []byte) ([]byte, error) {
	pod, err := FullPod(data)
	if err != nil {
		return nil, err
	}
	return rlp.EncodeToBytes(pod)
}

func FullStatePod(data []byte) (types.StatePod, error) {
	pod, err := FullPod(data)
	if err != nil {
		return types.StatePod{}, err
	}
	var passengers []common.Address
	for _, p := range pod.Passengers {
		passengers = append(passengers, common.BytesToAddress(p))
	}

	return types.StatePod{
		GasLimit:        pod.GasLimit,
		CurrentGasLimit: pod.CurrentGasLimit,
		Passengers:      passengers,
	}, nil
}
