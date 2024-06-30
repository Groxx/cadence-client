// Copyright (c) 2017-2021 Uber Technologies Inc.
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

package legacy

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/worker"
)

// RegisterActivityStruct re-implements an [activity.Register] feature prior to v2 that
// allowed passing in a struct, and getting all public methods implicitly registered.
//
// This may be removed eventually.  If you value this behavior, it is quite easy to build externally
// as it uses no real internal details.
//
// deprecated: we do not recommend using this struct-method registration.
// register individual functions instead, or re-implement this in code you control.
func RegisterActivityStruct(wk worker.Worker, aStruct interface{}, options activity.RegisterOptions) {
	registerStruct(aStruct, options, func(fn any, opts activity.RegisterOptions) {
		wk.RegisterActivityWithOptions(fn, opts)
	})
}

// GloballyRegisterActivityStruct re-implements an [activity.Register] feature prior to v2 that
// allowed passing in a struct, and getting all public methods implicitly registered.
//
// This may be removed eventually.  If you value this behavior, it is quite easy to build externally
// as it uses no real internal details.
//
// deprecated: we do not recommend using either global registration or this struct-method registration.
// register individual functions instead, or re-implement this in code you control.
func GloballyRegisterActivityStruct(aStruct interface{}, options activity.RegisterOptions) {
	registerStruct(aStruct, options, func(fn any, opts activity.RegisterOptions) {
		activity.RegisterWithOptions(fn, opts)
	})
}

func registerStruct(aStruct interface{}, options activity.RegisterOptions, register func(fn any, opts activity.RegisterOptions)) {
	structValue := reflect.ValueOf(aStruct)
	structType := structValue.Type()
	count := 0
	for i := 0; i < structValue.NumMethod(); i++ {
		methodValue := structValue.Method(i)
		method := structType.Method(i)
		// skip private method
		if method.PkgPath != "" {
			continue
		}
		methodName := getFunctionName(method.Func.Interface())
		if err := internal.ValidateFnFormat(method.Type, false); err != nil {
			panic(fmt.Errorf("failed to register activity method %v of %v: %e", methodName, structType.Name(), err))
		}

		structPrefix := options.Name
		registerName := methodName

		if len(structPrefix) > 0 {
			registerName = structPrefix + getShortFunctionName(methodName)
		}
		optsDup := options
		optsDup.Name = registerName

		register(methodValue.Interface(), optsDup)

		count++
	}

	if count == 0 {
		panic(fmt.Errorf("no activities (public methods) found in %v structure", structType.Name()))
	}
}

func getShortFunctionName(fnName string) string {
	elements := strings.Split(fnName, ".")
	return elements[len(elements)-1]
}

func getFunctionName(i interface{}) string {
	if fullName, ok := i.(string); ok {
		return fullName
	}
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	// This allows to call activities by method pointer
	// Compiler adds -fm suffix to a function name which has a receiver
	// Note that this works even if struct pointer used to get the function is nil
	// It is possible because nil receivers are allowed.
	// For example:
	// var a *Activities
	// ExecuteActivity(ctx, a.Foo)
	// will call this function which is going to return "Foo"
	return strings.TrimSuffix(fullName, "-fm")
}
