package main

import (
	"math/rand"
)

// viva chatGPT  :-D
var nomes = []string{
	"João",
	"Maria",
	"José",
	"Ana",
	"Carlos",
	"Juliana",
	"Pedro",
	"Paula",
	"Lucas",
	"Gabriela",
	"Miguel",
	"Rafaela",
	"André",
	"Laura",
	"Fernando",
	"Carolina",
	"Rodrigo",
	"Daniela",
	"Ricardo",
	"Sandra",
	"Gustavo",
	"Mariana",
	"Diego",
	"Camila",
	"Alexandre",
	"Patrícia",
	"Daniel",
	"Vanessa",
	"Luiz",
	"Beatriz",
	"Marcelo",
	"Natália",
	"Guilherme",
	"Isabela",
	"Fábio",
	"Tatiane",
	"Arthur",
	"Eduarda",
	"Leonardo",
	"Ana Paula",
	"Vinícius",
	"Aline",
	"Renato",
	"Amanda",
	"Roberto",
	"Carla",
	"Vitor",
	"Priscila",
}

func getRandomName() string {
	return nomes[rand.Intn(len(nomes))]
}
