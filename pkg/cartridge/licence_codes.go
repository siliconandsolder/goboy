package cartridge

import "fmt"

func getLicenceNameFromCode(code byte) (string, error) {
	switch code {
	default:
		return "", fmt.Errorf("unknown licence code: %d", code)
	}
}
