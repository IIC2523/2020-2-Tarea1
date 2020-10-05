package main

import (
	"fmt"
	"strings"
	"sync"
)

var key1 = "programar"
var key2 = "distribuir"


func mapper(inMapper <-chan string, outMapper chan<- map[string]int) {
	count := map[string]int{key1: 0, key2: 0}
	for word := range inMapper {
		if strings.Contains(word, key1) {
			count[key1] = count[key1] + 1
		}
		if strings.Contains(word, key2) {
			count[key2] = count[key2] + 1
		}
	}
	outMapper <- count
	close(outMapper)
}


func reducer(inReducer <-chan int, outReducer chan<- float32) {
	sum, count := 0, 0
	for n := range inReducer {
		sum += n
		count++
	}
	outReducer <- float32(sum) / float32(count)
	close(outReducer)
}

func inputReader(output [3]chan<- string) {
	input := [][]string{
		{
			"Este es una linea del primer archivo de programar",
			"busca explicar la materia de sistemas distribuir",
			"para esto se usara un test", "con texto falso",
			"distriwqesad"},
		{"distribuir", "segundo", "tercero", "mucho texto", "todo bien", "todo mal"},
		{"no", "hay ", "palabras", " claves"},
	}

	for i := range output {
		go func(ch chan<- string, word []string) {
			for _, w := range word {
				ch <- w
			}
			close(ch)
		}(output[i], input[i])
	}
}

func shuffler(inShuffler []<-chan map[string]int, outShuffler [2]chan<- int) {
	var wg sync.WaitGroup
	wg.Add(len(inShuffler))
	for _, ch := range inShuffler {
		go func(c <-chan map[string]int) {
			for m := range c {
				nc, ok := m[key1]
				if ok {
					outShuffler[0] <- nc
				}
				vc, ok := m[key2]
				if ok {
					outShuffler[1] <- vc
				}
			}
			wg.Done()
		}(ch)
	}
	go func() {
		wg.Wait()
		close(outShuffler[0])
		close(outShuffler[1])
	}()
}


func outputWriter(in []<-chan float32) {
	var wg sync.WaitGroup
	wg.Add(len(in))

	name := []string{key1, key2}
	for i := 0; i < len(in); i++ {
		go func(n int, c <-chan float32) {
			for avg := range c {
				fmt.Printf("El promedio de veces que la palabra \"%s\" aparece en los textos es: %f\n", name[n], avg)
			}
			wg.Done()
		}(i, in[i])
	}
	wg.Wait()
}

func main() {

	// Definios el tamanio
	size := 100
	// Definimos los textos que iran en el input
	text1 := make(chan string, size)
	text2 := make(chan string, size)
	text3 := make(chan string, size)

	go inputReader([3]chan<- string{text1, text2, text3})
	// Definimos los maps que se obtendran del mapper
	map1 := make(chan map[string]int, size)
	map2 := make(chan map[string]int, size)
	map3 := make(chan map[string]int, size)

	go mapper(text1, map1)
	go mapper(text2, map2)
	go mapper(text3, map3)

	// Definimos los reduce que se obtendran en el reducer
	reduce1 := make(chan int, size)
	reduce2 := make(chan int, size)

	// Ejecutamos el shuffler para juntar los resultados de una palabra en un arreglo distinto
	go shuffler([]<-chan map[string]int{map1, map2, map3}, [2]chan<- int{reduce1, reduce2})

	// Definimos el promedio de cada contador de palabras en los textos
	avg1 := make(chan float32, size)
	avg2 := make(chan float32, size)

	// Ejecutamos los procesos de reducer con cada conjunto
	go reducer(reduce1, avg1)
	go reducer(reduce2, avg2)

	// Escribimos el resultado final
	outputWriter([]<-chan float32{avg1, avg2})
}

// Ejemplo de código realizado a partir del ejemplo del repositorio github.com/appliedgo/mapreduce.
// En casos de querer hacer una ejecución de Select por ejemplo, se debe hacer el mismo proceso de division de las ejecuciones correspondientes
// Para obtener una relacion necesaria
