package main

import (
	"github.com/kkserver/kk-lib/kk"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const VERSION = "1.0.0"

func help() {
	log.Println("kk-route <name> [--local 0.0.0.0:8080] [--remote 127.0.0.1:8090] [--maxconnections 204800]")
}

func main() {

	var args = os.Args
	var name string = ""
	var remoteAddr string = ""
	var localAddr string = ""
	var maxconnections int = 204800

	if len(args) > 1 {
		name = args[1]
	} else {
		help()
		return
	}

	var i = 2

	for i < len(args) {

		if args[i] == "--local" {
			if i+1 < len(args) {
				localAddr = args[i+1]
				i += 1
			} else {
				break
			}
		} else if args[i] == "--remote" {
			if i+1 < len(args) {
				remoteAddr = args[i+1]
				i += 1
			} else {
				break
			}
		} else if args[i] == "--maxconnections" {
			if i+1 < len(args) {
				maxconnections, _ = strconv.Atoi(args[i+1])
				i += 1
			} else {
				break
			}
		}

		i += 1
	}

	var remote *kk.TCPClient = nil
	var server *kk.TCPServer = nil
	var remote_connect func() = nil
	var server_start func() = nil

	remote_connect = func() {
		log.Println("remote connect ...")
		remote = kk.NewTCPClient(name, remoteAddr, nil)
		remote.OnConnected = func() {
			log.Println("remote: " + remote.Address())
		}
		remote.OnDisconnected = func(err error) {
			log.Println("remote disconnected: " + remote.Address())
			kk.GetDispatchMain().AsyncDelay(remote_connect, time.Second)
		}
		remote.OnMessage = func(message *kk.Message) {
			if server != nil {
				server.Send(message, remote)
			}
		}
	}

	if remoteAddr != "" {
		remote_connect()
	}

	server_start = func() {

		log.Println("server start ...")

		server = kk.NewTCPServer(name, localAddr, maxconnections)

		server.OnStart = func() {
			log.Println(server.Address())
		}

		server.OnFail = func(err error) {
			log.Println("server error: " + err.Error())
			kk.GetDispatchMain().AsyncDelay(server_start, time.Second)
		}

		server.OnAccept = func(client *kk.TCPClient) {
			log.Println("accept: " + client.Address())
		}

		server.OnDisconnected = func(client *kk.TCPClient, err error) {
			log.Println("disconnected: " + client.Address() + " error: " + err.Error())
		}

		server.OnMessage = func(message *kk.Message, from kk.INeuron) {
			if strings.HasPrefix(message.To, name) {
				server.Send(message, from)
			} else if remote != nil {
				remote.Send(message, nil)
			}
		}
	}

	if localAddr != "" {
		server_start()
	}

	kk.DispatchMain()

}
