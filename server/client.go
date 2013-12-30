package server

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/martini"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path"
	"strings"
)

const (
	PREPROCCESS_NONE   = iota
	PREPROCCESS_COFFEE = iota
	PREPROCCESS_LESS   = iota
)

type ClientServer struct {
	conf ServerConfig
	m    *martini.ClassicMartini
}

func NewClientServer(conf ServerConfig) *ClientServer {
	c := &ClientServer{}
	c.conf = conf
	c.m = martini.Classic()
	c.routes()

	return c
}

func (c *ClientServer) Serve() {
	log.Println("[Client] Starting client server on", fmt.Sprintf("%s:%d", c.conf.HostName, c.conf.ClientPort))
	if e := http.ListenAndServe(fmt.Sprintf("%s:%d", c.conf.HostName, c.conf.ClientPort), c.m); e != nil {
		log.Println("[Client] Failed to start client server on", fmt.Sprintf("%s:%d", c.conf.HostName, c.conf.ClientPort), e)
	}
}

func (c *ClientServer) routes() {
	c.m.Use(c.clientAssets(c.conf.ClientAssets))
}

func (c *ClientServer) clientAssets(directory string) martini.Handler {
	dir := http.Dir(directory)
	return func(res http.ResponseWriter, req *http.Request, log *log.Logger) {
		file := req.URL.Path
		processor := PREPROCCESS_NONE
		f, err := dir.Open(file)
		if err != nil {
			if strings.HasSuffix(file, ".js") {
				coffeeFile := file + ".coffee"
				f, err = dir.Open(coffeeFile)
				processor = PREPROCCESS_COFFEE
				if err != nil {
					return
				}
			} else if strings.HasSuffix(file, ".css") {
				lessFile := file + ".less"
				f, err = dir.Open(lessFile)
				processor = PREPROCCESS_LESS
				if err != nil {
					return
				}
			} else {
				return
			}
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return
		}

		// Try to serve index.html
		if fi.IsDir() {

			// redirect if missing trailing slash
			if !strings.HasSuffix(file, "/") {
				http.Redirect(res, req, file+"/", http.StatusFound)
				return
			}

			file = path.Join(file, "index.html")
			f, err = dir.Open(file)
			if err != nil {
				return
			}
			defer f.Close()

			fi, err = f.Stat()
			if err != nil || fi.IsDir() {
				return
			}
		}

		log.Println("[Static] Serving " + file)
		if processor != PREPROCCESS_NONE {
			var cmd *exec.Cmd
			if processor == PREPROCCESS_COFFEE {
				cmd = exec.Command("coffee", "-sc")
			} else if processor == PREPROCCESS_LESS {
				cmd = exec.Command("lessc", "-")
			} else {
				panic("Invalid Preprocessor in clientAssets.")
			}
			if stdout, err := cmd.StdoutPipe(); err == nil {
				if stdin, err := cmd.StdinPipe(); err == nil {
					if stderr, err := cmd.StderrPipe(); err == nil {
						if err := cmd.Start(); err == nil {
							if b, err := ioutil.ReadAll(f); err == nil {
								stdin.Write(b)
								stdin.Close()
								out, _ := ioutil.ReadAll(stdout)
								stde, _ := ioutil.ReadAll(stderr)
								if e := cmd.Wait(); e == nil {
									http.ServeContent(res, req, file, fi.ModTime(), bytes.NewReader(out))
									return
								} else {
									res.WriteHeader(http.StatusInternalServerError)
									//http.ServeContent(res, req, file, fi.ModTime(), bytes.NewReader(stde))
									res.Write(stde)
									return
								}
							}
						}
					}
				}
			}
			res.WriteHeader(http.StatusInternalServerError)
			//http.ServeContent(res, req, file, fi.ModTime(), f)
		} else {
			http.ServeContent(res, req, file, fi.ModTime(), f)
		}
	}
}
