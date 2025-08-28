package abi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"web3_gui/chain/evm/common"
)

func parseTopicWithSetter(fields Arguments, topics []common.Hash, setter func(Argument, interface{})) error {
	// Sanity check that the fields and topics match up
	if len(fields) != len(topics) {
		return errors.New("topic/field count mismatch")
	}
	// Iterate over all the fields and reconstruct them from topics
	for i, arg := range fields {
		if !arg.Indexed {
			return errors.New("non-indexed field in topic reconstruction")
		}
		var reconstr interface{}
		switch arg.Type.T {
		case TupleTy:
			return errors.New("tuple type in topic reconstruction")
		case StringTy, BytesTy, SliceTy, ArrayTy:
			// Array types (including strings and bytes) have their keccak256 hashes stored in the topic- not a hash
			// whose bytes can be decoded to the actual value- so the best we can do is retrieve that hash
			reconstr = topics[i]
		case FunctionTy:
			if garbage := binary.BigEndian.Uint64(topics[i][0:8]); garbage != 0 {
				return fmt.Errorf("bind: got improperly encoded function type, got %v", topics[i].Bytes())
			}
			var tmp [24]byte
			copy(tmp[:], topics[i][8:32])
			reconstr = tmp
		default:
			var err error
			reconstr, err = toGoType(0, arg.Type, topics[i].Bytes())
			if err != nil {
				return err
			}
		}
		// Use the setter function to store the value
		setter(arg, reconstr)
	}

	return nil
}
func ParseTopics(out interface{}, fields Arguments, topics []common.Hash) error {
	return parseTopicWithSetter(fields, topics,
		func(arg Argument, reconstr interface{}) {
			field := reflect.ValueOf(out).Elem().FieldByName(ToCamelCase(arg.Name))
			field.Set(reflect.ValueOf(reconstr))
		})
}
