// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package internal

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	errMsgUnknownWorkflowType = "unable to find workflow type"
)

var once sync.Once

// Singleton to hold the host registration details.
var globalRegistry *registry

func newRegistry() *registry {
	return &registry{
		workflowFuncMap:  make(map[string]interface{}),
		workflowAliasMap: make(map[string]string),
		activityFuncMap:  make(map[string]activity),
		activityAliasMap: make(map[string]string),
		next:             getGlobalRegistry(),
	}
}

func getGlobalRegistry() *registry {
	once.Do(func() {
		globalRegistry = &registry{
			workflowFuncMap:  make(map[string]interface{}),
			workflowAliasMap: make(map[string]string),
			activityFuncMap:  make(map[string]activity),
			activityAliasMap: make(map[string]string),
		}
	})
	return globalRegistry
}

type registry struct {
	sync.Mutex
	workflowFuncMap  map[string]interface{}
	workflowAliasMap map[string]string
	activityFuncMap  map[string]activity
	activityAliasMap map[string]string
	next             *registry // Allows to chain registries
}

func (r *registry) RegisterWorkflow(af interface{}) {
	r.RegisterWorkflowWithOptions(af, RegisterWorkflowOptions{})
}

func (r *registry) RegisterWorkflowWithOptions(
	wf interface{},
	options RegisterWorkflowOptions,
) {
	// Validate that it is a function
	fnType := reflect.TypeOf(wf)
	if err := ValidateFnFormat(fnType, true); err != nil {
		panic(err)
	}
	fnName := getFunctionName(wf)
	alias := options.Name
	registerName := fnName

	if len(alias) > 0 {
		registerName = alias
	}

	r.Lock()
	defer r.Unlock()

	if !options.DisableAlreadyRegisteredCheck {
		if _, ok := r.getWorkflowNoLock(registerName); ok {
			panic(fmt.Sprintf("workflow name \"%v\" is already registered", registerName))
		}
	}
	r.workflowFuncMap[registerName] = wf
	if len(alias) > 0 {
		r.workflowAliasMap[fnName] = registerName
	}
}

func (r *registry) RegisterActivity(af interface{}) {
	r.RegisterActivityWithOptions(af, RegisterActivityOptions{})
}

func (r *registry) RegisterActivityWithOptions(af interface{}, options RegisterActivityOptions) {
	err := r.registerActivityFunction(af, options)
	if err != nil {
		panic(err)
	}
}

func (r *registry) GetRegisteredWorkflows() []string {
	return r.GetRegisteredWorkflowTypes()
}

func (r *registry) GetWorkflowFunc(registerName string) (interface{}, bool) {
	return r.getWorkflowFn(registerName)
}

func (r *registry) GetRegisteredActivities() []string {
	activities := r.getRegisteredActivities()
	activityNames := make([]string, 0, len(activities))
	for _, a := range activities {
		activityNames = append(activityNames, a.ActivityType().Name)
	}
	return activityNames
}

func (r *registry) GetActivityFunc(registerName string) (interface{}, bool) {
	a, ok := r.GetActivity(registerName)
	if !ok {
		return nil, false
	}
	return a.GetFunction(), ok
}

func (r *registry) registerActivityFunction(af interface{}, options RegisterActivityOptions) error {
	fnType := reflect.TypeOf(af)
	if err := ValidateFnFormat(fnType, false); err != nil {
		return fmt.Errorf("failed to register activity method: %v", err)
	}

	fnName := getFunctionName(af)
	alias := options.Name
	registerName := fnName

	if len(alias) > 0 {
		registerName = alias
	}

	r.Lock()
	defer r.Unlock()

	if !options.DisableAlreadyRegisteredCheck {
		if _, ok := r.getActivityNoLock(registerName); ok {
			return fmt.Errorf("activity type \"%v\" is already registered", registerName)
		}
	}
	r.activityFuncMap[registerName] = &activityExecutor{registerName, af, options}
	if len(alias) > 0 {
		r.activityAliasMap[fnName] = registerName
	}

	return nil
}

func (r *registry) getWorkflowAlias(fnName string) (string, bool) {
	r.Lock() // do not defer for Unlock to call next.getWorkflowAlias without lock
	alias, ok := r.workflowAliasMap[fnName]
	if !ok && r.next != nil {
		r.Unlock()
		return r.next.getWorkflowAlias(fnName)
	}
	r.Unlock()
	return alias, ok
}

func (r *registry) getWorkflowFn(fnName string) (interface{}, bool) {
	r.Lock() // do not defer for Unlock to call next.getWorkflowFn without lock
	fn, ok := r.workflowFuncMap[fnName]
	if !ok { // if exact match is not found, check for backwards compatible name without -fm suffix
		fn, ok = r.workflowFuncMap[strings.TrimSuffix(fnName, "-fm")]
	}
	if !ok && r.next != nil {
		r.Unlock()
		return r.next.getWorkflowFn(fnName)
	}
	r.Unlock()
	return fn, ok
}

func (r *registry) getWorkflowNoLock(registerName string) (interface{}, bool) {
	a, ok := r.workflowFuncMap[registerName]
	if !ok && r.next != nil {
		return r.next.getWorkflowNoLock(registerName)
	}
	return a, ok
}

func (r *registry) GetRegisteredWorkflowTypes() []string {
	r.Lock() // do not defer for Unlock to call next.getRegisteredWorkflowTypes without lock
	var result []string
	for t := range r.workflowFuncMap {
		result = append(result, t)
	}
	r.Unlock()
	if r.next != nil {
		nextTypes := r.next.GetRegisteredWorkflowTypes()
		result = append(result, nextTypes...)
	}
	return result
}

func (r *registry) getActivityAlias(fnName string) (string, bool) {
	r.Lock() // do not defer for Unlock to call next.getActivityAlias without lock
	alias, ok := r.activityAliasMap[fnName]
	if !ok && r.next != nil {
		r.Unlock()
		return r.next.getActivityAlias(fnName)
	}
	r.Unlock()
	return alias, ok
}

// Use in unit test only, otherwise deadlock will occur.
func (r *registry) addActivityWithLock(fnName string, a activity) {
	r.Lock()
	defer r.Unlock()
	r.activityFuncMap[fnName] = a
}

func (r *registry) GetActivity(fnName string) (activity, bool) {
	r.Lock() // do not defer for Unlock to call next.GetActivity without lock
	a, ok := r.activityFuncMap[fnName]
	if !ok { // if exact match is not found, check for backwards compatible name without -fm suffix
		a, ok = r.activityFuncMap[strings.TrimSuffix(fnName, "-fm")]
	}
	if !ok && r.next != nil {
		r.Unlock()
		return r.next.GetActivity(fnName)
	}
	r.Unlock()
	return a, ok
}

func (r *registry) getActivityNoLock(registerName string) (activity, bool) {
	a, ok := r.activityFuncMap[registerName]
	if !ok && r.next != nil {
		return r.next.getActivityNoLock(registerName)
	}
	return a, ok
}

func (r *registry) getRegisteredActivities() []activity {
	r.Lock() // do not defer for Unlock to call next.getRegisteredActivities without lock
	activities := make([]activity, 0, len(r.activityFuncMap))
	for _, a := range r.activityFuncMap {
		activities = append(activities, a)
	}
	r.Unlock()
	if r.next != nil {
		nextActivities := r.next.getRegisteredActivities()
		activities = append(activities, nextActivities...)
	}
	return activities
}

func (r *registry) getWorkflowDefinition(wt WorkflowType) (workflowDefinition, error) {
	lookup := getFunctionName(wt.Name)
	if alias, ok := r.getWorkflowAlias(lookup); ok {
		lookup = alias
	}
	wf, ok := r.getWorkflowFn(lookup)
	if !ok {
		supported := strings.Join(r.GetRegisteredWorkflowTypes(), ", ")
		return nil, fmt.Errorf(errMsgUnknownWorkflowType+": %v. Supported types: [%v]", lookup, supported)
	}
	wd := &workflowExecutor{workflowType: lookup, fn: wf}
	return newSyncWorkflowDefinition(wd), nil
}
