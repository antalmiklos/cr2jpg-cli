package main

import (
	"bytes"
	"flag"
	"image/jpeg"
	"log"
	"os"

	"github.com/nf/cr2"
	"github.com/nfnt/resize"
)

const (
	NO_RENAME = "noRename"
)

var (
	src         string
	dest        string
	destFile    string
	size        float64
	threadCount int
)

func main() {
	flag.StringVar(&src, "source", "", "source file path or directory")
	flag.StringVar(&dest, "dest", "", "destination path")
	flag.StringVar(&destFile, "destFile", "", "destination file name if source is not a path")
	flag.Float64Var(&size, "size", 1.0, "scale of the new image")
	flag.IntVar(&threadCount, "threads", 1, "number of threads to use WIP")

	flag.Parse()

	srcVal, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}

	if srcVal.IsDir() {
		files, err := os.ReadDir(src)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range files {
			orig := srcVal.Name() + "/" + v.Name()
			convert(orig, NO_RENAME)
		}
	} else {
		convert(src, destFile)
	}
}

func getXY(origX, origY int) (int, int) {
	return int(float64(origX) * size), int(float64(origY) * size)
}

func convert(src string, destFile string) error {

	// target dimensions of the image
	var newX, newY int

	orig, err := os.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}

	var origBuf bytes.Buffer
	origBuf.Write(orig)

	img, err := cr2.Decode(&origBuf)
	if err != nil {
		log.Fatal(err)
	}

	newX, newY = getXY(img.Bounds().Dx(), img.Bounds().Dy())

	newImage := resize.Resize(uint(newX), uint(newY), img, resize.Lanczos3)

	file, err := os.Stat(src)
	if err != nil {
		return err
	}

	if destFile == NO_RENAME {
		destFile = file.Name()
	}

	newFile, err := os.Create(destFile + ".jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	err = jpeg.Encode(newFile, newImage, nil)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
