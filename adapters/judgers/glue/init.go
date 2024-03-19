//go:build unsafe

package glue

import "github.com/fedstackjs/azukiiro/judge"

func init() {
	judge.RegisterAdapter(&GlueAdapter{})
}
