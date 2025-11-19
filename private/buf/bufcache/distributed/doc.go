// Copyright 2020-2025 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package distributed provides a distributed build cache implementation
// for buf that allows teams to share compilation results across team members
// and CI systems.
//
// The distributed cache supports multiple backends including S3, GCS, and HTTP.
// Cache keys are generated based on content hashes of input files, buf version,
// and compilation options to ensure deterministic cache hits.
package distributed
