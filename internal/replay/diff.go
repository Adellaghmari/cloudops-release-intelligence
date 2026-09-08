package replay

import "fmt"

type Kind string

const (
	Added     Kind = "added"
	Removed   Kind = "removed"
	Changed   Kind = "changed"
	Unchanged Kind = "unchanged"
)

type Field struct {
	Path string
	Kind Kind
	A    string
	B    string
}

func CompareStrings(path, a, b string) Field {
	k := Unchanged
	switch {
	case a == "" && b != "":
		k = Added
	case a != "" && b == "":
		k = Removed
	case a != b:
		k = Changed
	}
	return Field{Path: path, Kind: k, A: a, B: b}
}

func CompareInts(path string, a, b int) Field {
	return CompareStrings(path, fmt.Sprintf("%d", a), fmt.Sprintf("%d", b))
}

func CompareBools(path string, a, b bool) Field {
	return CompareStrings(path, fmt.Sprintf("%t", a), fmt.Sprintf("%t", b))
}
