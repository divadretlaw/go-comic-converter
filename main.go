package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/celogeek/go-comic-converter/internal/imageconverter"

	"github.com/celogeek/go-comic-converter/internal/epub"
)

type Profile struct {
	Code        string
	Description string
	Width       int
	Height      int
	Palette     color.Palette
}

var Profiles = []Profile{
	// Kindle
	{"K1", "Kindle 1", 600, 670, imageconverter.PALETTE_4},
	{"K11", "Kindle 11", 1072, 1448, imageconverter.PALETTE_16},
	{"K2", "Kindle 2", 600, 670, imageconverter.PALETTE_15},
	{"K34", "Kindle Keyboard/Touch", 600, 800, imageconverter.PALETTE_16},
	{"K578", "Kindle", 600, 800, imageconverter.PALETTE_16},
	{"KDX", "Kindle DX/DXG", 824, 1000, imageconverter.PALETTE_16},
	{"KPW", "Kindle Paperwhite 1/2", 758, 1024, imageconverter.PALETTE_16},
	{"KV", "Kindle Paperwhite 3/4/Voyage/Oasis", 1072, 1448, imageconverter.PALETTE_16},
	{"KPW5", "Kindle Paperwhite 5/Signature Edition", 1236, 1648, imageconverter.PALETTE_16},
	{"KO", "Kindle Oasis 2/3", 1264, 1680, imageconverter.PALETTE_16},
	{"KS", "Kindle Scribe", 1860, 2480, imageconverter.PALETTE_16},
	// Kobo
	{"KoMT", "Kobo Mini/Touch", 600, 800, imageconverter.PALETTE_16},
	{"KoG", "Kobo Glo", 768, 1024, imageconverter.PALETTE_16},
	{"KoGHD", "Kobo Glo HD", 1072, 1448, imageconverter.PALETTE_16},
	{"KoA", "Kobo Aura", 758, 1024, imageconverter.PALETTE_16},
	{"KoAHD", "Kobo Aura HD", 1080, 1440, imageconverter.PALETTE_16},
	{"KoAH2O", "Kobo Aura H2O", 1080, 1430, imageconverter.PALETTE_16},
	{"KoAO", "Kobo Aura ONE", 1404, 1872, imageconverter.PALETTE_16},
	{"KoN", "Kobo Nia", 758, 1024, imageconverter.PALETTE_16},
	{"KoC", "Kobo Clara HD/Kobo Clara 2E", 1072, 1448, imageconverter.PALETTE_16},
	{"KoL", "Kobo Libra H2O/Kobo Libra 2", 1264, 1680, imageconverter.PALETTE_16},
	{"KoF", "Kobo Forma", 1440, 1920, imageconverter.PALETTE_16},
	{"KoS", "Kobo Sage", 1440, 1920, imageconverter.PALETTE_16},
	{"KoE", "Kobo Elipsa", 1404, 1872, imageconverter.PALETTE_16},
}
var ProfilesIdx = map[string]int{}

func init() {
	for i, p := range Profiles {
		ProfilesIdx[p.Code] = i
	}
}

type Option struct {
	Input   string
	Output  string
	Profile string
	Author  string
	Title   string
	Quality int
	NoCrop  bool
	Algo    string
	LimitMb int
}

func (o *Option) String() string {
	var desc string
	var width, height, level int
	if i, ok := ProfilesIdx[o.Profile]; ok {
		profile := Profiles[i]
		desc = profile.Description
		width = profile.Width
		height = profile.Height
		level = len(profile.Palette)
	}
	limitmb := "nolimit"
	if o.LimitMb > 0 {
		limitmb = fmt.Sprintf("%d Mb", o.LimitMb)
	}

	return fmt.Sprintf(`Go Comic Converter

Options:
    Input   : %s
    Output  : %s
    Profile : %s - %s - %dx%d - %d levels of gray
    Author  : %s
    Title   : %s
    Quality : %d
    Crop    : %v
    Algo    : %s
    LimitMb : %s
`,
		o.Input,
		o.Output,
		o.Profile, desc, width, height, level,
		o.Author,
		o.Title,
		o.Quality,
		!o.NoCrop,
		o.Algo,
		limitmb,
	)
}

func main() {
	availableProfiles := make([]string, 0)
	for _, p := range Profiles {
		availableProfiles = append(availableProfiles, fmt.Sprintf(
			"    - %-7s ( %9s ) - %2d levels of gray - %s",
			p.Code,
			fmt.Sprintf("%dx%d", p.Width, p.Height),
			len(p.Palette),
			p.Description,
		))
	}
	availableAlgo := make([]string, 0)
	for a := range imageconverter.ALGO_GRAY {
		availableAlgo = append(availableAlgo, a)
	}

	opt := &Option{}
	flag.StringVar(&opt.Input, "input", "", "Source of comic to convert: directory, cbz, zip, cbr, rar, pdf")
	flag.StringVar(&opt.Output, "output", "", "Output of the epub (directory or epub): (default [INPUT].epub)")
	flag.StringVar(&opt.Profile, "profile", "", fmt.Sprintf("Profile to use: \n%s", strings.Join(availableProfiles, "\n")))
	flag.StringVar(&opt.Author, "author", "GO Comic Converter", "Author of the epub")
	flag.StringVar(&opt.Title, "title", "", "Title of the epub")
	flag.IntVar(&opt.Quality, "quality", 85, "Quality of the image")
	flag.BoolVar(&opt.NoCrop, "nocrop", false, "Disable cropping")
	flag.StringVar(&opt.Algo, "algo", "default", fmt.Sprintf("Algo for RGB to Grayscale: %s", strings.Join(availableAlgo, ", ")))
	flag.IntVar(&opt.LimitMb, "limitmb", 0, "Limit size of the ePub: Default nolimit (0), Minimum 20")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if opt.Input == "" {
		fmt.Fprintln(os.Stderr, "Missing input or output!")
		flag.Usage()
		os.Exit(1)
	}

	var defaultOutput string
	fi, err := os.Stat(opt.Input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}
	inputBase := filepath.Clean(opt.Input)
	if fi.IsDir() {
		defaultOutput = fmt.Sprintf("%s.epub", inputBase)
	} else {
		ext := filepath.Ext(inputBase)
		defaultOutput = fmt.Sprintf("%s.epub", inputBase[0:len(inputBase)-len(ext)])
	}

	if opt.Output == "" {
		opt.Output = defaultOutput
	}

	if filepath.Ext(opt.Output) != ".epub" {
		fo, err := os.Stat(opt.Output)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			flag.Usage()
			os.Exit(1)
		}
		if !fo.IsDir() {
			fmt.Fprintln(os.Stderr, "output must be an existing dir or end with .epub")
			flag.Usage()
			os.Exit(1)
		}
		opt.Output = filepath.Join(
			opt.Output,
			filepath.Base(defaultOutput),
		)
	}

	profileIdx, profileMatch := ProfilesIdx[opt.Profile]
	if !profileMatch {
		fmt.Fprintln(os.Stderr, "Profile doesn't exists!")
		flag.Usage()
		os.Exit(1)
	}
	profile := Profiles[profileIdx]

	if opt.LimitMb > 0 && opt.LimitMb < 20 {
		fmt.Fprintln(os.Stderr, "LimitMb should be 0 or >= 20")
		flag.Usage()
		os.Exit(1)
	}

	if opt.Title == "" {
		ext := filepath.Ext(defaultOutput)
		opt.Title = filepath.Base(defaultOutput[0 : len(defaultOutput)-len(ext)])
	}

	if _, ok := imageconverter.ALGO_GRAY[opt.Algo]; !ok {
		fmt.Fprintln(os.Stderr, "algo doesn't exists")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, opt)

	if err := epub.NewEpub(&epub.EpubOptions{
		Input:   opt.Input,
		Output:  opt.Output,
		LimitMb: opt.LimitMb,
		Title:   opt.Title,
		Author:  opt.Author,
		ImageOptions: &epub.ImageOptions{
			ViewWidth:  profile.Width,
			ViewHeight: profile.Height,
			Quality:    opt.Quality,
			Crop:       !opt.NoCrop,
			Algo:       opt.Algo,
			Palette:    profile.Palette,
		},
	}).Write(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
