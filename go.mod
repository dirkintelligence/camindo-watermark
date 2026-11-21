module github.com/dirkintelligence/camindo-watermark

go 1.24.0

toolchain go1.24.13

require github.com/andreasmayercode/image-watermark v1.0.0

require golang.org/x/image v0.36.0 // indirect

// local development: the library is checked out next to this repository.
// Remove before a release once image-watermark v1.0.0 is tagged on GitHub.
replace github.com/andreasmayercode/image-watermark => ../image-watermark
