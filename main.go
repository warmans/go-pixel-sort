package main

import (
	"os"
	"log"
	"flag"
	"image"
	"image/png"
	"path"
	"strings"
	"image/color"
	"sort"
)

var (
	outDir        = flag.String("out.dir", "", "output directory")
	outSuffix     = flag.String("out.suffix", "sorted", "output the sorted image with the given suffix")
	sortDirection = flag.String("sort.direction", "both", "x, y or both")
	sortThreshold = flag.Float64("sort.threshold", 50, "start a new chunk if the colour changes by more than this as a percentage")
	sortMinChunk  = flag.Int("sort.chunk.min", -1, "do not reset chunk unless it is at least this big (in pixels)")
)

func main() {

	flag.Parse()

	inFile, err := os.Open(os.Args[len(os.Args)-1])
	if err != nil {
		log.Fatalf("Failed to open image at %s caused by %s", os.Args[len(os.Args)-1], err)
	}

	extension := path.Ext(path.Base(inFile.Name()))
	if extension != ".png" {
		log.Fatalf("Use a .png image, not a %s", extension)
	}

	originalImage, _, err := image.Decode(inFile)
	if err != nil {
		log.Fatalf("Failed to decode image at %s caused by %s", os.Args[0], err)
	}

	// foo.png -> foo.sorted.png
	outFileName := strings.Replace(path.Base(inFile.Name()), extension, "." + *outSuffix+extension, 1)

	outFile, err := os.OpenFile(path.Join(*outDir, outFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to create output file %s caused by %s", outFileName, err)
	}

	sortedImage := originalImage
	if *sortDirection == "x" || *sortDirection == "both" {
		sortedImage = sortImageX(sortedImage)
	}
	if *sortDirection == "y" || *sortDirection == "both" {
		sortedImage = sortImageY(sortedImage)
	}
	if err := png.Encode(outFile, sortedImage); err != nil {
		log.Fatalf("Failed to encode image: %s", err)
	}
}

func sortImageY(img image.Image) image.Image {

	outImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))

	for col := 0; col < img.Bounds().Dx(); col++ {

		chunk := []color.Color{}

		for row := 0; row < img.Bounds().Dy(); row++ {

			curPixel := img.At(col, row)
			lastPixel := curPixel
			if len(chunk) > 1 {
				lastPixel = chunk[len(chunk)-1]
			}
			chunk = append(chunk, curPixel)

			if len(chunk) > (img.Bounds().Dy()-row) || (colorThreshold(curPixel, lastPixel) && chunkIsBigEnough(chunk)) {

				//sort by lightness
				sort.Slice(chunk, func(i, j int) bool {
					return lightness(chunk[i].RGBA()) > lightness(chunk[j].RGBA())
				})

				//write back into place
				for pos, pxlCol := range chunk {
					outImg.Set(col, (row-len(chunk))+pos, pxlCol)
				}
				chunk = []color.Color{}
			}
		}
	}
	return outImg
}

func sortImageX(img image.Image) image.Image {

	outImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))

	for row := 0; row < img.Bounds().Dy(); row++ {

		chunk := []color.Color{}

		for col := 0; col < img.Bounds().Dx(); col++ {

			curPixel := img.At(col, row)
			lastPixel := curPixel
			if len(chunk) > 1 {
				lastPixel = chunk[len(chunk)-1]
			}
			chunk = append(chunk, curPixel)

			if len(chunk) > (img.Bounds().Dx()-col) || (colorThreshold(curPixel, lastPixel) && chunkIsBigEnough(chunk)) {

				//sort by lightness
				sort.Slice(chunk, func(i, j int) bool {
					return lightness(chunk[i].RGBA()) > lightness(chunk[j].RGBA())
				})

				//write back into place
				for pos, pxlCol := range chunk {
					outImg.Set((col-len(chunk))+pos, row, pxlCol)
				}
				chunk = []color.Color{}
			}
		}
	}
	return outImg
}

func lightness(r, g, b, a uint32) float64 {
	if a == 0 {
		return 0.0
	}
	r = r / 0x101
	g = g / 0x101
	b = b / 0x101
	a = a / 0x101
	aMod := float64(a) / 255.0

	return 0.2126*(float64(r)*aMod) + 0.7152*(float64(g)*aMod) + 0.0722*(float64(b)*aMod)
}

func colorThreshold(curPixel, lastPixel color.Color) bool {
	r, g, b, a := curPixel.RGBA()
	if ( a / 0x101) < 32 {
		return false
	}
	l := lightness(r, g, b, a)

	return !(l < 64 || l > 96)
}

func changeThreshold(curPixel, lastPixel color.Color) bool {

	curTotal := lightness(curPixel.RGBA())
	lastTotal := lightness(lastPixel.RGBA())

	diff := ((curTotal - lastTotal) / lastTotal) * 100

	change := diff < 0 - *sortThreshold || diff > *sortThreshold
	return change
}

func chunkIsBigEnough(chunk []color.Color) bool {
	if *sortMinChunk == -1 {
		return true
	}
	if len(chunk) >= *sortMinChunk {
		return true
	}
	return false
}