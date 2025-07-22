package main

import (
	"fmt"
	"unicode"
)

func describe_rune(r rune) {
	fmt.Printf("🔤 Rune: %q\n", r)
	fmt.Printf("📜 Unicode Code Point: U+%04X\n", r)
	fmt.Printf("🔢 Decimal Value: %d\n", r)
	fmt.Printf("🧬 UTF-8 Encoding (bytes): % X\n", []byte(string(r)))

	// General Unicode Properties
	fmt.Printf("✅ Is Letter: %v\n", unicode.IsLetter(r))
	fmt.Printf("🔢 Is Digit: %v\n", unicode.IsDigit(r))
	fmt.Printf("🔣 Is Symbol: %v\n", unicode.IsSymbol(r))
	fmt.Printf("📛 Is Punctuation: %v\n", unicode.IsPunct(r))
	fmt.Printf("🌐 Is Space: %v\n", unicode.IsSpace(r))
	fmt.Printf("🧪 Is Control Character: %v\n", unicode.IsControl(r))
	fmt.Printf("🪪 Is Graphic (visible): %v\n", unicode.IsGraphic(r))
	fmt.Printf("🗂️ Unicode Category: %s\n", unicode.SimpleFold(r))
}

func main() {
	runes := []rune{'	','	'}

	for _, r := range runes {
		describe_rune(r)
		fmt.Println("")
	}
}