package redis

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	hashmap "slashing/redis/HashMap"
	"slashing/utils"

	"github.com/tidwall/redcon"
)

func ListenAndServeRedisServer(addr string) error {
	var items *hashmap.HashMap //"Lockless"
	var ps redcon.PubSub

	fileDir := utils.CacheDir("cache-redis")
	fileName := "/kv.db"

	data, err := ioutil.ReadFile(filepath.Join(".", fileDir, fileName))
	if err != nil {
		items = hashmap.NewFromBinary(data)
	} else {
		items = hashmap.New()
	}
	err = redcon.ListenAndServe(addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			switch strings.ToLower(string(cmd.Args[0])) {
			default:
				conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
			case "ping":
				conn.WriteString("PONG")
			case "quit":
				conn.WriteString("OK")
				conn.Close()
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				items.Set(string(cmd.Args[1]), cmd.Args[2])
				conn.WriteString("OK")
			case "mset":
				ks := []interface{}{}
				vs := []interface{}{}
				for i := 2; i < len(cmd.Args); i += 2 {
					ks = append(ks, cmd.Args[i])
					vs = append(vs, cmd.Args[i+1])
				}
				items.MSet(ks, vs)
				conn.WriteString("OK")
			case "mget":
				conn.WriteArray(len(cmd.Args) - 1)
				for i := 1; i < len(cmd.Args); i++ {
					data, _ := items.Get(cmd.Args[i])
					conn.WriteBulk(data.([]byte))
				}

			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				val, ok := items.Get(cmd.Args[1])
				if !ok {
					conn.WriteNull()
				} else {
					conn.WriteBulk(val.([]byte))
				}
			case "del":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				items.Del(string(cmd.Args[1]))
				conn.WriteString("OK")
			case "save":
				data, _ := items.ToBinary()
				utils.FilePutContents(data, fileDir, fileName)

			case "publish":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				conn.WriteInt(ps.Publish(string(cmd.Args[1]), string(cmd.Args[2])))
			case "subscribe", "psubscribe":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				command := strings.ToLower(string(cmd.Args[0]))
				for i := 1; i < len(cmd.Args); i++ {
					if command == "psubscribe" {
						ps.Psubscribe(conn, string(cmd.Args[i]))
					} else {
						ps.Subscribe(conn, string(cmd.Args[i]))
					}
				}
			}

		},
		func(conn redcon.Conn) bool {
			// Use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// This is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	)
	return err
}
