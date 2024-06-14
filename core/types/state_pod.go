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

package types

import (
	"github.com/foreverbit/biternal/common"
)

//go:generate go run ../../rlp/rlpgen -type StatePod -out gen_pod_rlp.go

type StatePod struct {
	GasLimit        uint64
	CurrentGasLimit uint64
	Passengers      []common.Address
}
