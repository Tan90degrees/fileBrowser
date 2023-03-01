package server

import (
	"bytes"
	"fileBrowser/auth"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"
)

const (
	FAVICON_REQ_PATH  = "/favicon.ico"
	LOGIN_REQ_PATH    = "/login"
	TEMPLATE_PATH     = "./templates/template1.tmpl"
	FAVICON_FILE_PATH = "./static/favicon.ico"
	LOGIN_HTML_PATH   = "./static/login.html"
)

var HttpStatusHtmlPath [600]string

type Info struct {
	Path    string
	Name    string
	Size    string
	ModTime string
}

type LIST struct {
	Root  string
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

func initHttpStatusHtmlPath() {
	HttpStatusHtmlPath[http.StatusProcessing] = "./static/102.html"
	HttpStatusHtmlPath[http.StatusBadRequest] = "./static/400.html"
	HttpStatusHtmlPath[http.StatusNotFound] = "./static/404.html"
	HttpStatusHtmlPath[http.StatusBadGateway] = "./static/502.html"
}

func simpleRequest(w http.ResponseWriter, code uint) error {
	if code > 599 {
		return fmt.Errorf(fmt.Sprintf("status code(%v) is too large", code))
	}
	if HttpStatusHtmlPath[code] != "" {
		fp, err := os.Open(HttpStatusHtmlPath[code])
		if err != nil {
			log.Println(err)
			return fmt.Errorf("file of status code(%v) is not found", code)
		}
		w.WriteHeader(http.StatusBadRequest)
		io.Copy(w, fp)
		fp.Close()
		return nil
	} else {
		return fmt.Errorf(fmt.Sprintf("status code(%v) is not supported", code))
	}
}

func loginPage(w http.ResponseWriter) {
	fp, err := os.Open(LOGIN_HTML_PATH)
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

func servGetDir(conn http.ResponseWriter, dirPath string, root string) {
	aimDir := new(LIST)
	aimDir.Root = "/"
	aimDir.Path = dirPath + "/"
	preDir, _ := filepath.Split(dirPath)
	if len(preDir) > len(root) {
		aimDir.Pre = preDir[len(root):]
	} else {
		aimDir.Pre = ""
	}
	dp, err := os.ReadDir(dirPath)
	if err != nil {
		log.Println(err)
		err = simpleRequest(conn, http.StatusNotFound)
		if err != nil {
			log.Println(err)
		}
		return
	}
	for _, v := range dp {
		if v.IsDir() {
			fi, err := os.Stat(filepath.Join(dirPath, v.Name()))
			if err != nil {
				log.Println(err)
				err = simpleRequest(conn, http.StatusNotFound)
				if err != nil {
					log.Println(err)
				}
				return
			}
			var modTime string
			name := fi.Name()
			if len(name) > 50 {
				name = formatTail(fi.Name(), 50)
				modTime = fi.ModTime().Format(time.UnixDate)
			} else {
				modTime = formatHead(fi.ModTime().Format(time.UnixDate), caclEmpty(fi.Name(), 50))
			}
			size := formatHead(strconv.FormatInt(fi.Size(), 10), 20)
			aimDir.Dirs = append(aimDir.Dirs, Info{fi.Name(), name, size, modTime})
		} else {
			fi, err := os.Stat(filepath.Join(dirPath, v.Name()))
			if err != nil {
				log.Println(err)
				err = simpleRequest(conn, http.StatusNotFound)
				if err != nil {
					log.Println(err)
				}
				return
			}
			var modTime string
			name := fi.Name()
			if len(name) > 50 {
				name = formatTail(fi.Name(), 50)
				modTime = fi.ModTime().Format(time.UnixDate)
			} else {
				modTime = formatHead(fi.ModTime().Format(time.UnixDate), caclEmpty(fi.Name(), 50))
			}
			size := formatHead(strconv.FormatInt(fi.Size(), 10), 20)
			aimDir.Files = append(aimDir.Files, Info{fi.Name(), name, size, modTime})
		}
	}
	tmp, err := template.ParseFiles(TEMPLATE_PATH)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	err = tmp.Execute(conn, aimDir)
	if err != nil {
		log.Println(err)
		err = simpleRequest(conn, http.StatusNotFound)
		if err != nil {
			log.Println(err)
		}
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

func servPost(h *ServHandler, w http.ResponseWriter, r *http.Request, servPath string) {
	err := checkPath(h.Pre, servPath)
	if err != nil {
		log.Println(err)
		err = simpleRequest(w, http.StatusBadRequest)
		if err != nil {
			log.Println(err)
		}
		return
	}
	i, err := os.Stat(servPath)
	if err != nil {
		log.Println(err)
		err = simpleRequest(w, http.StatusNotFound)
		if err != nil {
			log.Println(err)
		}
		return
	}
	if !i.IsDir() {
		err = simpleRequest(w, http.StatusBadRequest)
		if err != nil {
			log.Println(err)
		}
		return
	}
	mr, err := r.MultipartReader()
	if err != nil {
		log.Println(err)
		return
	}

	if !diskCapacityEnough(servPath, uint64(r.ContentLength)) {
		log.Println("Insufficient disk space")
		err = simpleRequest(w, http.StatusBadRequest)
		if err != nil {
			log.Println(err)
		}
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		formName := part.FormName()
		if formName == "" {
			continue
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		fPath := filepath.Join(servPath, fileName)
		log.Println(fPath)
		_, err = os.Stat(fPath)
		if err == nil {
			log.Println("File already exist")
			continue
		}
		fp, err := os.OpenFile(fPath, os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Println(err)
			err = simpleRequest(w, http.StatusBadGateway)
			if err != nil {
				log.Println(err)
			}
			return
		}
		io.Copy(fp, part)
		fp.Close()
	}
}

func servGet(h *ServHandler, w http.ResponseWriter, r *http.Request, servPath string) {
	err := checkPath(h.Pre, servPath)
	if err != nil {
		log.Println(err)
		return
	}
	fi, err := os.Stat(servPath)
	if err != nil {
		log.Println(err)
		err = simpleRequest(w, http.StatusNotFound)
		if err != nil {
			log.Println(err)
		}
		return
	}
	if fi.IsDir() {
		servGetDir(w, servPath, h.Root)
	} else {
		fp, err := os.Open(servPath)
		if err != nil {
			log.Println(err)
			err = simpleRequest(w, http.StatusNotFound)
			if err != nil {
				log.Println(err)
			}
			return
		}
		fi, err := fp.Stat()
		if err != nil {
			log.Println(err)
			err = simpleRequest(w, http.StatusNotFound)
			if err != nil {
				log.Println(err)
			}
			return
		}
		w.Header().Add("content-length", strconv.FormatInt(fi.Size(), 10))
		io.Copy(w, fp)
		fp.Close()
	}
}

func (h ServHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqPath := path.Clean(r.URL.Path)
	if reqPath == FAVICON_REQ_PATH {
		fp, err := os.Open(FAVICON_FILE_PATH)
		if err != nil {
			log.Println(err)
			return
		}
		io.Copy(w, fp)
		return
	}
	reqCookie := r.Cookies()
	if len(reqCookie) == 0 {
		if reqPath == LOGIN_REQ_PATH {
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
				err := simpleRequest(w, http.StatusNotFound)
				if err != nil {
					log.Println(err)
				}
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
		http.Redirect(w, r, "/", http.StatusFound)
		return
	case auth.COOKIE_EXPIRED:
		clearCookie(w, reqCookie[0])
		http.Redirect(w, r, "/", http.StatusFound)
		return
	case auth.UNDEFINED_WRONG:
		err := simpleRequest(w, http.StatusNotFound)
		if err != nil {
			log.Println(err)
		}
		return
	default:
		err := simpleRequest(w, http.StatusNotFound)
		if err != nil {
			log.Println(err)
		}
		return
	}
	if r.Method == http.MethodGet {
		servGet(&h, w, r, filepath.Join(h.Root, reqPath))
	} else if r.Method == http.MethodPost {
		servPost(&h, w, r, filepath.Join(h.Root, reqPath))
		http.Redirect(w, r, reqPath, http.StatusFound)
	} else {
		err := simpleRequest(w, http.StatusBadRequest)
		if err != nil {
			log.Println(err)
		}
		log.Println("Unknown request method", r.Method)
	}
}

func RunServer(serv *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	initHttpStatusHtmlPath()
	err := serv.ListenAndServe()
	if err != nil {
		log.Println(err)
		return
	}
}
