package main

type Ci struct{}

func (m *Ci) Spectral() *Container {
	return dag.Spectral().
		WithSource(dag.Host().Directory("./testdata/spectral")).
		Lint("openapi.json")
}
