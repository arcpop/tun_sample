// +build linux

package main

import (
    "github.com/arcpop/tun"
	"net"
	"fmt"
)

type simpleForwarding struct {
    TunAdapter tun.TunInterface
    RemoteConn *net.UDPConn
    close chan interface{}
}


func NewSimpleForwarding(tunName string, localAddr, remoteAddr *net.UDPAddr) (*simpleForwarding, error) {
    var s simpleForwarding
    var err error
    
    s.TunAdapter, err = tun.New(tunName)
    if err != nil {
        return nil, err
    }

    s.RemoteConn, err = net.DialUDP("udp", localAddr, remoteAddr)
    if err != nil {
        s.TunAdapter.Close()
        return nil, err
    }
    s.close = make(chan interface{})
    go s.ReadWorker()
    go s.WriteWorker()
    return &s, nil
}

func (s *simpleForwarding) ReadWorker() {
    for {
        select {
            case <-s.close:
                return
            default:
                buffer := make([]byte, 1600)
                n, err := s.RemoteConn.Read(buffer)
                if err != nil {
                    fmt.Println("ReadWorker: " + err.Error())
                    continue
                }
                _, err = s.TunAdapter.Write(buffer[:n])
                if err != nil {
                    fmt.Println("ReadWorker: " + err.Error())
                    continue
                }
        }
    }
}
func (s *simpleForwarding) WriteWorker() {
    for {
        select {
            case <-s.close:
                return
            default:
                buffer := make([]byte, 1600)
                n, err := s.TunAdapter.Read(buffer)
                if err != nil {
                    fmt.Println("ReadWorker: " + err.Error())
                    continue
                }
                _, err = s.RemoteConn.Write(buffer[:n])
                if err != nil {
                    fmt.Println("ReadWorker: " + err.Error())
                    continue
                }
        }
    }
}
func (s *simpleForwarding) Close() error {
    s.TunAdapter.Close()
    s.RemoteConn.Close()
    s.close <- nil
    s.close <- nil
    return nil
}

func main()  {
    addr1, err := net.ResolveUDPAddr("udp", "192.168.56.1:5001")
    if err != nil {
        fmt.Println("ResolveUDPAddr", err)
        return
    }

    addr2, err := net.ResolveUDPAddr("udp", "192.168.56.100:5002")  
    if err != nil {
        fmt.Println("ResolveUDPAddr", err)
        return
    }

    s2, err := NewSimpleForwarding("", addr2, addr1)
    if err != nil {
        fmt.Println("NewSimpleForwarding", err)
        return
    } 
    defer s2.Close()

    err = s2.TunAdapter.SetIPAddress(
        net.IP{192, 168, 100, 3}, 
        net.IP{192, 168, 100, 255},
        net.IP{255, 255, 255, 0})
    if err != nil {
        fmt.Println("SetIPAddress", err)
        return
    }

    addr4, err := net.ResolveTCPAddr("tcp", "192.168.100.3:6002")  
    if err != nil {
        fmt.Println("ResolveTCPAddr", err)
        return
    } 
    
    listener, err := net.ListenTCP("tcp", addr4)
    defer listener.Close()
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Accept", err)
            return
        } 
        buffer := make([]byte, 2048)
        n, err := conn.Read(buffer)
        msg := string(buffer[:n])
        fmt.Println("Received: " + msg)
        msg += " Hello World 2!"
        conn.Write([]byte(msg))
        conn.Close()
    }
}
