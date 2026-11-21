// camindo-watermark: the camindo CMS integration of the invisible image
// watermark image-watermark. A thin program on that separate library
// (github.com/andreasmayercode/image-watermark): namespace "camindo-watermark", key variable CAMINDO_WM_KEY and codes of
// 1–6 characters (A–Z and "-") per instance instead of bare numbers.
//
// Copyright (C) 2026 the camindo authors (https://github.com/dirkintelligence)
// SPDX-License-Identifier: AGPL-3.0-or-later
// See LICENSE and THIRD-PARTY-NOTICES.md.
package main

import (
	_ "embed"

	"github.com/andreasmayercode/image-watermark/cli"
)

var version = "1.0.0"

// namespace of the data format: part of the key derivation, never change it -
// images marked since 1.0 stay readable only with it
const namespace = "camindo-watermark"

//go:embed web/index.html
var indexHTML []byte

func main() {
	cli.Main(cli.Config{
		Name:      "camindo-watermark",
		Version:   version,
		KeyEnv:    "CAMINDO_WM_KEY",
		Namespace: namespace,
		Codec:     code{},
		IndexHTML: indexHTML,
	})
}
