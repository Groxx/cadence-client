package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"go.uber.org/cadence/internal/tools/structfill"
)

func main() {
	singlechecker.Main(structfill.Analyzer)
}
