package defaultdirs

type TestRefUpConstraintType interface {
	DummyMethod()
}

func TestRefUpToConstraintType[T TestRefUpConstraintType]() {}
