# go-winio

<<<<<<< HEAD
<<<<<<< HEAD
This repository contains utilities for efficiently performing Win32 IO operations in
Go. Currently, this is focused on accessing named pipes and other file handles, and
for using named pipes as a net transport.

This code relies on IO completion ports to avoid blocking IO on system threads, allowing Go
to reuse the thread to schedule another goroutine. This limits support to Windows Vista and
=======
This repository contains utilities for efficiently performing Win32 IO operations in 
Go. Currently, this is focused on accessing named pipes and other file handles, and
for using named pipes as a net transport.

This code relies on IO completion ports to avoid blocking IO on system threads, allowing Go 
to reuse the thread to schedule another goroutine. This limits support to Windows Vista and 
>>>>>>> c73b1ae... switch to engine-api; update beacon to be more efficient
=======
This repository contains utilities for efficiently performing Win32 IO operations in
Go. Currently, this is focused on accessing named pipes and other file handles, and
for using named pipes as a net transport.

This code relies on IO completion ports to avoid blocking IO on system threads, allowing Go
to reuse the thread to schedule another goroutine. This limits support to Windows Vista and
>>>>>>> 12a5469... start on swarm services; move to glade
newer operating systems. This is similar to the implementation of network sockets in Go's net
package.

Please see the LICENSE file for licensing information.

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> 12a5469... start on swarm services; move to glade
This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/). For more information
see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.

Thanks to natefinch for the inspiration for this library. See https://github.com/natefinch/npipe
<<<<<<< HEAD
=======
Thanks to natefinch for the inspiration for this library. See https://github.com/natefinch/npipe 
>>>>>>> c73b1ae... switch to engine-api; update beacon to be more efficient
=======
>>>>>>> 12a5469... start on swarm services; move to glade
for another named pipe implementation.
