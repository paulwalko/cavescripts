package main

/*
	#cgo LDFLAGS: -lm
	#include "img.h"
*/
import "C"
import (
	"bufio"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

const Ft_to_M = 0.3048

func main() {
	// Process args
	var lstFile string
	var sefFile string
	flag.StringVar(&lstFile, "lst", "", "Walls LST file containing vertex data")
	flag.StringVar(&sefFile, "sef", "", "Walls SEF file containing LRUD data")
	flag.Parse()
	if lstFile == "" {
		log.Fatal("please pass -lst arg")
	}
	if sefFile == "" {
		log.Fatal("please pass -sef arg")
	}

	// Determine output file
	outFile := lstFile[0:len(lstFile)-4] + ".3d"
	log.Println("Output file: " + outFile)

	/* SEF */
	f, _ := os.Open(sefFile)
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)
	lrudMap := make(map[string][]float64)
	inSurvey := false
	surveyLineLen := 0
	lrudIndex := []int{-1, -1, -1, -1, -1, -1} // from, to, L, R, U, D
	for fileScanner.Scan() {
		lineDataStr := fileScanner.Text()
		lineData := strings.Split(lineDataStr, ",")

		/* Get LRUD indexes */
		if strings.HasPrefix(lineDataStr, "#endctsurvey") || strings.HasPrefix(lineDataStr, "#endcpoint") {
			lrudIndex = []int{-1, -1, -1, -1, -1, -1}
			inSurvey = false
		} else if strings.HasPrefix(lineDataStr, "#data") {
			inSurvey = true
			surveyLineLen = len(lineData)
			for i := 0; i < len(lineData); i++ {
				if lineData[i] == "from" {
					lrudIndex[0] = i
				} else if lineData[i] == "to" {
					lrudIndex[1] = i
				} else if lineData[i] == "left" {
					lrudIndex[2] = i
				} else if lineData[i] == "right" {
					lrudIndex[3] = i
				} else if lineData[i] == "ceil" {
					lrudIndex[4] = i
				} else if lineData[i] == "floor" {
					lrudIndex[5] = i
				}
			}
			/* Get LRUDs */
		} else if inSurvey && len(lineData) == surveyLineLen && strings.HasPrefix(lineDataStr, " ") {
			if lrudIndex[0] > -1 && lrudIndex[1] > -1 {
				//from := lineData[lrudIndex[0]]
				to := lineData[lrudIndex[1]]
				left, _ := strconv.ParseFloat(lineData[lrudIndex[2]], 64)
				right, _ := strconv.ParseFloat(lineData[lrudIndex[3]], 64)
				up, _ := strconv.ParseFloat(lineData[lrudIndex[4]], 64)
				down, _ := strconv.ParseFloat(lineData[lrudIndex[5]], 64)
				lrudMap[to] = []float64{left, right, up, down}
			}

		}
	}

	/* LST */
	f, _ = os.Open(lstFile)

	// Setup line scanner
	fileScanner = bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)

	// Setup .3d file
	var pimg *C.img
	labelMap := make(map[string]bool)

	// Process lines
	Title := ""
	lineNum := 0
	PrefixOffset := 1 // Lines are 6 fields long w/ prefix, otherwise 5
	for fileScanner.Scan() {
		lineNum += 1
		lineDataStr := fileScanner.Text()
		lineData := strings.Fields(lineDataStr)

		/* Init 3d file */
		if lineNum == 1 {
			Title = strings.TrimSpace(lineDataStr)
			pimg = C.img_open_write_cs(C.CString(outFile), C.CString(Title), nil, 0)
		}
		/* Skip header & footer */
		if lineNum < 10 || len(lineData) == 0 || lineData[0] == "Vectors" {
			continue
		}
		/* Check if Prefix exists */
		if strings.HasPrefix(lineDataStr, "\t") {
			PrefixOffset = 0
		}

		/* Parse Prefix, Label */
		Prefix := ""
		Name := strings.Replace(lineData[PrefixOffset], ".", "_", -1)
		if PrefixOffset == 1 {
			Prefix = lineData[0] + "."
		}
		Label := Prefix + Name

		/* Parse X Y Z */
		MvXFt, _ := strconv.ParseFloat(lineData[1+PrefixOffset], 64)
		MvYFt, _ := strconv.ParseFloat(lineData[2+PrefixOffset], 64)
		MvZFt, _ := strconv.ParseFloat(lineData[3+PrefixOffset], 64)

		/* Survey fields */
		pimg.style = C.img_STYLE_NORMAL
		if _, ok := lrudMap[Name]; ok {
			pimg.l = C.double(lrudMap[Name][0])
			pimg.r = C.double(lrudMap[Name][1])
			pimg.u = C.double(lrudMap[Name][2])
			pimg.d = C.double(lrudMap[Name][3])
		}

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
		if len(lineData) == 4+PrefixOffset {
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
		if len(lineData) == 5+PrefixOffset {
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
