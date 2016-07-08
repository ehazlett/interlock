// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

<<<<<<< HEAD
<<<<<<< HEAD
=======
// +build windows
>>>>>>> c73b1ae... switch to engine-api; update beacon to be more efficient
=======
>>>>>>> 12a5469... start on swarm services; move to glade
// +build go1.4

package windows

import "syscall"

func Unsetenv(key string) error {
	// This was added in Go 1.4.
	return syscall.Unsetenv(key)
}
