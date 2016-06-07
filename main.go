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
    addr1, err := net.ResolveUDPAddr("udp", "localhost:5001")
    if err != nil {
        fmt.Println(err)
        return
    }
    addr2, err := net.ResolveUDPAddr("udp", "localhost:5002")  
    if err != nil {
        fmt.Println(err)
        return
    }  

    s1, err := NewSimpleForwarding("", addr1, addr2)
    if err != nil {
        fmt.Println(err)
        return
    } 
    defer s1.Close()
    s2, err := NewSimpleForwarding("", addr2, addr1)
    if err != nil {
        fmt.Println(err)
        return
    } 
    defer s2.Close()

    s1.TunAdapter.Write([]byte("Hello World!"))
    buf := make([]byte, 2048)
    s2.TunAdapter.Read(buf)
    fmt.Println(string(buf))
}
