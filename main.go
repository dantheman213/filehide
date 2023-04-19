package main

import (
	"bufio"
	"filehide/pkg/compression"
	"filehide/pkg/encryption"
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
	key := []byte (args[1])
	inFile := args[2]
	outFile := args[3]

	if action == "encrypt" {
		encryptFileAction(key, inFile, outFile)
	} else if action == "decrypt" {
		decryptFileAction(key, inFile, outFile)
	}

	fmt.Println("completed!")
}

func encryptFileAction(key []byte, sourceFilePath, encryptedFilePath string) {
	sourceFileBytes, err := loadBinaryFileToMemory(sourceFilePath)
	if err != nil {
		panic(err)
	}

	compressedFileBytes := compression.CompressBinary(sourceFileBytes)
	nonce, payload := encryption.Encrypt(key, compressedFileBytes)

	// add nonce -> payload -> eof signature to new payload
	payload = append(nonce, payload...)
	payload = append(payload, eofSignature...)

	// pad bytes so they're divisible by 3
	for len(payload) % 3 != 0 {
		payload = append(payload, 0x00)
	}
	byteCount := len(payload)
	pixelCount := byteCount / 3
	pixelDimension := math.Sqrt(float64(pixelCount))

	// round up if there's a remainder
	if pixelDimension != float64(int64(pixelDimension)) {
		pixelDimension = float64(int64(pixelDimension + 1))
	}

	width := int(pixelDimension)
	height := width

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	bPos := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if bPos >= byteCount {
				//img.Set(x, y, color.RGBA{0x00, 0x00,0x00, 0x00})
				//continue
				break
			}

			//fmt.Printf("Added pixel at %d, %d, bytePos at %d \\ %d \n", x, y, bPos, byteCount)
			img.Set(x, y, color.RGBA{payload[bPos], payload[bPos+1], payload[bPos+2], 0xFF})
			bPos += 3
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

func decryptFileAction(key []byte, sourceImagePath, decryptedFilePath string) {
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
	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			r, g, b = r>>8, g>>8, b>>8 // convert from 32 bit to 8 bit
			dat = append(dat, uint8(r), uint8(g), uint8(b))
		}
	}

	// remove all signatures bytes and anything afterward
	for bPos := len(dat) - len(eofSignature); bPos > 0; bPos-- {
		if compareSlice(dat[bPos:bPos+len(eofSignature)], eofSignature) {
			dat = dat[0:bPos]
			break
		}
	}

	nonce := dat[0:12]
	dat = dat[12:] // remove nonce from decryption payload
	decryptedData := encryption.Decrypt(key, nonce, dat)
	decompressedData := compression.DecompressBinary(decryptedData)

	f, err := os.Create(decryptedFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(decompressedData); err != nil {
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
