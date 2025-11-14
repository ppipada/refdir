package defaultdirs

func TestRefDownToStringConst() {
	_ = StringConstAtEnd // want "const reference StringConstAtEnd is before definition"
}

const StringConstAtEnd = "end"
