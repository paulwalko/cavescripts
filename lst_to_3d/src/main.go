package main

/*
	#cgo LDFLAGS: -lm
	#include "img.h"
*/
import "C"
import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"
)

const Ft_to_M = 0.3048

func main() {
	// Process args
	var lstFile string
	flag.StringVar(&lstFile, "lst", "", "Usage")
	flag.Parse()
	if lstFile == "" {
		panic("please pass -lst arg")
	}

	// Determine output file
	outFile := strings.Join([]string{lstFile[0 : len(lstFile)-4], ".3d"}, "")

	// Open file
	f, err := os.Open(lstFile)
	if err != nil {
		panic(err)
	}

	// Setup line scanner
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)

	// Setup .3d file
	var pimg *C.img
	labelMap := make(map[string]bool)

	// Process lines
	Title := ""
	lineNum := 0
	for fileScanner.Scan() {
		lineNum += 1
		lineData := strings.Fields(fileScanner.Text())

		/* Init 3d file */
		if lineNum == 1 {
			Title = lineData[0]
			pimg = C.img_open_write_cs(C.CString(outFile), C.CString(Title), nil, 0)
		}
		/* Skip header & footer */
		if lineNum < 10 || len(lineData) == 0 || lineData[0] == "Vectors" {
			continue
		}

		/* Fields in go format */
		Label := strings.Join([]string{lineData[0], lineData[1]}, ".")
		MvXFt, _ := strconv.ParseFloat(lineData[2], 64)
		MvYFt, _ := strconv.ParseFloat(lineData[3], 64)
		MvZFt, _ := strconv.ParseFloat(lineData[4], 64)

		/* Survey Style */
		pimg.style = C.img_STYLE_NORMAL

		/* LABEL */
		if _, ok := labelMap[Label]; !ok {
			Code := C.img_LABEL
			Flags := 0x02

			C.img_write_item(
				pimg,
				C.int(Code),
				C.int(Flags),
				C.CString(Label),
				C.double(MvXFt*Ft_to_M),
				C.double(MvYFt*Ft_to_M),
				C.double(MvZFt*Ft_to_M),
			)
			labelMap[Label] = true
		}

		/* MOVE */
		if len(lineData) == 5 {
			Code := C.img_MOVE
			Flags := 0

			C.img_write_item(
				pimg,
				C.int(Code),
				C.int(Flags),
				C.CString(""),
				C.double(MvXFt*Ft_to_M),
				C.double(MvYFt*Ft_to_M),
				C.double(MvZFt*Ft_to_M),
			)
		}

		/* LINE */
		if len(lineData) == 6 {
			Code := C.img_LINE
			Flags := 0

			C.img_write_item(
				pimg,
				C.int(Code),
				C.int(Flags),
				C.CString(""),
				C.double(MvXFt*Ft_to_M),
				C.double(MvYFt*Ft_to_M),
				C.double(MvZFt*Ft_to_M),
			)
		}
	}

	// Close 3d file
	C.img_close(pimg)

	// Close lst file
	f.Close()
}
