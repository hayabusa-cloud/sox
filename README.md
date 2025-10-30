## sox(WIP)

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)
[![Linux Kernel](https://img.shields.io/badge/Linux-6.12%2B-FCC624?logo=linux&logoColor=black)](https://kernel.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Pure Go asynchronous network I/O library using io_uring

## Overview

Sox provides asynchronous socket I/O and event notification for Go applications on Linux, using io_uring for efficient network operations and event handling.

## Basic Concepts

- Low-copy I/O implementation on sockets  
- Reduced kernel-userspace context switches   
- Lock-free programming friendly   

## Use Cases

- Real-time communication systems   
- High-performance network servers   
- Networking utilities   

## Requirements

- **Linux only** (kernel 6.12+)
- Go 1.25+
- Recommended: Debian 13 (Trixie) or later

For WSL2 development setup, see: [WSL2-Linux-Kernel](https://github.com/microsoft/WSL2-Linux-Kernel)

## Status

⚠️ **Work in Progress** 

### Next Steps

- [ ] Completion event handling
- [ ] Connection lifecycle management
- [ ] Public API stabilization
- [ ] Built-in Middlewares 

## License

©2022 Hayabusa Cloud Co., Ltd.  
#5F Eclat BLDG, 3-6-2 Shibuya, Shibuya City, Tokyo 150-0002, Japan  
Released under the MIT license