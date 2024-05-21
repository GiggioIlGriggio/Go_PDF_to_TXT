package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"rsc.io/pdf"
)

func main() {
	inFile := "Aprile Telepass 3.pdf"
	var file *os.File
	var buf bytes.Buffer
	file, err := os.Open(inFile)
	if err != nil {
		fmt.Println("Error during opening", err.Error())
		return
	}
	conf := model.NewDefaultConfiguration()
	ctx, err := api.ReadValidateAndOptimize(file, conf)
	if err != nil {
		fmt.Println("READ validate error", err.Error())
		return
	}
	api.Write(ctx, &buf, conf)
	f, err := pdf.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		fmt.Println("Open PDF file error:", err.Error())
		return
	}
	content := ""
	n := f.NumPage()
	//PageLoop:
	for pageNum := 1; pageNum <= n; pageNum++ {
		p := f.Page(pageNum)
		txt := parsePage(pageNum, p)
		for _, tw := range txt {
			content += tw.S + " "
		}
	}
	fmt.Println(content)

}

func parsePage(num int, p pdf.Page) []pdf.Text {
	content := p.Content()

	var text []pdf.Text
	for _, t := range content.Text {
		// .... potentially applies custom filter to single character
		text = append(text, t)
	}

	// tokenize in words
	text = findWords(text)

	// Page header must say Instruction Details.
	return text
}

type RectHorizontal []pdf.Rect

func (x RectHorizontal) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x RectHorizontal) Less(i, j int) bool { return x[i].Min.X < x[j].Min.X }
func (x RectHorizontal) Len() int           { return len(x) }

func match(t pdf.Text, font string, size float64, substr string) bool {
	return t.Font == font && math.Abs(t.FontSize-size) < 0.1 && strings.Contains(t.S, substr)
}

func findWords(chars []pdf.Text) (words []pdf.Text) {
	// Sort by Y coordinate and normalize.
	const nudge = 1
	sort.Sort(pdf.TextVertical(chars))
	old := -100000.0
	for i, c := range chars {
		if c.Y != old && math.Abs(old-c.Y) < nudge {
			chars[i].Y = old
		} else {
			old = c.Y
		}
	}

	// Sort by Y coordinate, breaking ties with X.
	// This will bring letters in a single word together.
	sort.Sort(pdf.TextVertical(chars))

	// Loop over chars.
	for i := 0; i < len(chars); {
		// Find all chars on line.
		j := i + 1
		for j < len(chars) && chars[j].Y == chars[i].Y {
			j++
		}
		var end float64
		// Split line into words (really, phrases).
		for k := i; k < j; {
			ck := &chars[k]
			s := ck.S
			end = ck.X + ck.W
			charSpace := ck.FontSize / 6
			wordSpace := ck.FontSize * 2 / 3
			l := k + 1
			for l < j {
				// Grow word.
				cl := &chars[l]
				if sameFont(cl.Font, ck.Font) && math.Abs(cl.FontSize-ck.FontSize) < 0.1 && cl.X <= end+charSpace {
					s += cl.S
					end = cl.X + cl.W
					l++
					continue
				}
				// Add space to phrase before next word.
				if sameFont(cl.Font, ck.Font) && math.Abs(cl.FontSize-ck.FontSize) < 0.1 && cl.X <= end+wordSpace {
					s += " " + cl.S
					end = cl.X + cl.W
					l++
					continue
				}
				break
			}
			f := ck.Font
			f = strings.TrimSuffix(f, ",Italic")
			f = strings.TrimSuffix(f, "-Italic")
			words = append(words, pdf.Text{f, ck.FontSize, ck.X, ck.Y, end - ck.X, s})
			k = l
		}
		i = j
	}

	return words
}

func sameFont(f1, f2 string) bool {
	f1 = strings.TrimSuffix(f1, ",Italic")
	f1 = strings.TrimSuffix(f1, "-Italic")
	f2 = strings.TrimSuffix(f1, ",Italic")
	f2 = strings.TrimSuffix(f1, "-Italic")
	return strings.TrimSuffix(f1, ",Italic") == strings.TrimSuffix(f2, ",Italic") || f1 == "Symbol" || f2 == "Symbol" || f1 == "TimesNewRoman" || f2 == "TimesNewRoman"
}
