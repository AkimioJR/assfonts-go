package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AkimioJR/assfonts-go/ass"
	"github.com/AkimioJR/assfonts-go/font"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
	ColorCyan   = "\033[36m"
)

var (
	dbPath                = flag.String("db", "", "Path to the font database file, if not specified it will rebuild database and don't save the it")
	inputASSPath          = flag.String("input", "", "Path to the input ass file")
	outputASSPath         = flag.String("output", "", "Path to the input ass file")
	customFontsDir        = flag.String("fontdir", "", "Path to the font dir in order to build database, use ',' to split it")
	withSystemDefaultFont = flag.Bool("system", true, "Include system default fonts when building database")
)

func logger(err error) bool {
	switch err.(type) {
	case *font.ErrUnsupportedPlatform, *font.ErrUnsupportedID:

	case *font.InfoMsg:
		fmt.Printf("%s[INFO]%s %s\n", ColorBlue, ColorReset, err.Error())
	case *font.WarningMsg:
		fmt.Printf("%s[WARNING]%s %s\n", ColorYellow, ColorReset, err.Error())
	default:
		fmt.Printf("%s[ERROR]%s %s\n", ColorRed, ColorReset, err.Error())
	}
	return true
}

func main() {
	flag.Parse()

	db, err := font.NewFontDataBase(nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if *dbPath != "" {
		err = db.LoadDB(*dbPath)
		if err != nil {
			panic(err)
		}
	} else {
		err = db.BuildDB(strings.Split(*customFontsDir, ","), *withSystemDefaultFont, logger)
		if err != nil {
			panic(err)
		}
	}

	inputASS, err := os.Open(*inputASSPath)
	if err != nil {
		panic(err)
	}
	defer inputASS.Close()

	ap, err := ass.NewASSParser(inputASS)
	if err != nil {
		panic(err)
	}
	ap.Parse()

	data, err := db.Subset(ap, font.WithCheckErr(logger), font.WithConcurrent(), font.WithCheckGlyph())
	if err != nil {
		panic(err)
	}

	outputASS, err := os.Create(*outputASSPath)
	if err != nil {
		panic(err)
	}

	err = ap.WriteWithEmbeddedFonts(data, outputASS)
	if err != nil {
		panic(err)
	}

	fmt.Println("success!")
}
