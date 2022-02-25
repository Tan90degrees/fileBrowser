package main

import (
	"context"
	"fileBrowser/auth"
	"fileBrowser/server"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	fmt.Println(`
 ___  ___  _______   ___       ___       ________  ___
|\  \|\  \|\  ___ \ |\  \     |\  \     |\   __  \|\  \
\ \  \\\  \ \   __/|\ \  \    \ \  \    \ \  \|\  \ \  \
 \ \   __  \ \  \_|/_\ \  \    \ \  \    \ \  \\\  \ \  \
  \ \  \ \  \ \  \_|\ \ \  \____\ \  \____\ \  \\\  \ \__\
   \ \__\ \__\ \_______\ \_______\ \_______\ \_______\|__|
    \|__|\|__|\|_______|\|_______|\|_______|\|_______|   ___
                                                        |\__\
                                                        \|__|`)
	if len(os.Args) != 2 {
		log.Fatal("Usage : fileBrowser[.exe] [path]")
	}
	root := os.Args[1]
	stat, err := os.Stat(root)
	if err != nil || !stat.IsDir() {
		log.Fatal("\"", os.Args[1], "\" is not a directory.")
	}
	var alive uint8
	alive = 0
	var cmd string
	var wg sync.WaitGroup
	preRoot, _ := filepath.Split(root)
	serv := &http.Server{
		Addr:    ":10086",
		Handler: server.ServHandler{Root: root, Pre: preRoot, UserOnline: auth.InitUserList()},
	}
	for {
		fmt.Print(">>>")
		fmt.Scanln(&cmd)
		switch cmd {
		case "start":
			if alive == 1 {
				fmt.Println("Server has already started")
				break
			}
			wg.Add(1)
			go server.RunServer(filepath.Clean(root), serv, &wg)
			alive = 1
			fmt.Println("Server start")
		case "shutdown":
			if err = serv.Shutdown(context.TODO()); err != nil {
				log.Println("Can not shutdown server")
				break
			}
			wg.Wait()
			alive = 0
			fmt.Println("Server shutdown")
			serv = &http.Server{
				Addr:    ":10086",
				Handler: server.ServHandler{Root: root, Pre: preRoot, UserOnline: auth.InitUserList()},
			}
		case "register":
			err = auth.RegUser()
			if err != nil {
				log.Println(err)
				break
			}
			fmt.Println("Registered")
		case "exit":
			if alive == 0 {
				os.Exit(0)
			}
			if err = serv.Shutdown(context.TODO()); err != nil {
				log.Println("Can not shutdown server")
				break
			}
			wg.Wait()
			fmt.Println("Server shutdown")
			os.Exit(0)
		default:
			fmt.Print(">>>")
		}
	}
}
