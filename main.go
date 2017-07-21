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
	outDir    = flag.String("out.dir", "", "output directory")
	outSuffix = flag.String("out.suffix", "sorted", "output the sorted image with the given suffix")
	sortChunk = flag.Int("sort.chunk", -1, "size of each sorted chunk")
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

	originalImage, err := png.Decode(inFile)
	if err != nil {
		log.Fatalf("Failed to decode image at %s caused by %s", os.Args[0], err)
	}

	// foo.png -> foo.sorted.png
	outFileName := strings.Replace(path.Base(inFile.Name()), extension, "." + *outSuffix+extension, 1)

	outFile, err := os.OpenFile(path.Join(*outDir, outFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to create output file %s caused by %s", outFileName, err)
	}

	sortedImage := sortImage(originalImage)

	if err := png.Encode(outFile, sortedImage); err != nil {
		log.Fatalf("Failed to encode image: %s", err)
	}
}

func sortImage(img image.Image) image.Image {

	outImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))

	chunkSize := *sortChunk
	if chunkSize == -1 {
		chunkSize = img.Bounds().Dy()
	}

	chunk := []color.Color{}
	for col := 0; col < img.Bounds().Dx(); col++ {
		//does not wrap
		chunk = []color.Color{}
		for row := 0; row < img.Bounds().Dy(); row++ {

			chunk = append(chunk, img.At(col, row))
			if len(chunk) > chunkSize || (row == img.Bounds().Dy()-1 || col == img.Bounds().Dx()-1) {

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

func lightness(r, g, b, a uint32) float64 {
	if a == 0 {
		return 0.0
	}
	r = r/0x101
	g = g/0x101
	b = b/0x101
	a = a/0x101
	aMod := float64(a) / 255.0
	l := 0.2126*(float64(r) * aMod) + 0.7152*(float64(g) * aMod) + 0.0722*(float64(b)*aMod)
	return l
}
