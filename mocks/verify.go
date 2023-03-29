// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package mocks

import (
	"go.uber.org/cadence/client"
	internal "go.uber.org/cadence/internal"
	"go.uber.org/cadence/worker"
)

// use `make go-generate` to re-generate mocks
//go:generate mockery --dir=../client --name=Client --quiet
//go:generate mockery --dir=../client --name=DomainClient --quiet
//go:generate mockery --dir=../internal --name=HistoryEventIterator --quiet
//go:generate mockery --dir=../internal --name=Value --quiet
//go:generate mockery --dir=../internal --name=WorkflowRun --quiet
//go:generate mockery --dir=../worker --name=Registry --quiet
//go:generate mockery --dir=../worker --name=Worker --quiet

// make sure mocks are in sync with interfaces
var _ client.Client = (*Client)(nil)
var _ client.DomainClient = (*DomainClient)(nil)
var _ internal.HistoryEventIterator = (*HistoryEventIterator)(nil)
var _ internal.Value = (*Value)(nil)
var _ internal.WorkflowRun = (*WorkflowRun)(nil)
var _ worker.Registry = (*Registry)(nil)
var _ worker.Worker = (*Worker)(nil)
