// Package secretary provides methods for ciphering.
package secretary

// Secretary defines a set of methods for types implementing Secretary.
type Secretary interface {
	Encode(data string) string
	Decode(msg string) (string, error)
}
