package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"syscall"
)

func main() {
	var err error

	args := flag.Args()
	port := 8080
	if len(args) > 1 {
		port, err = strconv.Atoi(args[0])
		if err != nil {
			panic(err)
		}
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	file, err := ln.(*net.TCPListener).File()
	if err != nil {
		panic(err)
	}
	fdLn := int(file.Fd())

	var readfds syscall.FdSet
	fdMax := fdLn
	fmt.Printf("Server FD: %d\n", fdLn)
	connMap := make(map[int]net.Conn)
	for {
		// After each syscall.Select, readfds set only contains the file descriptors that are currently ready for reading
		// The file descriptors that are not ready for reading are cleared from the set
		// So, we need to add the server file descriptor and all the client file descriptors to the set before each syscall.Select
		FDZero(&readfds)
		FDSet(fdLn, &readfds)
		for fd := range connMap {
			FDSet(fd, &readfds)
		}

		numReady, err := syscall.Select(fdMax+1, &readfds, nil, nil, nil)
		if err != nil {
			// handle occasional "interrupted system call" error
			if err == syscall.EINTR {
				continue
			} else {
				panic(err)
			}
		}

		if numReady == 0 {
			continue
		}

		for fd := 0; fd < fdMax+1; fd++ {
			if FDIsSet(fd, &readfds) {
				if fd == fdLn {
					conn, err := ln.Accept()
					if err != nil {
						fmt.Println(err)
						continue
					}

					connFile, err := conn.(*net.TCPConn).File()
					if err != nil {
						fmt.Println(err)
						conn.Close()
						continue
					}
					fdConn := int(connFile.Fd())
					connMap[fdConn] = conn
					FDSet(fdConn, &readfds)
					if fdConn > fdMax {
						fdMax = fdConn
					}

					fmt.Printf("%s %s connected\n", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
					fmt.Printf("Connection FD: %d\n", fdConn)
				} else {
					conn := connMap[fd]
					buf := make([]byte, 1024)

					n, err := conn.Read(buf)
					if err != nil || n == 0 {
						if err != nil {
							fmt.Println(err)
						}
						FDClr(fd, &readfds)
						delete(connMap, fd)
						conn.Close()
						fmt.Printf("%s %s hung up\n", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
					} else {
						fmt.Println(string(buf[:n]))
					}
				}
			}
		}

	}
}

// FDZero set to zero the fdSet
func FDZero(p *syscall.FdSet) {
	p.Bits = [16]int64{}
}

// FDSet set a fd of fdSet
func FDSet(fd int, p *syscall.FdSet) {
	// p.Bits[fd/32] |= (1 << (uint(fd) % 32))
	p.Bits[fd/64] |= 1 << (fd % 64)
}

// FDClr clear a fd of fdSet
func FDClr(fd int, p *syscall.FdSet) {
	// p.Bits[fd/32] &^= (1 << (uint(fd) % 32))
	p.Bits[fd/64] &^= 1 << (fd % 64)
}

// FDIsSet return true if fd is set
func FDIsSet(fd int, p *syscall.FdSet) bool {
	// return p.Bits[fd/32]&(1<<(uint(fd)%32)) != 0
	return p.Bits[fd/64]&(1<<(fd%64)) != 0
}
