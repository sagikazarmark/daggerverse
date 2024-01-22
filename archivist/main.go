package main

// Archivist provides methods to create and extract archives.
type Archivist struct{}

func arc() *Arc {
	return dag.Arc(ArcOpts{
		Version: "3.5.0", // pin version
	})
}

// Create and extract ".tar" archives.
func (m *Archivist) Tar() *Tar {
	return &Tar{}
}

// Create and extract ".tar" archives.
type Tar struct{}

func (m *Tar) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).Tar()
}

// Create and extract ".tar.br" (and ".tbr") archives.
func (m *Archivist) TarBr() *TarBr {
	return &TarBr{}
}

// Create and extract ".tar.br" (and ".tbr") archives.
type TarBr struct{}

func (m *TarBr) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).TarBr()
}

// Create and extract ".tar.bz2" (and ".tbz2") archives.
func (m *Archivist) TarBz2() *TarBz2 {
	return &TarBz2{}
}

// Create and extract ".tar.bz2" (and ".tbz2") archives.
type TarBz2 struct{}

func (m *TarBz2) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).TarBz2()
}

// Create and extract ".tar.gz" (and ".tgz") archives.
func (m *Archivist) TarGz() *TarGz {
	return &TarGz{}
}

// Create and extract ".tar.gz" (and ".tgz") archives.
type TarGz struct{}

func (m *TarGz) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).TarGz()
}

// Create and extract ".tar.lz4" (and ".tlz4") archives.
func (m *Archivist) TarLz4() *TarLz4 {
	return &TarLz4{}
}

// Create and extract ".tar.lz4" (and ".tlz4") archives.
type TarLz4 struct{}

func (m *TarLz4) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).TarLz4()
}

// Create and extract ".tar.sz" (and ".tsz") archives.
func (m *Archivist) TarSz() *TarSz {
	return &TarSz{}
}

// Create and extract ".tar.sz" (and ".tsz") archives.
type TarSz struct{}

func (m *TarSz) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).TarSz()
}

// Create and extract ".tar.xz" (and ".txz") archives.
func (m *Archivist) TarXz() *TarXz {
	return &TarXz{}
}

// Create and extract ".tar.xz" (and ".txz") archives.
type TarXz struct{}

func (m *TarXz) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).TarXz()
}

// Create and extract ".tar.zst" archives.
func (m *Archivist) TarZst() *TarZst {
	return &TarZst{}
}

// Create and extract ".tar.zst" archives.
type TarZst struct{}

func (m *TarZst) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).TarZst()
}

// Create and extract ".zip" archives.
func (m *Archivist) Zip() *Zip {
	return &Zip{}
}

// Create and extract ".zip" archives.
type Zip struct{}

func (m *Zip) Archive(name string, source *Directory) *File {
	return arc().ArchiveDirectory(name, source).Zip()
}
