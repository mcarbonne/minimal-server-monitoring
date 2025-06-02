package utils

import "fmt"

func SafeClose(closer interface{ Close() error }) {
	if err := closer.Close(); err != nil {
		panic(fmt.Errorf("failed to close file: %v", err))
	}
}
