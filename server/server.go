package server

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"text/template"
	"time"
)

type Info struct {
	Path    string
	Name    string
	Size    string
	ModTime string
}

type LIST struct {
	Pre   string
	Path  string
	Dirs  []Info
	Files []Info
}

type servHandler int8

func server(conn io.Writer, dirPath string) {
	aimDir := new(LIST)
	aimDir.Path = dirPath + "/"
	aimDir.Pre, _ = path.Split(dirPath)
	dp, err := os.ReadDir(dirPath)
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range dp {
		if v.IsDir() {
			fs, err := os.Stat(dirPath + "/" + v.Name())
			if err != nil {
				log.Println(err)
				os.Exit(0)
			}
			var modTime string
			name := fs.Name()
			if len(name) > 50 {
				name = formatTail(fs.Name(), 50)
				modTime = fs.ModTime().Format(time.UnixDate)
			} else {
				modTime = formatHead(fs.ModTime().Format(time.UnixDate), caclEmpty(fs.Name(), 50))
			}
			size := formatHead(strconv.FormatInt(fs.Size(), 10), 20)
			aimDir.Dirs = append(aimDir.Dirs, Info{fs.Name(), name, size, modTime})
			if err != nil {
				log.Println(err)
				os.Exit(0)
			}
		} else {
			fs, err := os.Stat(dirPath + "/" + v.Name())
			if err != nil {
				log.Println(err)
				os.Exit(0)
			}
			var modTime string
			name := fs.Name()
			if len(name) > 50 {
				name = formatTail(fs.Name(), 50)
				modTime = fs.ModTime().Format(time.UnixDate)
			} else {
				modTime = formatHead(fs.ModTime().Format(time.UnixDate), caclEmpty(fs.Name(), 50))
			}
			size := formatHead(strconv.FormatInt(fs.Size(), 10), 20)
			aimDir.Files = append(aimDir.Files, Info{fs.Name(), name, size, modTime})
			if err != nil {
				log.Println(err)
				os.Exit(0)
			}
		}
	}
	tmp, err := template.ParseFiles("./templates/template1.tmpl")
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	err = tmp.Execute(conn, aimDir)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
}

func (h servHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// log.Println(r.RemoteAddr)
	path := path.Clean(r.URL.Path)
	if path == "/favicon.ico" {
		return
	}
	i, err := os.Stat(path)
	if err != nil {
		log.Println(err)
		efp, err := os.Open("./templates/404.html")
		if err != nil {
			log.Println(err)
			return
		}
		io.Copy(w, efp)
		return
	}
	if i.IsDir() {
		if r.URL.Path[len(r.URL.Path)-1] != '/' {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
		}
		server(w, path)
	} else {
		fp, err := os.Open(path)
		if err != nil {
			log.Println(err)
			efp, err := os.Open("./templates/404.html")
			if err != nil {
				log.Println(err)
				return
			}
			io.Copy(w, efp)
			return
		}
		io.Copy(w, fp)
		fp.Close()
	}
}

func RunServer() {
	var handler servHandler = 0
	err := http.ListenAndServe(":10086", handler)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
}
