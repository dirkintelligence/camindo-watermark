package main

import (
	"errors"
	"fmt"
	"strings"
)

// Codes: up to six characters from A–Z and "-" (one per camindo instance, e.g.
// GS-NRD). Each character is a digit 1–27 in base 28 (0 = no character),
// left-aligned: 28^6 = 481 890 304 values fit into the 32-bit payload (27^7
// would not, so six is the maximum).
const (
	codeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ-"
	codeMaxLen   = 6
	codeBase     = 28
)

// code is the cli.Codec: flag -code, JSON field "code".
type code struct{}

func (code) Name() string { return "code" }

func (code) Usage() string {
	return `code to hide: 1–6 characters A–Z and "-", e.g. GS-NRD (lower case is converted)`
}

func (code) Encode(text string) (uint32, error) { return codeToID(text) }

func (code) Decode(id uint32) (string, bool) {
	c := idToCode(id)
	if c == nil {
		return "", false
	}
	return *c, true
}

// codeToID packs a code into the 32-bit payload. Lower case is accepted.
func codeToID(code string) (uint32, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) == 0 || len(code) > codeMaxLen {
		return 0, fmt.Errorf("-code must have 1 to %d characters", codeMaxLen)
	}
	var v uint32
	for i := 0; i < codeMaxLen; i++ {
		d := 0
		if i < len(code) {
			d = strings.IndexByte(codeAlphabet, code[i]) + 1
			if d == 0 {
				return 0, errors.New(`-code may only contain the letters A–Z and "-"`)
			}
		}
		v = v*codeBase + uint32(d)
	}
	return v, nil
}

// idToCode returns the code of a payload, nil if the number is no valid code
// (too large, empty, or a gap before the last character).
func idToCode(id uint32) *string {
	if id >= codeBase*codeBase*codeBase*codeBase*codeBase*codeBase || id == 0 {
		return nil
	}
	var digits [codeMaxLen]int
	for i := codeMaxLen - 1; i >= 0; i-- {
		digits[i] = int(id % codeBase)
		id /= codeBase
	}
	var b []byte
	ended := false
	for _, d := range digits {
		if d == 0 {
			ended = true
			continue
		}
		if ended {
			return nil
		}
		b = append(b, codeAlphabet[d-1])
	}
	s := string(b)
	return &s
}
