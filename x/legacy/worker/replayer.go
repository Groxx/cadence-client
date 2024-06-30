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

/*
Package worker contains the "shadower" and "replayer" code that used to live in [go.uber.org/cadence/worker].

This package is *mostly* stable and useful, but we have discovered a significant number of flawed edge cases and
inflexible details that mean we will not be supporting this API in the future.

In particular:
  - Replayer is *extremely* useful, but it has some strange edge case bugs, and does not expose enough information
    to use it in almost any reliable way.  We intend to re-implement this at some point, and will keep the functionality
    intact and maintained until then (though we may make small API changes if needed).
  - Shadower is likely to be deleted eventually.  It can be relatively easily re-implemented in external code anyway.
    It was originally built up for a feature that we never completed and have since decided we do not want at all.
*/
package worker

import (
	"context"
	"io"

	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

type (
	// WorkflowReplayer supports replaying a workflow from its event history.
	// Use for troubleshooting and backwards compatibility unit tests.
	// For example if a workflow failed in production then its history can be downloaded through UI or CLI
	// and replayed in a debugger as many times as necessary.
	// Use this class to create unit tests that check if workflow changes are backwards compatible.
	// It is important to maintain backwards compatibility through use of workflow.GetVersion
	// to ensure that new deployments are not going to break open workflows.
	WorkflowReplayer interface {
		worker.WorkflowRegistry
		worker.ActivityRegistry

		// ReplayWorkflowHistory executes a single decision task for the given json history file.
		// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
		// The logger is an optional parameter. Defaults to the noop logger.
		ReplayWorkflowHistory(logger *zap.Logger, history *shared.History) error

		// ReplayWorkflowHistoryFromJSONFile executes a single decision task for the json history file downloaded from the cli.
		// To download the history file: cadence workflow showid <workflow_id> -of <output_filename>
		// See https://github.com/uber/cadence/blob/master/tools/cli/README.md for full documentation
		// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
		// The logger is an optional parameter. Defaults to the noop logger.
		//
		// Deprecated: prefer ReplayWorkflowHistoryFromJSON
		ReplayWorkflowHistoryFromJSONFile(logger *zap.Logger, jsonfileName string) error

		// ReplayPartialWorkflowHistoryFromJSONFile executes a single decision task for the json history file upto provided
		// lastEventID(inclusive), downloaded from the cli.
		// To download the history file: cadence workflow showid <workflow_id> -of <output_filename>
		// See https://github.com/uber/cadence/blob/master/tools/cli/README.md for full documentation
		// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
		// The logger is an optional parameter. Defaults to the noop logger.
		//
		// Deprecated: prefer ReplayPartialWorkflowHistoryFromJSON
		ReplayPartialWorkflowHistoryFromJSONFile(logger *zap.Logger, jsonfileName string, lastEventID int64) error

		// ReplayWorkflowExecution loads a workflow execution history from the Cadence service and executes a single decision task for it.
		// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
		// The logger is the only optional parameter. Defaults to the noop logger.
		ReplayWorkflowExecution(ctx context.Context, service workflowserviceclient.Interface, logger *zap.Logger, domain string, execution workflow.Execution) error

		// ReplayWorkflowHistoryFromJSON executes a single decision task for the json history file downloaded from the cli.
		// To download the history file:
		//  cadence workflow showid <workflow_id> -of <output_filename>
		// See https://github.com/uber/cadence/blob/master/tools/cli/README.md for full documentation
		// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
		// The logger is an optional parameter. Defaults to the noop logger.
		ReplayWorkflowHistoryFromJSON(logger *zap.Logger, reader io.Reader) error

		// ReplayPartialWorkflowHistoryFromJSON executes a single decision task for the json history file upto provided
		// lastEventID(inclusive), downloaded from the cli.
		// To download the history file:
		//   cadence workflow showid <workflow_id> -of <output_filename>
		// See https://github.com/uber/cadence/blob/master/tools/cli/README.md for full documentation
		// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
		// The logger is an optional parameter. Defaults to the noop logger.
		ReplayPartialWorkflowHistoryFromJSON(logger *zap.Logger, reader io.Reader, lastEventID int64) error
	}

	// WorkflowShadower retrieves and replays workflow history from Cadence service to determine if there's any nondeterministic changes in the workflow definition
	WorkflowShadower interface {
		worker.WorkflowRegistry

		Run() error
	}

	// ShadowOptions is used to configure a WorkflowShadower.
	ShadowOptions = internal.ShadowOptions
	// ShadowMode is an enum for configuring if shadowing should continue after all workflows matches the WorkflowQuery have been replayed.
	ShadowMode = internal.ShadowMode
	// TimeFilter represents a time range through the min and max timestamp
	TimeFilter = internal.TimeFilter
	// ShadowExitCondition configures when the workflow shadower should exit.
	// If not specified shadower will exit after replaying all workflows satisfying the visibility query.
	ShadowExitCondition = internal.ShadowExitCondition

	// ReplayOptions is used to configure the replay decision task worker.
	ReplayOptions = internal.ReplayOptions
)

const (
	// ShadowModeNormal is the default mode for workflow shadowing.
	// Shadowing will complete after all workflows matches WorkflowQuery have been replayed.
	ShadowModeNormal = internal.ShadowModeNormal
	// ShadowModeContinuous mode will start a new round of shadowing
	// after all workflows matches WorkflowQuery have been replayed.
	// There will be a 5 min wait period between each round,
	// currently this wait period is not configurable.
	// Shadowing will complete only when ExitCondition is met.
	// ExitCondition must be specified when using this mode
	ShadowModeContinuous = internal.ShadowModeContinuous
)

// NewWorkflowReplayer creates a WorkflowReplayer instance.
func NewWorkflowReplayer() WorkflowReplayer {
	return internal.NewWorkflowReplayer()
}

// NewWorkflowReplayerWithOptions creates an instance of the WorkflowReplayer
// with provided replay worker options
func NewWorkflowReplayerWithOptions(
	options ReplayOptions,
) WorkflowReplayer {
	return internal.NewWorkflowReplayerWithOptions(options)
}

// NewWorkflowShadower creates a WorkflowShadower instance.
func NewWorkflowShadower(
	service workflowserviceclient.Interface,
	domain string,
	shadowOptions ShadowOptions,
	replayOptions ReplayOptions,
	logger *zap.Logger,
) (WorkflowShadower, error) {
	return internal.NewWorkflowShadower(service, domain, shadowOptions, replayOptions, logger)
}
