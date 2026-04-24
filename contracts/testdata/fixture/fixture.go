// Package fixture is a tiny package used to test the extractor.
package fixture

type Thing struct {
	Name string
	Size int
}

type Doer interface {
	Do() error
}

func ApplyThing(t *Thing) *Thing {
	return t
}

func privateHelper() {}
