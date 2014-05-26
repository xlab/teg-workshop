package util

import "fmt"

// AlphaHex adds and alpha value to a 6-digit hex string.
func AlphaHex(hex string, alpha int) string {
	return fmt.Sprintf("#%2d%s", alpha, hex[1:])
}
