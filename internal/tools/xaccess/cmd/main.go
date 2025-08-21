package main

import (
	"go.uber.org/cadence/internal/tools/xaccess"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(xaccess.Analyzer)
}
