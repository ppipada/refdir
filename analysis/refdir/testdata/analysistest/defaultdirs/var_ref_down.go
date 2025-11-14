package defaultdirs

func TestRefDownToStringVar() {
	_ = StringVarAtEnd // want "var reference StringVarAtEnd is before definition"
}

var StringVarAtEnd string
