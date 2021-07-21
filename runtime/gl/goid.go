/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package gl

import (
	"github.com/petermattis/goid"
	"strconv"
)

func getGoId() (string, bool) {
	return strconv.FormatInt(goid.Get(), 10), true
}
