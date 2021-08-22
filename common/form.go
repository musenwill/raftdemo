package common

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

type Tag struct {
	K string
	V string
}

type Form struct {
	tags   []Tag
	header []string
	rows   [][]string
	cols   int
}

func NewForm() *Form {
	return &Form{rows: make([][]string, 0)}
}

func (f *Form) SetTags(tags []Tag) {
	f.tags = tags
}

func (f *Form) SetHeader(header []string) {
	f.header = header
	f.cols = len(header)
}

func (f *Form) AddRow(row []string) error {
	if len(row) > f.cols {
		return fmt.Errorf("row has unexpected columns")
	}
	f.rows = append(f.rows, row)
	return nil
}

func (f *Form) String() string {
	colLen := make([]int, f.cols)
	for i, h := range f.header {
		colLen[i] = runewidth.StringWidth(h)
	}
	for _, r := range f.rows {
		for i, s := range r {
			if runewidth.StringWidth(s) > colLen[i] {
				colLen[i] = runewidth.StringWidth(s)
			}
		}
	}

	sb := &strings.Builder{}

	{
		// print tags
		maxKeyWidth := 0
		for _, t := range f.tags {
			if runewidth.StringWidth(t.K) > maxKeyWidth {
				maxKeyWidth = runewidth.StringWidth(t.K)
			}
		}
		for _, t := range f.tags {
			k := fixWidth(t.K, 1, maxKeyWidth)
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, t.V))
		}
	}

	{
		if len(f.header) > 0 {
			// print header
			for i, h := range f.header {
				f.header[i] = fixWidth(h, 0, colLen[i])
			}
			sb.WriteString(fmt.Sprintln(strings.Join(f.header, " ")))
		}
	}

	{
		if len(colLen) > 0 {
			// print seperator
			var sps []string
			for _, c := range colLen {
				sps = append(sps, strings.Repeat("-", c))
			}
			sb.WriteString(fmt.Sprintln(strings.Join(sps, " ")))
		}
	}

	{
		// print rows
		for _, row := range f.rows {
			var line []string
			for i, r := range row {
				line = append(line, fixWidth(r, -1, colLen[i]))
			}
			sb.WriteString(fmt.Sprintln(strings.Join(line, " ")))
		}
	}

	return sb.String()
}

// align: -1 left, 0 centre, 1 right
func fixWidth(text string, align int, fixWidth int) string {
	textWidth := runewidth.StringWidth(text)
	if textWidth >= fixWidth {
		return text
	}

	var printText string
	if align < 0 {
		printText = text + strings.Repeat(" ", fixWidth-textWidth)
	} else if align == 0 {
		leftAlign := (fixWidth - textWidth) / 2
		rightAligin := fixWidth - textWidth - leftAlign
		printText = strings.Repeat(" ", leftAlign) + text + strings.Repeat(" ", rightAligin)
	} else {
		printText = strings.Repeat(" ", fixWidth-textWidth) + text
	}

	return printText
}
