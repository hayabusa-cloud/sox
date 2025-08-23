## sox (WIP)

Sox is an asynchronous socket I/O and event notification library.  
It can also be used as a tool library for networking, event handling,  
message packaging, etc.

### Basic Concepts
* Low-copy I/O implementation for TCP, UDP, SCTP and Unix domain sockets
* Low kernel-userspace context switch implementation for event notifications
* Compatible with low-lock programming

### Environment Requirements
Currently **only** supports **Linux** systems.   
Kernel version must be **6.12** or later.   
The recommended distribution is Debian 13 (Trixie) or later.   
You can compile a custom kernel for WSL2 as a development environment.   
For detailed instructions, please visit: [WSL2-Linux-Kernel](https://github.com/microsoft/WSL2-Linux-Kernel)

### License
©2022 Hayabusa Cloud Co., Ltd.  
#5F Eclat BLDG, 3-6-2 Shibuya, Shibuya City, Tokyo 150-0002, Japan  
Released under the MIT license