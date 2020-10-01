package main

import (
	"bufio"   // Leer STDIN
	"flag"    // Leer flags
	"fmt"     // Print
	"log"     // Escribir Logs
	"os"      // Leer sistema de archivos
	"runtime" // Imprimir # de GoRouitnes
	"strings" // Usar Replace()

	"image"       // Clase Image
	"image/color" // Clase Color
	"image/png"   // Formato .png

	"strconv" // Convertir de string a int
)

func ExtValue(img image.Image, x, y int) (r, g, b uint32) {
	var boundX, boundY, maxX, maxY int = x, y, img.Bounds().Max.X, img.Bounds().Max.Y

	if x < 0 {
		boundX = 0
	}
	if y < 0 {
		boundY = 0
	}
	if x >= maxX {
		boundX = maxX - 1
	}
	if y >= maxY {
		boundY = maxY - 1
	}

	r, g, b, _ = img.At(boundX, boundY).RGBA()
	return r, g, b
}

func BlurPixel(img image.Image, mask Mask, x, y int) color.RGBA {
	var sumR, sumG, sumB uint32 = 0, 0, 0

	minW := -int((mask.width - 1) / 2)
	maxW := int((mask.width - 1) / 2)

	minH := -int((mask.height - 1) / 2)
	maxH := int((mask.height - 1) / 2)

	total := uint32(mask.width * mask.height)
	_, _, _, alpha := img.At(x, y).RGBA()

	for i := minW; i <= maxW; i++ {
		for j := minH; j <= maxH; j++ {
			r, g, b := ExtValue(img, x+i, y+j)
			sumR += r
			sumG += g
			sumB += b
		}
	}

	finalR, finalG, finalB := sumR/total/257, sumG/total/257, sumB/total/257
	return color.RGBA{uint8(finalR), uint8(finalG), uint8(finalB), uint8(alpha)}
}

func WritePixel(img image.Image, newImg *image.RGBA, pix chan Pixel, mask Mask) {
	fmt.Println("Number of GoRoutines: ", runtime.NumGoroutine())
	for p := range pix {
		rgba := BlurPixel(img, mask, p.x, p.y)
		newImg.Set(p.x, p.y, rgba)
	}
}

type Mask struct {
	width  int
	height int
}

type Pixel struct {
	x int
	y int
}

func main() {
	numThreads := flag.Int("threads", 1, "Number of threads to use")
	flag.Parse()
	fmt.Println("Threads:", *numThreads)

	imagePath := os.Args[3]

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter width:")
	strW, _ := reader.ReadString('\n')
	strW = strings.Replace(strW, "\n", "", -1)
	w, _ := strconv.Atoi(strW)

	fmt.Println("Enter height:")
	strH, _ := reader.ReadString('\n')
	strH = strings.Replace(strH, "\n", "", -1)
	h, _ := strconv.Atoi(strH)

	mask := Mask{w, h}

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	imgFile, _ := os.Open(imagePath)
	img, _, _ := image.Decode(imgFile)

	newImage := image.NewRGBA(img.Bounds())
	
	pix := make(chan Pixel)
	for i := 0; i < *numThreads; i++ {
		go WritePixel(img, newImage, pix, mask)
	}

	var maxX, maxY int = img.Bounds().Max.X, img.Bounds().Max.Y
	for i := 0; i < maxX; i++ {
		for j := 0; j < maxY; j++ {
			pix <- Pixel{i, j}
		}
	}
	close(pix)
	fmt.Println("Channel closed")

	newFile, err := os.Create("blur.png")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Saving File")
	err = png.Encode(newFile, newImage)
	if err != nil {
		log.Fatal(err)
	}
}