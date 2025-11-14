package defaultdirs

func TestRefDownToConstraintType[T TestRefDownConstraintType]() {} // want "type reference TestRefDownConstraintType is before definition"

type TestRefDownConstraintType interface {
	DummyMethod()
}
