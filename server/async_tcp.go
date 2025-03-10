package server

import (
	"log"
	"net"
	"syscall"
	"time"

	"github.com/diceclone/config"
	"github.com/diceclone/core"
)

// var connectedClients int = 0
var cronFrequency time.Duration = 1 * time.Second
var lastCronExectime time.Time = time.Now()

func RunAsyncTCPServer(host string, port int) error {
	log.Println("starting asynchronous TCP server on ", config.Host, config.Port)

	// we are dealing with low level socket connection
	// first create a server socket that is bound to host and port, the connection should be asychronous
	// monitor the serverFD through epoll event
	// if there is new event, it can be of 2 types
	// a new connection request -> received on serverFD -> create a new socket and add it to epoll/ kqueue for monitoring
	// data bytes on existing socket -> received on a client FD

	connectedClients := 0
	maxClients := 10000

	// create a non blocking socket with maximus connections, bind it to ipv4 address and port
	serverFD, err := createServerSocket(maxClients)
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	// do async I/O
	// create a kernel event queue to track events on registered FDs, equivalent to EPOLL_CREATE
	epollFD, err := syscall.Kqueue()
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollFD)

	// ask the kernel event queue to monitor READ event on the server socket
	socketServerEvent := syscall.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
	}

	// equivalent of EPOLL_CTL
	_, err = syscall.Kevent(epollFD, []syscall.Kevent_t{socketServerEvent}, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	var events []syscall.Kevent_t = make([]syscall.Kevent_t, maxClients)

	for {

		// every one min, run a check on the keys to delete the expired keys
		// take 20 keys at one time
		// if more than 25% of the sampled entries have expired, then there are lot of stale items in the cache
		// run the check again on the next 20 keys
		if time.Now().After(lastCronExectime.Add(cronFrequency)) {
			core.SafeDeleteExpiredKeys()
			lastCronExectime = time.Now()
		}

		// equivalent of EPOLL_WAIT
		nEvents, err := syscall.Kevent(epollFD, nil, events[:], nil)
		if err != nil {
			continue
		}

		// there are two possibilities - either a new connection or data on an existing connection
		for i := 0; i < nEvents; i++ {
			if int(events[i].Ident) == serverFD {
				// new connection
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}
				connectedClients += 1
				syscall.SetNonblock(fd, true)
				socketClientEvent := syscall.Kevent_t{
					Ident:  uint64(fd),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
				}

				// equivalent of EPOLL_CTL; register the new client FD with epoll
				_, err = syscall.Kevent(epollFD, []syscall.Kevent_t{socketClientEvent}, nil, nil)
				if err != nil {
					log.Fatal(err)
				}

			} else {
				// data on an existing connection
				comm := core.FDComm{Fd: int(events[i].Ident)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Ident))
					connectedClients -= 1
					continue
				}
				respond(comm, cmd)
			}
		}
	}

}

func createServerSocket(connections int) (int, error) {
	// create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Println("error creating socket", err)
		return -1, err
	}

	// set the socket to operate in non blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return -1, err
	}

	// bind the socket to the host and port
	ip4 := net.ParseIP(config.Host)
	err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]}},
	)
	if err != nil {
		return -1, err
	}

	// listen on the socket
	err = syscall.Listen(serverFD, connections)
	if err != nil {
		return -1, err
	}

	return serverFD, nil
}
