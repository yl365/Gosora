/* WIP Under Construction */
package main

import "log"

//import "strings"
import "os"

var routeList []*RouteImpl
var routeGroups []*RouteGroup

func main() {
	log.Println("Generating the router...")

	// Load all the routes...
	routes()

	var out string
	var fileData = "// Code generated by. DO NOT EDIT.\n/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */\n"

	for _, route := range routeList {
		var end int
		if route.Path[len(route.Path)-1] == '/' {
			end = len(route.Path) - 1
		} else {
			end = len(route.Path) - 1
		}
		out += "\n\t\tcase \"" + route.Path[0:end] + "\":"
		if route.Before != "" {
			out += "\n\t\t\t" + route.Before
		}
		out += "\n\t\t\terr = " + route.Name + "(w,req,user"
		for _, item := range route.Vars {
			out += "," + item
		}
		out += `)
			if err != nil {
				router.handleError(err,w,req,user)
			}`
	}

	for _, group := range routeGroups {
		var end int
		if group.Path[len(group.Path)-1] == '/' {
			end = len(group.Path) - 1
		} else {
			end = len(group.Path) - 1
		}
		out += `
		case "` + group.Path[0:end] + `":`
		for _, callback := range group.Before {
			out += `
			err = ` + callback + `(w,req,user)
			if err != nil {
				router.handleError(err,w,req,user)
				return
			}
			`
		}
		out += "\n\t\t\tswitch(req.URL.Path) {"

		var defaultRoute = blankRoute()
		for _, route := range group.RouteList {
			if group.Path == route.Path {
				defaultRoute = route
				continue
			}

			out += "\n\t\t\t\tcase \"" + route.Path + "\":"
			if route.Before != "" {
				out += "\n\t\t\t\t\t" + route.Before
			}
			out += "\n\t\t\t\t\terr = " + route.Name + "(w,req,user"
			for _, item := range route.Vars {
				out += "," + item
			}
			out += ")"
		}

		if defaultRoute.Name != "" {
			out += "\n\t\t\t\tdefault:"
			if defaultRoute.Before != "" {
				out += "\n\t\t\t\t\t" + defaultRoute.Before
			}
			out += "\n\t\t\t\t\terr = " + defaultRoute.Name + "(w,req,user"
			for _, item := range defaultRoute.Vars {
				out += ", " + item
			}
			out += ")"
		}
		out += `
			}
			if err != nil {
				router.handleError(err,w,req,user)
			}`
	}

	fileData += `package main

import "log"
import "strings"
import "sync"
import "errors"
import "net/http"

var ErrNoRoute = errors.New("That route doesn't exist.")

type GenRouter struct {
	UploadHandler func(http.ResponseWriter, *http.Request)
	extra_routes map[string]func(http.ResponseWriter, *http.Request, User) RouteError
	
	sync.RWMutex
}

func NewGenRouter(uploads http.Handler) *GenRouter {
	return &GenRouter{
		UploadHandler: http.StripPrefix("/uploads/",uploads).ServeHTTP,
		extra_routes: make(map[string]func(http.ResponseWriter, *http.Request, User) RouteError),
	}
}

func (router *GenRouter) handleError(err RouteError, w http.ResponseWriter, r *http.Request, user User) {
	if err.Handled() {
		return
	}
	
	if err.Type() == "system" {
		InternalErrorJSQ(err,w,r,err.Json())
		return
	}
	LocalErrorJSQ(err.Error(),w,r,user,err.Json())
}

func (router *GenRouter) Handle(_ string, _ http.Handler) {
}

func (router *GenRouter) HandleFunc(pattern string, handle func(http.ResponseWriter, *http.Request, User) RouteError) {
	router.Lock()
	router.extra_routes[pattern] = handle
	router.Unlock()
}

func (router *GenRouter) RemoveFunc(pattern string) error {
	router.Lock()
	_, ok := router.extra_routes[pattern]
	if !ok {
		router.Unlock()
		return ErrNoRoute
	}
	delete(router.extra_routes,pattern)
	router.Unlock()
	return nil
}

func (router *GenRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//if req.URL.Path == "/" {
	//	default_route(w,req)
	//	return
	//}
	if len(req.URL.Path) == 0 || req.URL.Path[0] != '/' {
		w.WriteHeader(405)
		w.Write([]byte(""))
		return
	}
	
	var prefix, extra_data string
	prefix = req.URL.Path[0:strings.IndexByte(req.URL.Path[1:],'/') + 1]
	if req.URL.Path[len(req.URL.Path) - 1] != '/' {
		extra_data = req.URL.Path[strings.LastIndexByte(req.URL.Path,'/') + 1:]
		req.URL.Path = req.URL.Path[:strings.LastIndexByte(req.URL.Path,'/') + 1]
	}
	
	if dev.SuperDebug {
		log.Print("before routeStatic")
		log.Print("prefix: ", prefix)
		log.Print("req.URL.Path: ", req.URL.Path)
		log.Print("extra_data: ", extra_data)
		log.Print("req.Referer(): ", req.Referer())
	}
	
	if prefix == "/static" {
		req.URL.Path += extra_data
		routeStatic(w,req)
		return
	}
	
	if dev.SuperDebug {
		log.Print("before PreRoute")
	}
	
	// Deal with the session stuff, etc.
	user, ok := PreRoute(w,req)
	if !ok {
		return
	}
	
	if dev.SuperDebug {
		log.Print("after PreRoute")
	}
	
	var err RouteError
	switch(prefix) {` + out + `
		case "/uploads":
			if extra_data == "" {
				NotFound(w,req)
				return
			}
			req.URL.Path += extra_data
			// TODO: Find a way to propagate errors up from this?
			router.UploadHandler(w,req)
		case "":
			// Stop the favicons, robots.txt file, etc. resolving to the topics list
			// TODO: Add support for favicons and robots.txt files
			switch(extra_data) {
				case "robots.txt":
					err = routeRobotsTxt(w,req)
					if err != nil {
						router.handleError(err,w,req,user)
					}
					return
			}
			
			if extra_data != "" {
				NotFound(w,req)
				return
			}
			config.DefaultRoute(w,req,user)
		default:
			// A fallback for the routes which haven't been converted to the new router yet or plugins
			router.RLock()
			handle, ok := router.extra_routes[req.URL.Path]
			router.RUnlock()
			
			if ok {
				req.URL.Path += extra_data
				err = handle(w,req,user)
				if err != nil {
					router.handleError(err,w,req,user)
				}
				return
			}
			NotFound(w,req)
	}
}
`
	writeFile("./gen_router.go", fileData)
	log.Println("Successfully generated the router")
}

func writeFile(name string, content string) {
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
	f.Sync()
	f.Close()
}
