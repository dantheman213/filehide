package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

var eofSignature []byte = []byte{0x05, 0xFF, 0x05, 0xFA, 0x55}

func main() {
	args := os.Args[1:]
	action := args[0]
	keyPhrase := args[1]
	inFile := args[2]
	outFile := args[3]

	if action == "encrypt" {
		encryptFileAction(keyPhrase, inFile, outFile)
	} else if action == "decrypt" {
		decryptFileAction(keyPhrase, inFile, outFile)
	}

	fmt.Println("completed!")
}

func encryptFileAction(key, sourceFilePath, encryptedFilePath string) {
	sourceFileBytes, err := loadBinaryFileToMemory(sourceFilePath)
	if err != nil {
		panic(err)
	}


	// add eof signature
	sourceFileBytes = append(sourceFileBytes, eofSignature...)

	// pad bytes so they're divisible by 4
	for len(sourceFileBytes) % 4 != 0 {
		sourceFileBytes = append(sourceFileBytes, 0x00)
	}
	byteCount := len(sourceFileBytes)
	pixelCount := byteCount / 4
	pixelDimension := math.Sqrt(float64(pixelCount))

	// round up if there's a remainder
	if pixelDimension != float64(int64(pixelDimension)) {
		pixelDimension = float64(int64(pixelDimension + 1))
	}

	width := int(pixelDimension)
	height := width

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	bPos := 0
	for y := 0; y <= height; y++ {
		for x := 0; x <= width; x++ {
			if bPos >= byteCount {
				//img.Set(x, y, color.RGBA{0x00, 0x00,0x00, 0x00})
				//continue
				break
			}

			//fmt.Printf("Added pixel at %d, %d, bytePos at %d \\ %d \n", x, y, bPos, byteCount)
			img.Set(x, y, color.RGBA{sourceFileBytes[bPos], sourceFileBytes[bPos+1], sourceFileBytes[bPos+2], sourceFileBytes[bPos+3]})
			bPos += 4
		}
	}

	f, _ := os.Create(encryptedFilePath)
	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
}

func loadBinaryFileToMemory(filename string) ([]byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return bytes, err
}

func decryptFileAction(key, sourceImagePath, decryptedFilePath string) {
	sourceImage, err := os.Open(sourceImagePath)
	if err != nil {
		panic(err)
	}
	defer sourceImage.Close()

	img, err := png.Decode(sourceImage)
	if err != nil {
		panic(err)
	}

	var dat = []byte{}
	for y := 0; y <= img.Bounds().Max.Y; y++ {
		for x := 0; x <= img.Bounds().Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			r, g, b, a = r>>8, g>>8, b>>8, a>>8 // convert from 32 bit to 8 bit
			dat = append(dat, uint8(r), uint8(g), uint8(b), uint8(a))

		}
	}

	// remove all signatures bytes and anything afterward
	for bPos := len(dat) - len(eofSignature); bPos > 0; bPos-- {
		if compareSlice(dat[bPos:bPos+len(eofSignature)], eofSignature) {
			dat = dat[0:bPos]
			break
		}
	}

	f, err := os.Create(decryptedFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(dat); err != nil {
		panic(err)
	}
}

func compareSlice(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
