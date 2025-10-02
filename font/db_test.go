package font_test

import (
	"os"
	"testing"

	"github.com/AkimioJR/assfonts-go/ass"
	"github.com/AkimioJR/assfonts-go/font"
)

const (
	fontDir = "../test_case/fonts"
	assPath = "../test_case/subtitles/[UHA-WINGS&VCB-Studio] EIGHTY SIX [S01E01][Ma10p_1080p][x265_flac_aac].chs.ass"
)

func BenchmarkBuildDB(b *testing.B) {
	for b.Loop() {
		db, err := font.NewFontDataBase(nil)
		if err != nil {
			b.Fatal(err)
		}
		defer db.Close()
		err = db.BuildDB([]string{fontDir}, true, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSubset(b *testing.B) {
	db, err := font.NewFontDataBase(nil)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	err = db.BuildDB([]string{fontDir}, true, nil)
	if err != nil {
		b.Fatal(err)
	}
	file, err := os.Open(assPath)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	ap, err := ass.NewASSParser(file)
	if err != nil {
		b.Fatal(err)
	}
	err = ap.Parse()
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		_, err := db.Subset(ap)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSubsetWithBigMemoryMode(b *testing.B) {
	db, err := font.NewFontDataBase(nil)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	db.BigMemoryMode = true
	err = db.BuildDB([]string{fontDir}, true, nil)
	if err != nil {
		b.Fatal(err)
	}
	file, err := os.Open(assPath)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	ap, err := ass.NewASSParser(file)
	if err != nil {
		b.Fatal(err)
	}
	err = ap.Parse()
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		_, err := db.Subset(ap)
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkSubsetConcurrent(b *testing.B) {
	db, err := font.NewFontDataBase(nil)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	err = db.BuildDB([]string{fontDir}, true, nil)
	if err != nil {
		b.Fatal(err)
	}
	file, err := os.Open(assPath)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	ap, err := ass.NewASSParser(file)
	if err != nil {
		b.Fatal(err)
	}
	err = ap.Parse()
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		_, err := db.Subset(ap, font.WithConcurrent())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSubsetConcurrentWithBigMemoryMode(b *testing.B) {
	db, err := font.NewFontDataBase(nil)
	if err != nil {
		b.Fatal(err)
	}
	db.BigMemoryMode = true
	defer db.Close()
	err = db.BuildDB([]string{fontDir}, true, nil)
	if err != nil {
		b.Fatal(err)
	}
	file, err := os.Open(assPath)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	ap, err := ass.NewASSParser(file)
	if err != nil {
		b.Fatal(err)
	}
	err = ap.Parse()
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		_, err := db.Subset(ap, font.WithConcurrent())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSubsetCompelte(b *testing.B) {
	db, err := font.NewFontDataBase(nil)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	err = db.BuildDB([]string{fontDir}, true, nil)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		file, err := os.Open(assPath)
		if err != nil {
			b.Fatal(err)
		}

		ap, err := ass.NewASSParser(file)
		if err != nil {
			b.Fatal(err)
		}
		err = ap.Parse()
		if err != nil {
			b.Fatal(err)
		}
		_, err = db.Subset(ap, font.WithConcurrent())
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
	}
}
