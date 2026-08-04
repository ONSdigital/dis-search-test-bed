package testset

import (
	"embed"
	"io/fs"
)

var (
	//go:embed documents/*.json
	documentFixturesFS embed.FS
	//go:embed judgements/*.json
	judgementFixturesFS embed.FS
	//go:embed terms/*.json
	termFixturesFS embed.FS
)

// Directory constants for the testset items.
const (
	DocumentFixturesDir  = "documents"
	JudgementFixturesDir = "judgements"
	TermFixturesDir      = "terms"
)

// DocumentFixturesFS returns the embedded, read-only filesystem holding the
// document fixtures under [DocumentFixturesDir].
func DocumentFixturesFS() fs.FS {
	return documentFixturesFS
}

// JudgementFixturesFS returns the embedded, read-only filesystem holding the
// judgement fixtures under [JudgementFixturesDir].
func JudgementFixturesFS() fs.FS {
	return judgementFixturesFS
}

// TermFixturesFS returns the embedded, read-only filesystem holding the
// term fixtures under [TermFixturesDir].
func TermFixturesFS() fs.FS {
	return termFixturesFS
}
