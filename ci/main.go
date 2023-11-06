package main

type Ci struct{}

func (m *Ci) Bats() *Container {
	return dag.Bats().
		WithSource(dag.Host().Directory("./testdata/bats")).
		Run([]string{"test.bats"})
}

func (m *Ci) Golang() *Container {
	return dag.Golang().
		Exec([]string{"go", "version"})
}

func (m *Ci) Kafka() *Container {
	kafka := dag.Kafka()

	return kafka.Container().
		WithServiceBinding("kafka", kafka.Service()).
		WithExec([]string{"kafka-topics.sh", "--list", "--bootstrap-server", "kafka:9092"})
}

func (m *Ci) Spectral() *Container {
	return dag.Spectral().
		WithSource(dag.Host().Directory("./testdata/spectral")).
		Lint("openapi.json")
}
