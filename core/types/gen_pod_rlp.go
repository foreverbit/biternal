// Code generated by rlpgen. DO NOT EDIT.

//go:build !norlpgen
// +build !norlpgen

package types

import "github.com/foreverbit/biternal/rlp"
import "io"

func (obj *StatePod) EncodeRLP(_w io.Writer) error {
	w := rlp.NewEncoderBuffer(_w)
	_tmp0 := w.List()
	w.WriteUint64(obj.GasLimit)
	w.WriteUint64(obj.CurrentGasLimit)
	_tmp1 := w.List()
	for _, _tmp2 := range obj.Passengers {
		w.WriteBytes(_tmp2[:])
	}
	w.ListEnd(_tmp1)
	w.ListEnd(_tmp0)
	return w.Flush()
}
