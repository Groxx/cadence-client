package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"go.uber.org/cadence/internal/tools/publictypes"
)

func main() {
	singlechecker.Main(publictypes.Analyzer)
}
