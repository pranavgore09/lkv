package server

import (
	"log"
	"net"
	"strconv"
	"syscall"

	"github.com/pranavgore09/lkv/config"
	"github.com/pranavgore09/lkv/core"
)

func AsyncRun() error {
	var connection_str string = config.Host + ":" + strconv.Itoa(config.Port)

	log.Println("Config: ", connection_str)

	max_clients := 20000

	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)

	if err != nil {
		return err
	}

	defer syscall.Close(serverFD)

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	ip4 := net.ParseIP(config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	if err := syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	// ASYNC MODE ON ->
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer syscall.Close(epollFD)

	var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		return err
	}

	con_clients := 0

	log.Println("Ok1")
	for {

		nEvents, e := syscall.EpollWait(epollFD, events[:], -1)
		if e != nil {
			continue
		}

		for i := 0; i < nEvents; i++ {

			if int(events[i].Fd) == serverFD {
				// server IO, accept connection
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("Error:", err)
					continue
				}

				con_clients++
				syscall.SetNonblock(serverFD, true)

				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}
				if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &socketClientEvent); err != nil {
					log.Fatal(err)
				}

			} else {
				// client send some command
				int_fd := int(events[i].Fd)
				comm := core.FDComm{Fd: int_fd}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int_fd)
					con_clients--
					continue
				}
				sendResponse(comm, cmd)
			}

		}

	}
}
