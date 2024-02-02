# Helm docs

**A tool for automatically generating markdown documentation for helm charts.**

## Usage

```go
readme := dag.HelmDocs().Generate(dag.CurrentModule().Source().Directory("path/to/chart"))
```
