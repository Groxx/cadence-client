package structfill

import (
	"context"
	"sync"
	"time"
)

// Test struct that embeds default-skipped types
// lint:must-fill
type StructWithDefaults struct { // want StructWithDefaults:""
	Name string
	Mu   sync.Mutex
	Rw   sync.RWMutex
	Wg   sync.WaitGroup
	Once sync.Once
	Cond sync.Cond
	Ctx  context.Context
	T    time.Time
	D    time.Duration
}

// These should be valid without requiring explicit field initialization
// for the sync types in the default skip list (Mu, Rw, Wg, Once)
var example1 = StructWithDefaults{ // want `missing "Cond"` `missing "Ctx"` `missing "T"` `missing "D"`
	Name: "test1",
	// Mu, Rw, Wg, Once are automatically skipped (zero values are safe)
	// but Cond, Ctx, T, D must be explicitly filled
}

// This should trigger errors for missing Name and non-default-skipped fields
var example2 = StructWithDefaults{ // want `missing "Name"` `missing "Cond"` `missing "Ctx"` `missing "T"` `missing "D"`
	// Name is missing and Cond, Ctx, T, D are missing
	// but Mu, Rw, Wg, Once are auto-skipped
}

// This should be completely valid - all required fields are filled
// (Mu, Rw, Wg, Once are auto-skipped and don't need to be specified)
var example3 = StructWithDefaults{
	Name: "test3",
	Cond: sync.Cond{},
	Ctx:  context.Background(),
	T:    time.Now(),
	D:    time.Second,
}
