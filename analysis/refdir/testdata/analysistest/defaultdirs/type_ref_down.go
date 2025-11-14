package defaultdirs

func TestTypeRefDown() {
	_ = TestTypeRefDownType{} // want "type reference TestTypeRefDownType is before definition"
}

type TestTypeRefDownType struct{}
