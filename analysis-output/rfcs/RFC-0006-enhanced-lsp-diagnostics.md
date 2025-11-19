# RFC-0006: Enhanced LSP Diagnostics

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 2-3 weeks
**Category:** Strategic

---

## Summary

Enhance the Language Server Protocol (LSP) implementation with real-time diagnostics, better code intelligence, and performance optimizations for improved IDE experience.

---

## Motivation

The current LSP implementation provides basic functionality but has room for improvement:

1. **Delayed diagnostics** - Lint errors shown only after save
2. **Limited completion** - Missing context-aware suggestions
3. **No quick fixes** - Can't auto-fix lint violations
4. **Performance** - Full recompilation on each change

Better IDE integration would:
- Reduce context switching (no terminal needed)
- Catch errors earlier in development
- Improve developer productivity
- Match expectations from other language tools

---

## Detailed Design

### 1. Real-Time Diagnostics

Provide diagnostics as user types:

```go
// private/buf/buflsp/diagnostics.go
package buflsp

func (s *server) TextDocumentDidChange(
    ctx context.Context,
    params *protocol.DidChangeTextDocumentParams,
) error {
    uri := params.TextDocument.URI

    // Debounce rapid changes
    s.debounceDiagnostics(uri, 300*time.Millisecond, func() {
        // Incremental compilation
        image, err := s.compileIncremental(ctx, uri)
        if err != nil {
            s.publishCompileErrors(ctx, uri, err)
            return
        }

        // Run lint on changed file
        annotations, err := s.lintFile(ctx, image, uri)
        if err != nil {
            return
        }

        // Publish diagnostics
        s.publishDiagnostics(ctx, uri, annotations)
    })

    return nil
}
```

### 2. Code Actions (Quick Fixes)

Auto-fix common lint violations:

```go
// private/buf/buflsp/codeaction.go
package buflsp

var codeActions = map[string]CodeActionProvider{
    "FIELD_LOWER_SNAKE_CASE": &fieldLowerSnakeCaseAction{},
    "ENUM_PASCAL_CASE":       &enumPascalCaseAction{},
    "SERVICE_SUFFIX":         &serviceSuffixAction{},
    "COMMENT_FIELD":          &addCommentAction{},
}

func (s *server) TextDocumentCodeAction(
    ctx context.Context,
    params *protocol.CodeActionParams,
) ([]protocol.CodeAction, error) {
    var actions []protocol.CodeAction

    for _, diagnostic := range params.Context.Diagnostics {
        if provider, ok := codeActions[diagnostic.Code]; ok {
            action, err := provider.Provide(ctx, params.TextDocument, diagnostic)
            if err != nil {
                continue
            }
            actions = append(actions, action)
        }
    }

    return actions, nil
}

// Example: Fix field naming
type fieldLowerSnakeCaseAction struct{}

func (a *fieldLowerSnakeCaseAction) Provide(
    ctx context.Context,
    doc protocol.TextDocumentIdentifier,
    diag protocol.Diagnostic,
) (protocol.CodeAction, error) {
    // Extract field name from diagnostic
    oldName := extractFieldName(diag.Message)
    newName := toLowerSnakeCase(oldName)

    return protocol.CodeAction{
        Title: fmt.Sprintf("Rename to '%s'", newName),
        Kind:  protocol.QuickFix,
        Edit: &protocol.WorkspaceEdit{
            Changes: map[string][]protocol.TextEdit{
                doc.URI: {
                    {
                        Range:   diag.Range,
                        NewText: newName,
                    },
                },
            },
        },
        IsPreferred: true,
    }, nil
}
```

### 3. Enhanced Completions

Context-aware completions:

```go
// private/buf/buflsp/completion.go
func (s *server) TextDocumentCompletion(
    ctx context.Context,
    params *protocol.CompletionParams,
) (*protocol.CompletionList, error) {
    context := s.analyzeContext(params)

    var items []protocol.CompletionItem

    switch context.Type {
    case ContextTypeImport:
        // Suggest available modules from BSR
        items = s.getModuleCompletions(ctx)

    case ContextTypeFieldType:
        // Suggest message types, well-known types
        items = s.getTypeCompletions(ctx, context)

    case ContextTypeOption:
        // Suggest available options for context
        items = s.getOptionCompletions(ctx, context)

    case ContextTypeRPCInput, ContextTypeRPCOutput:
        // Suggest request/response message types
        items = s.getMessageCompletions(ctx, context)

    case ContextTypeEnum:
        // Suggest enum naming patterns
        items = s.getEnumCompletions(ctx, context)
    }

    return &protocol.CompletionList{
        IsIncomplete: false,
        Items:        items,
    }, nil
}
```

### 4. Incremental Compilation

Only recompile changed files:

```go
// private/buf/buflsp/incremental.go
type IncrementalCompiler struct {
    cache      map[string]*CompiledFile
    depGraph   *DependencyGraph
    mu         sync.RWMutex
}

func (c *IncrementalCompiler) Compile(
    ctx context.Context,
    changedURI string,
    content string,
) (*bufimage.Image, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Parse changed file
    parsed, err := c.parseFile(changedURI, content)
    if err != nil {
        return nil, err
    }

    // Find affected files (dependents)
    affected := c.depGraph.GetDependents(changedURI)

    // Recompile only affected
    for _, uri := range append([]string{changedURI}, affected...) {
        compiled, err := c.compileFile(ctx, uri)
        if err != nil {
            return nil, err
        }
        c.cache[uri] = compiled
    }

    // Merge into image
    return c.buildImage()
}
```

### 5. Semantic Tokens

Provide rich syntax highlighting:

```go
// private/buf/buflsp/semantic.go
func (s *server) TextDocumentSemanticTokensFull(
    ctx context.Context,
    params *protocol.SemanticTokensParams,
) (*protocol.SemanticTokens, error) {
    file, err := s.getFile(params.TextDocument.URI)
    if err != nil {
        return nil, err
    }

    var tokens []uint32

    // Walk AST and generate tokens
    for _, node := range file.AST.Nodes {
        switch n := node.(type) {
        case *MessageNode:
            tokens = append(tokens, tokenize(n.Name, SemanticTypeClass)...)
        case *FieldNode:
            tokens = append(tokens, tokenize(n.Name, SemanticTypeProperty)...)
        case *ServiceNode:
            tokens = append(tokens, tokenize(n.Name, SemanticTypeInterface)...)
        case *RPCNode:
            tokens = append(tokens, tokenize(n.Name, SemanticTypeMethod)...)
        }
    }

    return &protocol.SemanticTokens{Data: tokens}, nil
}
```

### 6. Hover Information

Rich hover documentation:

```go
// private/buf/buflsp/hover.go
func (s *server) TextDocumentHover(
    ctx context.Context,
    params *protocol.HoverParams,
) (*protocol.Hover, error) {
    symbol, err := s.findSymbol(params.TextDocument.URI, params.Position)
    if err != nil {
        return nil, nil
    }

    var content string

    switch sym := symbol.(type) {
    case *MessageSymbol:
        content = fmt.Sprintf(
            "```protobuf\nmessage %s\n```\n\n%s\n\n**Fields:** %d\n**Used by:** %d messages",
            sym.Name,
            sym.Comment,
            len(sym.Fields),
            len(sym.UsedBy),
        )

    case *FieldSymbol:
        content = fmt.Sprintf(
            "```protobuf\n%s %s = %d;\n```\n\n%s",
            sym.Type,
            sym.Name,
            sym.Number,
            sym.Comment,
        )
    }

    return &protocol.Hover{
        Contents: protocol.MarkupContent{
            Kind:  protocol.Markdown,
            Value: content,
        },
    }, nil
}
```

---

## Implementation Plan

### Phase 1: Incremental Compilation (1 week)
1. Implement dependency graph
2. Add incremental compiler
3. Optimize for latency

### Phase 2: Real-Time Diagnostics (0.5 week)
1. Add debounced diagnostics
2. Implement diagnostic publishing
3. Handle compile errors

### Phase 3: Code Actions (1 week)
1. Implement code action framework
2. Add quick fixes for common rules
3. Test with IDEs

### Phase 4: Enhanced Completions (0.5 week)
1. Context analysis
2. Module/type completions
3. BSR integration

---

## Backwards Compatibility

**Impact:** None

All enhancements are additive to existing LSP capabilities.

---

## Success Criteria

- [ ] Diagnostics appear within 500ms of typing
- [ ] Quick fixes for top 10 lint rules
- [ ] Context-aware completions
- [ ] >98% diagnostic accuracy
- [ ] Memory usage <200MB for large projects

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Developer Experience Team
