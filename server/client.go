package server

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/martini"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	PREPROCCESS_NONE   = iota
	PREPROCCESS_COFFEE = iota
	PREPROCCESS_LESS   = iota
	PREPROCCESS_UGLY   = iota
)

type ClientServer struct {
	conf  ServerConfig
	m     *martini.ClassicMartini
	start time.Time
}

func NewClientServer(conf ServerConfig) *ClientServer {
	c := &ClientServer{}
	c.conf = conf
	c.m = martini.Classic()
	c.start = time.Now()
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

func (c *ClientServer) preprocessFile(cmd *exec.Cmd, f io.Reader) ([]byte, error) {
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
							return out, nil
						} else {
							return stde, fmt.Errorf("Preprocessor failed.")
						}
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (c *ClientServer) clientAssets(directory string) martini.Handler {
	dir := http.Dir(directory)
	ca_dir := http.Dir(c.conf.CompiledAssetPath)
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
		} else {
			if strings.HasSuffix(file, ".js") && !strings.HasSuffix(file, "min.js") {
				if c.conf.UglifyPath != "" && c.conf.Mode == MODE_PRODUCTION {
					processor = PREPROCCESS_UGLY
				}
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

		if ca_dir != "" {
			if processor != PREPROCCESS_NONE && c.conf.Mode == MODE_PRODUCTION {
				nf, err := ca_dir.Open(file)
				if err == nil {
					defer nf.Close()
					nfi, err := nf.Stat()
					if err == nil {
						if nfi.ModTime().After(fi.ModTime()) {
							f = nf
							fi = nfi
							processor = PREPROCCESS_NONE
						}
					}
				}
			}
		}

		if processor != PREPROCCESS_NONE {
			var cmd *exec.Cmd
			if processor == PREPROCCESS_COFFEE {
				cmd = exec.Command(c.conf.CoffeePath, "-sc")
			} else if processor == PREPROCCESS_LESS {
				cmd = exec.Command(c.conf.LessPath, "-")
			} else if processor == PREPROCCESS_UGLY {
				cmd = exec.Command(c.conf.UglifyPath, "-")
			} else {
				panic("Invalid Preprocessor in clientAssets.")
			}
			if out, err := c.preprocessFile(cmd, f); err == nil {
				mod := fi.ModTime()
				if c.conf.Mode == MODE_DEVELOPMENT {
					if c.start.After(mod) {
						mod = c.start
					}
				}
				if processor == PREPROCCESS_COFFEE && c.conf.Mode == MODE_PRODUCTION {
					out, err = c.preprocessFile(exec.Command(c.conf.UglifyPath, "-"), bytes.NewReader(out))
					if err != nil {
						if out == nil {
							res.WriteHeader(http.StatusInternalServerError)
						} else {
							res.WriteHeader(http.StatusInternalServerError)
							res.Write(out)
						}
					}
				}
				http.ServeContent(res, req, file, mod, bytes.NewReader(out))
				if ca_dir != "" && c.conf.Mode == MODE_PRODUCTION {
					os.MkdirAll(path.Dir(path.Join(string(ca_dir), file)), 0755)
					if rf, err := os.Create(path.Join(string(ca_dir), file)); err == nil {
						rf.Write(out)
					}
				}
				return
			} else {
				if out == nil {
					res.WriteHeader(http.StatusInternalServerError)
				} else {
					res.WriteHeader(http.StatusInternalServerError)
					res.Write(out)
				}
				return
			}

			//http.ServeContent(res, req, file, fi.ModTime(), f)
		} else {
			mod := fi.ModTime()
			if c.conf.Mode == MODE_DEVELOPMENT {
				if c.start.After(mod) {
					mod = c.start
				}
			}
			http.ServeContent(res, req, file, mod, f)
		}
	}
}
