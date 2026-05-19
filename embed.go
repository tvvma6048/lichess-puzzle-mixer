package main

import "embed"

//go:embed all:web
var webFS embed.FS

//go:embed testdata/lichess_sample.csv
var sampleCSV []byte
