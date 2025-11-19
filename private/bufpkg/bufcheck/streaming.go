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

package bufcheck

import (
	"context"
	"log/slog"

	"github.com/bufbuild/buf/private/bufpkg/bufanalysis"
	"github.com/bufbuild/buf/private/bufpkg/bufconfig"
	"github.com/bufbuild/buf/private/bufpkg/bufimage"
)

// StreamingLintClient provides streaming lint capabilities.
type StreamingLintClient struct {
	client Client
	logger *slog.Logger
}

// NewStreamingLintClient creates a new StreamingLintClient.
func NewStreamingLintClient(client Client, logger *slog.Logger) *StreamingLintClient {
	return &StreamingLintClient{
		client: client,
		logger: logger,
	}
}

// LintStream lints files as they become available and streams results.
// The callback is invoked for each file annotation as it is discovered.
// If the callback returns an error, streaming stops and the error is returned.
func (c *StreamingLintClient) LintStream(
	ctx context.Context,
	config bufconfig.LintConfig,
	files <-chan bufimage.ImageFile,
	callback func(bufanalysis.FileAnnotation) error,
	options ...LintOption,
) error {
	// Collect files as they stream in
	var imageFiles []bufimage.ImageFile
	for file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			imageFiles = append(imageFiles, file)
		}
	}

	// If no files were received, return early
	if len(imageFiles) == 0 {
		return nil
	}

	// Create image from collected files
	image, err := bufimage.NewImage(imageFiles)
	if err != nil {
		return err
	}

	// Run lint and handle results
	err = c.client.Lint(ctx, config, image, options...)
	if err != nil {
		// Check if error is a FileAnnotationSet
		if annotationSet, ok := err.(bufanalysis.FileAnnotationSet); ok {
			// Stream each annotation through the callback
			for _, annotation := range annotationSet.FileAnnotations() {
				if callbackErr := callback(annotation); callbackErr != nil {
					return callbackErr
				}
			}
			return err // Return the original annotation set error
		}
		return err
	}

	return nil
}

// LintImageStream lints an image and streams results through a channel.
// Returns a channel of file annotations and an error channel.
func (c *StreamingLintClient) LintImageStream(
	ctx context.Context,
	config bufconfig.LintConfig,
	image bufimage.Image,
	options ...LintOption,
) (<-chan bufanalysis.FileAnnotation, <-chan error) {
	results := make(chan bufanalysis.FileAnnotation, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(results)
		defer close(errChan)

		err := c.client.Lint(ctx, config, image, options...)
		if err != nil {
			// Check if error is a FileAnnotationSet
			if annotationSet, ok := err.(bufanalysis.FileAnnotationSet); ok {
				// Stream each annotation
				for _, annotation := range annotationSet.FileAnnotations() {
					select {
					case results <- annotation:
					case <-ctx.Done():
						errChan <- ctx.Err()
						return
					}
				}
				errChan <- err
				return
			}
			errChan <- err
			return
		}
	}()

	return results, errChan
}

// LintStreamProgress provides progress information during streaming lint.
type LintStreamProgress struct {
	// TotalFiles is the total number of files to lint.
	TotalFiles int
	// ProcessedFiles is the number of files processed so far.
	ProcessedFiles int
	// AnnotationCount is the number of annotations found so far.
	AnnotationCount int
	// CurrentFile is the path of the file currently being processed.
	CurrentFile string
}

// LintWithProgress lints an image and reports progress through a callback.
func (c *StreamingLintClient) LintWithProgress(
	ctx context.Context,
	config bufconfig.LintConfig,
	image bufimage.Image,
	progress func(LintStreamProgress),
	options ...LintOption,
) error {
	totalFiles := len(image.Files())

	// Report initial progress
	if progress != nil {
		progress(LintStreamProgress{
			TotalFiles:      totalFiles,
			ProcessedFiles:  0,
			AnnotationCount: 0,
		})
	}

	// Run lint
	err := c.client.Lint(ctx, config, image, options...)

	// Report completion
	if progress != nil {
		annotationCount := 0
		if annotationSet, ok := err.(bufanalysis.FileAnnotationSet); ok {
			annotationCount = len(annotationSet.FileAnnotations())
		}
		progress(LintStreamProgress{
			TotalFiles:      totalFiles,
			ProcessedFiles:  totalFiles,
			AnnotationCount: annotationCount,
		})
	}

	return err
}

// LintByFile lints each file individually and streams results.
// This provides per-file streaming but may be slower than batch linting.
func (c *StreamingLintClient) LintByFile(
	ctx context.Context,
	config bufconfig.LintConfig,
	image bufimage.Image,
	callback func(file string, annotations []bufanalysis.FileAnnotation) error,
	options ...LintOption,
) error {
	// Run full lint first
	err := c.client.Lint(ctx, config, image, options...)

	// Group annotations by file
	fileAnnotations := make(map[string][]bufanalysis.FileAnnotation)

	if annotationSet, ok := err.(bufanalysis.FileAnnotationSet); ok {
		for _, annotation := range annotationSet.FileAnnotations() {
			filePath := ""
			if annotation.FileInfo() != nil {
				filePath = annotation.FileInfo().Path()
			}
			fileAnnotations[filePath] = append(fileAnnotations[filePath], annotation)
		}
	}

	// Stream results per file
	for _, file := range image.Files() {
		if file.IsImport() {
			continue // Skip imports
		}

		path := file.Path()
		annotations := fileAnnotations[path]

		if callbackErr := callback(path, annotations); callbackErr != nil {
			return callbackErr
		}
	}

	return err
}

// StreamingBreakingClient provides streaming breaking change detection.
type StreamingBreakingClient struct {
	client Client
	logger *slog.Logger
}

// NewStreamingBreakingClient creates a new StreamingBreakingClient.
func NewStreamingBreakingClient(client Client, logger *slog.Logger) *StreamingBreakingClient {
	return &StreamingBreakingClient{
		client: client,
		logger: logger,
	}
}

// BreakingStream checks for breaking changes and streams results.
func (c *StreamingBreakingClient) BreakingStream(
	ctx context.Context,
	config bufconfig.BreakingConfig,
	image bufimage.Image,
	againstImage bufimage.Image,
	callback func(bufanalysis.FileAnnotation) error,
	options ...BreakingOption,
) error {
	err := c.client.Breaking(ctx, config, image, againstImage, options...)
	if err != nil {
		if annotationSet, ok := err.(bufanalysis.FileAnnotationSet); ok {
			for _, annotation := range annotationSet.FileAnnotations() {
				if callbackErr := callback(annotation); callbackErr != nil {
					return callbackErr
				}
			}
		}
		return err
	}
	return nil
}

// BreakingImageStream checks for breaking changes and streams results through channels.
func (c *StreamingBreakingClient) BreakingImageStream(
	ctx context.Context,
	config bufconfig.BreakingConfig,
	image bufimage.Image,
	againstImage bufimage.Image,
	options ...BreakingOption,
) (<-chan bufanalysis.FileAnnotation, <-chan error) {
	results := make(chan bufanalysis.FileAnnotation, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(results)
		defer close(errChan)

		err := c.client.Breaking(ctx, config, image, againstImage, options...)
		if err != nil {
			if annotationSet, ok := err.(bufanalysis.FileAnnotationSet); ok {
				for _, annotation := range annotationSet.FileAnnotations() {
					select {
					case results <- annotation:
					case <-ctx.Done():
						errChan <- ctx.Err()
						return
					}
				}
				errChan <- err
				return
			}
			errChan <- err
			return
		}
	}()

	return results, errChan
}
