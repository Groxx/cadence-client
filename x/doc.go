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
Package x holds all e"x"perimental APIs.

Use of anything in this package is strictly AT YOUR OWN RISK: we intend to make breaking changes,
and we will NOT be making a new major client version whenever we do so.  Possibly not even a minor version.

In particular: if your code breaks because you are depending on code in here and we change something, we expect
you to check the git history / PRs involved and figure out how to handle it on your own.
You are welcome to ask questions and suggest changes of course (or submit a PR), but this is strictly "power-user"
territory and support will generally be minimal.

Broadly, APIs in here are divided into three categories:

 1. New APIs being developed that are not yet ready / not yet API-stable.
    These allow early adopters to try things, with the caveat that their code will likely need to change
    to handle changes as part of normal development, and when the APIs are fully stabilized and moved out of /x/.
 2. Existing pre-v2 APIs which are still useful but are not stable or reliable enough to consider "recommended for use".
    These are likely to be "stable" in practice, as a replacement will probably be built separately,
    but may be removed at any time if we decide that the maintenance cost is too high.
 3. APIs which are inherently unstable *by design*.  Among others, these may be "leakages" of internal details
    for a variety of uses, e.g. introspection into a worker's current caches.  These are likely to change
    frequently, but are also likely to be easy to adjust to the changes as long as usage is kept simple:
    we will be doing the same thing for our internal wrappers / debug tooling / etc.

For further details, see individual sub-packages.
*/
package x
