package mbase

import (
	"strconv"
	"sync/atomic"

	"github.com/google/uuid"
)

var (
	BootID = uuid.New().String()
)

// SessionNameGenerator used to generate session name automatically,
// you can replace it
var SessionNameGenerator = PrefixedSessionNameGenerator(BootID[:24])

func UsePrefixedSessionName(prefix string) {
	SessionNameGenerator = PrefixedSessionNameGenerator(prefix + BootID[:24])
}

func PrefixedSessionNameGenerator(header string) func(...string) string {
	var counter uint64
	return func(prefix ...string) string {
		// return bootid related sequential name
		const h = "000000000000"
		var seq = atomic.AddUint64(&counter, 1)
		var s = strconv.FormatUint(seq, 10)
		if len(prefix) > 0 {
			return prefix[0] + header + h[:12-len(s)] + s
		} else {
			return header + h[:12-len(s)] + s
		}
	}
}
