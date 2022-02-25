package server

import (
	"bytes"
	"fileBrowser/auth"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"
)

const (
	FAVICON     = "/favicon.ico"
	UNIXFAVICON = "./static/favicon.ico"
	// WINFAVICON = ".\\static\\favicon.ico"
	UNIXNOTFOUND = "./static/404.html"
	// WINNOTFOUND = ".\\static\\404.html"
	UNIXTEMPLATE1 = "./templates/template1.tmpl"
	// WINTEMPLATE1 = ".\\templates\\template1.tmpl"
	UNIXLOGIN = "./static/login.html"
	// WINLOGIN   = ".\\static\\login.html"
	LOGIN_PATH = "/login"
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

type ServHandler struct {
	Root       string
	Pre        string
	UserOnline *auth.USER_NODE
}

func notFound(w http.ResponseWriter) {
	fp, err := os.Open(UNIXNOTFOUND)
	if err != nil {
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	io.Copy(w, fp)
	fp.Close()
}

func loginPage(w http.ResponseWriter) {
	fp, err := os.Open(UNIXLOGIN)
	if err != nil {
		log.Println(err)
		return
	}
	io.Copy(w, fp)
	fp.Close()
}

func clearCookie(w http.ResponseWriter, cookie *http.Cookie) {
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

func server(conn http.ResponseWriter, dirPath string, root string) {
	aimDir := new(LIST)
	aimDir.Path = dirPath + "/"
	// aimDir.Path = dirPath + "\\"
	preDir, _ := filepath.Split(dirPath)
	if len(preDir) > len(root) {
		aimDir.Pre = preDir[len(root):]
	} else {
		aimDir.Pre = ""
	}
	dp, err := os.ReadDir(dirPath)
	if err != nil {
		log.Println(err)
		notFound(conn)
		return
	}
	for _, v := range dp {
		if v.IsDir() {
			fs, err := os.Stat(filepath.Join(dirPath, v.Name()))
			if err != nil {
				log.Println(err)
				notFound(conn)
				return
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
				notFound(conn)
				return
			}
		} else {
			fs, err := os.Stat(filepath.Join(dirPath, v.Name()))
			if err != nil {
				log.Println(err)
				notFound(conn)
				return
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
				notFound(conn)
				return
			}
		}
	}
	tmp, err := template.ParseFiles(UNIXTEMPLATE1)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	err = tmp.Execute(conn, aimDir)
	if err != nil {
		log.Println(err)
		notFound(conn)
		return
	}
}

func (h ServHandler) login(w http.ResponseWriter, r *http.Request) int8 {
	buf := make([]byte, 64)
	r.Body.Read(buf)
	ui := bytes.IndexByte(buf, '=')
	m := bytes.IndexByte(buf, '&')
	pi := bytes.LastIndexByte(buf, '=')
	end := bytes.IndexByte(buf, 0)
	if m == -1 || ui == -1 || ui >= m || pi <= m {
		log.Println("Invalid post value")
		return -1
	}
	userName := string(buf[ui+1 : m])
	if userName == "root" {
		log.Println("Invalid username")
		return -1
	}
	passWord := buf[pi+1 : end]
	cookie, loginFlag, _ := h.UserOnline.Login(userName, passWord)
	switch loginFlag {
	case auth.LOGIN_SUCCESS:
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusFound)
		w.Write([]byte("Login success"))
		log.Println("Login success")
	case auth.LOGIN_NO_USER:
		log.Println("Invalid user name")
	case auth.LOGIN_WRONG_PASSWD:
		log.Println("Wrong password")
	default:
		log.Println("Undefined mistake")
	}
	return loginFlag
}

func (h ServHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == FAVICON {
		fp, err := os.Open(UNIXFAVICON)
		if err != nil {
			log.Println(err)
			return
		}
		io.Copy(w, fp)
		return
	}
	reqCookie := r.Cookies()
	fmt.Println(reqCookie)
	if len(reqCookie) == 0 {
		if r.URL.Path == LOGIN_PATH {
			isLogin := h.login(w, r)
			switch isLogin {
			case auth.LOGIN_SUCCESS:
				return
			case auth.LOGIN_NO_USER:
				loginPage(w)
				w.Write([]byte("Invalid user name"))
				return
			case auth.LOGIN_WRONG_PASSWD:
				loginPage(w)
				w.Write([]byte("Wrong password"))
				return
			default:
				notFound(w)
				w.Write([]byte("Undefined mistake"))
				return
			}
		} else {
			loginPage(w)
			return
		}
	}
	cookieFlag := h.UserOnline.Check(reqCookie[0].Name, reqCookie[0].Value)
	switch cookieFlag {
	case auth.COOKIE_GOOD:
	case auth.COOKIE_NOTFOUND:
		clearCookie(w, reqCookie[0])
		return
	case auth.COOKIE_EXPIRED:
		clearCookie(w, reqCookie[0])
		loginPage(w)
		return
	case auth.UNDEFINED_WRONG:
		notFound(w)
		return
	default:
		notFound(w)
		return
	}
	servPath := filepath.Join(h.Root, filepath.Clean(r.URL.Path))
	if h.Pre == servPath {
		return
	}
	i, err := os.Stat(servPath)
	if err != nil {
		log.Println(err)
		notFound(w)
		return
	}
	if i.IsDir() {
		if r.URL.Path[len(r.URL.Path)-1] != '/' {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
		} else {
			server(w, servPath, h.Root)
			return
		}
	} else {
		fp, err := os.Open(servPath)
		if err != nil {
			log.Println(err)
			notFound(w)
			return
		}
		fs, err := fp.Stat()
		if err != nil {
			log.Println(err)
			notFound(w)
			return
		}
		w.Header().Add("content-length", strconv.FormatInt(fs.Size(), 10))
		io.Copy(w, fp)
		fp.Close()
	}
}

func RunServer(root string, serv *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	err := serv.ListenAndServe()
	if err != nil {
		log.Println(err)
		return
	}
}
