//go:build !unsafe

package deno

var (
	additionalDenoArgs = []string{"--no-remote", "--allow-env", "--allow-hrtime"}
)
