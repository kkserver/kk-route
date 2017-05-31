package main

import (
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/inifile"
	"github.com/kkserver/kk-lib/kk/json"
	"github.com/kkserver/kk-lua/lua"
	"log"
	"os"
	"strings"
	"time"
)

const VERSION = "1.0.0"

func NewMessageFunction(L *lua.State) int {

	var v = kk.Message{}

	var n = L.GetTop()

	if n > 0 {
		v.Method = L.ToString(-n)
	}

	if n > 1 {
		v.From = L.ToString(-n + 1)
	}

	if n > 2 {
		v.To = L.ToString(-n + 2)
	}

	if n > 3 {
		v.Type = L.ToString(-n + 3)
	}

	if n > 4 {
		v.Content = []byte(L.ToString(-n + 4))
	}

	L.PushObject(&v)

	return 1
}

func OpenLibs(L *lua.State) {

	L.NewTable()

	L.PushString("NewMessage")
	L.PushFunction(NewMessageFunction)

	L.RawSet(-3)

	L.SetGlobal("kk")

}

type Route struct {
	Name           string
	Address        string
	Remote         string
	MaxConnections int
	LuaFile        string
	Ping           string
}

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	log.Printf("VERSION: %s\n", VERSION)

	env := "./config/env.ini"

	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	var route = Route{}

	err := inifile.DecodeFile(&route, "./app.ini")

	if err != nil {
		log.Panicln(err)
	}

	err = inifile.DecodeFile(&route, env)

	if err != nil {
		log.Panicln(err)
	}

	var L *lua.State = nil

	defer L.Close()

	{
		if route.LuaFile != "" {

			L = lua.NewState()

			defer L.Close()

			L.Openlibs()

			OpenLibs(L)

			if L.LoadFile(route.LuaFile) == 0 {

				if L.Call(0, 0) != 0 {
					log.Panicln(L.ToString(-1))
					L.Pop(1)
				}

			} else {
				log.Panicln(L.ToString(-1))
				L.Pop(1)
			}

		}
	}

	var remote *kk.TCPClient = nil
	var server *kk.TCPServer = nil
	var remote_connect func() = nil
	var server_start func() = nil

	remote_connect = func() {

		log.Println("remote connect ...")

		remote = kk.NewTCPClient(route.Name, route.Remote, nil)

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

	if route.Remote != "" {
		remote_connect()
	}

	server_start = func() {

		log.Println("server start ...")

		server = kk.NewTCPServer(route.Name, route.Address, route.MaxConnections)

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

		server.OnConnected = func(client *kk.TCPClient) {
			log.Println("connected: " + client.Address() + " name:" + client.Name())

			if route.Ping != "" {

				var v = kk.Message{}
				v.Method = "PING"
				v.From = client.Name()
				v.To = route.Ping
				v.Type = "text/json"
				v.Content, _ = json.Encode(map[string]interface{}{"address": client.Address(), "options": client.Options()})

				kk.GetDispatchMain().Async(func() {
					server.OnMessage(&v, nil)
				})

			}

		}

		server.OnDisconnected = func(client *kk.TCPClient, err error) {
			log.Println("disconnected: " + client.Address() + " error: " + err.Error())

			if route.Ping != "" {
				var v = kk.Message{}
				v.Method = "PING.DISCONNECTED"
				v.From = client.Name()
				v.To = route.Ping
				v.Type = "text/json"
				v.Content, _ = json.Encode(map[string]interface{}{"address": client.Address(), "options": client.Options()})

				kk.GetDispatchMain().Async(func() {
					server.OnMessage(&v, nil)
				})

			}

		}

		server.OnMessage = func(message *kk.Message, from kk.INeuron) {
			if strings.HasPrefix(message.To, route.Name) {
				server.Send(message, from)
			} else if remote != nil {
				if L != nil {
					L.GetGlobal("OnMessage")
					if L.IsFunction(-1) {
						L.PushObject(message)
						L.PushObject(from)
						L.PushObject(server)
						if L.Call(3, 1) != 0 {
							log.Println("[FAIL][LUA] ", L.ToString(-1))
							var m = kk.Message{"DONE", route.Name, message.From, "text", []byte(message.Method)}
							from.Send(&m, nil)
						} else if L.IsBoolean(-1) && L.ToBoolean(-1) {
							remote.Send(message, nil)
						} else {
							var m = kk.Message{"DONE", route.Name, message.From, "text", []byte(message.Method)}
							from.Send(&m, nil)
						}
					}
					L.Pop(L.GetTop())
				} else {
					remote.Send(message, nil)
				}
			}
		}
	}

	if route.Address != "" {
		server_start()
	}

	kk.DispatchMain()

}
