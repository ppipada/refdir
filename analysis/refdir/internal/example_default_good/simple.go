package example

import "log"

func MakePancake() {
	readRecipe()

	visitStore()

	mixFlour()

	FindOven().
		WarmUp().
		WarmUp().
		Bake()

	Enjoy("pancake")
}

func visitStore() {
	readRecipe()
	buyFlour()
}

func mixFlour() {
	log.Println("mixing flour~~")
}

func readRecipe() {
	log.Println("reading recepie")
}

func buyFlour() {
	log.Println("bought flour")
}

type Oven struct {
	temperature float32
}

func FindOven() *Oven { return &Oven{} }

func (s *Oven) WarmUp() *Oven {
	s.temperature += 42
	return s
}

func (s *Oven) Bake() {
	log.Println("done!")
}

func Enjoy(food string) {
	log.Println(food + " is so good")
}
