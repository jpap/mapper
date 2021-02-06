// Copyright 2020 John Papandriopoulos.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package mapper

import (
	"fmt"
	"sync"
	"unsafe"
)

// Mapper maps between Go values and pointers suitable for passing to C via Cgo.
type Mapper struct {
	m sync.Map
}

type mapperKey struct {
	// We can't have an empty struct, otherwise allocations are not distinct.
	_ uint8
}

// G is the global mapper... for users who don't care about lock contention.
// For those that do, it is recommended to use a separate Mapper instance.
var G Mapper

// New creates a new mapping to the Go value v.
//
// The mapping is a pointer that can be passed to C via Cgo.  When Cgo
// calls back into Go, supplying the pointer, the client code can use
// Mapper.Get to retrieve the Go object, after type conversion.
func (mapper *Mapper) New(v interface{}) unsafe.Pointer {
	// Create a new unique token by using the pointer value.
	//
	// This value can safely be passed to C via Cgo because it doesn't
	// contain any pointers to Go memory.
	//
	// We could've also used an atomic counter, and typecasted it to a pointer
	// value; might be a good idea to profile it vs this approach.  The advantage
	// there is that it puts less pressure on the GC.
	k := &mapperKey{}
	mapper.m.Store(k, v)
	return unsafe.Pointer(k)
}

// Get retrieves the Go value v from the Cgo pointer k.
func (mapper *Mapper) Get(k unsafe.Pointer) (v interface{}) {
	var ok bool
	v, ok = mapper.m.Load((*mapperKey)(k))
	if !ok {
		panic(fmt.Errorf("mapper: ptr not mapped: %p", k))
	}
	return
}

// Delete mapping via the Cgo pointer k.
func (mapper *Mapper) Delete(k unsafe.Pointer) {
	mapper.m.Delete(k)
}
