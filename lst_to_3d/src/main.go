package main

/*
	#cgo LDFLAGS: -lm
	#include "img.h"
*/
import "C"
import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const Ft_to_M = 0.3048

func main() {
	// Open file
	f, err := os.Open("data.lst")
	if err != nil {
		panic(err)
	}

	// Setup line scanner
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)

	// Setup .3d file
	pimg := C.img_open_write_cs(C.CString("output.3d"), C.CString("my survey"), nil, 0)
	labelMap := make(map[string]bool)

	// Process each line
	//	lineNum := 0
	for fileScanner.Scan() {
		lineData := strings.Fields(fileScanner.Text())
		//
		//		lineNum += 1
		//		if lineNum < 10 || len(lineData) < 5 || lineNum > 1000 {
		//			continue
		//		}

		/* Fields in go format */
		Label := strings.Join([]string{lineData[0], lineData[1]}, ".")
		MvXFt, _ := strconv.ParseFloat(lineData[2], 64)
		MvYFt, _ := strconv.ParseFloat(lineData[3], 64)
		MvZFt, _ := strconv.ParseFloat(lineData[4], 64)

		/* LABEL */
		if _, ok := labelMap[Label]; !ok {
			Code := C.img_LABEL
			Flags := 0x02

			// Write fields
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

	// Close 3d file
	C.img_close(pimg)

	// Close lst file
	f.Close()
}
