# Graph Builder

A pattern for building software with graph-based dependencies. Nodes handle business logic and declare their dependencies explicitly.

This pattern is an alternative to dependency injection without the magic and with more idiomatic go.

## Why this approach?

DI frameworks in Go often feel awkward. They rely on compile-time code gen or runtime reflection, which adds complexity and obscures what's actually happening.

This pattern uses init functions and well-defined package boundaries instead. Each node knows its dependencies, and the graph assembles itself at startup. It's particularly useful when multiple teams work in a single system and need clear ownership boundaries.

## Examples

**[basic/](./basic/)** — A hello world dependency graph. Runs the entire graph once.

```bash
pushd basic; go run .; popd
```

**[server/](./server/)** — Builds on basic with a web server that constructs minimal subgraphs on-demand. Useful for HTTP services where different endpoints need different slices of the dependency tree.

```bash
pushd server; go run .; popd
```
