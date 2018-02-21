// Code generated by thriftrw v1.8.0. DO NOT EDIT.
// @generated

package bar_test

import (
	"fmt"
	"go.uber.org/thriftrw/wire"
)

type UUID string

// ToWire translates UUID into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
func (v UUID) ToWire() (wire.Value, error) {
	x := (string)(v)
	return wire.NewValueString(x), error(nil)
}

// String returns a readable string representation of UUID.
func (v UUID) String() string {
	x := (string)(v)
	return fmt.Sprint(x)
}

// FromWire deserializes UUID from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
func (v *UUID) FromWire(w wire.Value) error {
	x, err := w.GetString(), error(nil)
	*v = (UUID)(x)
	return err
}

// Equals returns true if this UUID is equal to the provided
// UUID.
func (lhs UUID) Equals(rhs UUID) bool {
	return (lhs == rhs)
}
