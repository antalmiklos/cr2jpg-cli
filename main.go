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

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

type Box struct {
	Size uint32
	Type [4]byte
}
type BigBox struct {
	Size int64
	Type [4]byte
}

func ReadData(f io.ReadSeeker, size int) ([]byte, error) {
	d := make([]byte, size)
	err := binary.Read(f, binary.BigEndian, &d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

type MDAT struct {
	Size      uint64
	Type      [4]byte
	LargeSize uint64
	UUID      [16]byte
}

type DataBox struct {
	Data []byte
	Type string
}

type FTYP struct {
	Box
	Major [4]byte
	Minor [4]byte
}

const ()

func (b *FTYP) print() {
	fmt.Println(string(b.Major[:]))
	fmt.Println(string(b.Minor[:]))
}

func cmpType(t string, a [4]byte) bool {
	for i, c := range a {
		if c != t[i] {
			return false
		}
	}
	return true
}

func writeToFile(d []byte, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	_, _ = f.Write(d)
}

func findBox(container []byte) []DataBox {
	var box Box
	buf := bytes.NewReader(container)
	boxes := []DataBox{}

	for {
		err := binary.Read(buf, binary.BigEndian, &box)
		d := make([]byte, box.Size-8)
		if err != nil {
			break
		}
		err = binary.Read(buf, binary.BigEndian, &d)
		if err != nil {
			fmt.Println(err)
		}
		boxes = append(boxes, DataBox{Data: d, Type: string(box.Type[:])})
		// buf.Seek(int64(box.Size)-8, io.SeekCurrent)
	}
	return boxes
}

func findTagIndex(box []byte, flag string) int {
	return bytes.Index(box, []byte(flag))
}

func cr3() {
	var acnt int
	for {
		acnt++
		file, err := os.Open("test.cr3")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		s, _ := file.Stat()
		size := s.Size()

		// var offset int64

		var box Box
		var offset int64
		var moov []byte
		var mdat []byte
		for {
			err = binary.Read(file, binary.BigEndian, &box)
			if err != nil {
				if err.Error() == "EOF" {
					fmt.Println("end of file")
					return
				}
				log.Fatal(err)
			}
			if cmpType("mdat", box.Type) {
				file.Seek(offset, io.SeekStart)
				mdat, _ = ReadData(file, int(size-offset))
				break
			} else if cmpType("moov", box.Type) {
				moov, _ = ReadData(file, int(box.Size))
				offset, _ = file.Seek(-8, io.SeekCurrent)
			} else {
				offset, _ = file.Seek(int64(box.Size)-8, io.SeekCurrent)
			}
		}
		moovsub := findBox(moov)
		var trakCnt int = 1
		for _, v := range moovsub {
			var jpgoffset uint32
			var jpgsize uint32
			if v.Type == "trak" {
				traksub := findBox(v.Data)
				for _, x := range traksub {
					if x.Type == "mdia" {
						mdiasub := findBox(x.Data)
						for _, y := range mdiasub {
							if y.Type == "minf" {
								minfsub := findBox(y.Data)
								for _, z := range minfsub {
									if z.Type == "stbl" {
										stblsub := findBox(z.Data)
										for _, zz := range stblsub {
											if zz.Type == "co64" {
												jpgoffset = binary.BigEndian.Uint32(zz.Data[len(zz.Data)-4:])
											}
											if zz.Type == "stsz" {
												jpgsize = binary.BigEndian.Uint32(zz.Data[len(zz.Data)-4:])
											}
											if jpgsize != 0 && trakCnt == 1 {
												img := make([]byte, jpgsize)
												jpgoffset = 16
												img = mdat[jpgoffset : jpgsize+uint32(offset)]
												writeToFile(img, "as2.jpg")
											}
										}

									}
								}
							}
						}
					}
				}
				v.Type = fmt.Sprintf("trak%d", trakCnt)
				trakCnt++
			}
			// fmt.Println(v.Type)
		}
		fmt.Println(acnt)

	}

}
