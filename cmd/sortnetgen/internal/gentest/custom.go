package gentest

type Custom struct {
	Foo int
}

func CustomCASLess(a *Custom, b *Custom) {
	if a.Foo < b.Foo {
		*a, *b = *b, *a
	}
}

func CustomCASGreater(a *Custom, b *Custom) {
	if a.Foo > b.Foo {
		*a, *b = *b, *a
	}
}
