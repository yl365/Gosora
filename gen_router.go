// Code generated by. DO NOT EDIT.
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main

import "log"
import "strings"
import "sync"
import "errors"
import "net/http"

var ErrNoRoute = errors.New("That route doesn't exist.")

type GenRouter struct {
	UploadHandler func(http.ResponseWriter, *http.Request)
	extra_routes map[string]func(http.ResponseWriter, *http.Request, User)
	
	sync.RWMutex
}

func NewGenRouter(uploads http.Handler) *GenRouter {
	return &GenRouter{
		UploadHandler: http.StripPrefix("/uploads/",uploads).ServeHTTP,
		extra_routes: make(map[string]func(http.ResponseWriter, *http.Request, User)),
	}
}

func (router *GenRouter) Handle(_ string, _ http.Handler) {
}

func (router *GenRouter) HandleFunc(pattern string, handle func(http.ResponseWriter, *http.Request, User)) {
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
	
	switch(prefix) {
		case "/api":
			routeAPI(w,req,user)
			return
		case "/overview":
			routeOverview(w,req,user)
			return
		case "/forums":
			routeForums(w,req,user)
			return
		case "/forum":
			routeForum(w,req,user,extra_data)
			return
		case "/theme":
			routeChangeTheme(w,req,user)
			return
		case "/report":
			switch(req.URL.Path) {
				case "/report/submit/":
					routeReportSubmit(w,req,user,extra_data)
					return
			}
		case "/topics":
			switch(req.URL.Path) {
				case "/topics/create/":
					routeTopicCreate(w,req,user,extra_data)
					return
				default:
					routeTopics(w,req,user)
					return
			}
		case "/panel":
			switch(req.URL.Path) {
				case "/panel/forums/":
					routePanelForums(w,req,user)
					return
				case "/panel/forums/create/":
					routePanelForumsCreateSubmit(w,req,user)
					return
				case "/panel/forums/delete/":
					routePanelForumsDelete(w,req,user,extra_data)
					return
				case "/panel/forums/delete/submit/":
					routePanelForumsDeleteSubmit(w,req,user,extra_data)
					return
				case "/panel/forums/edit/":
					routePanelForumsEdit(w,req,user,extra_data)
					return
				case "/panel/forums/edit/submit/":
					routePanelForumsEditSubmit(w,req,user,extra_data)
					return
				case "/panel/forums/edit/perms/submit/":
					routePanelForumsEditPermsSubmit(w,req,user,extra_data)
					return
				case "/panel/settings/":
					routePanelSettings(w,req,user)
					return
				case "/panel/settings/edit/":
					routePanelSetting(w,req,user,extra_data)
					return
				case "/panel/settings/edit/submit/":
					routePanelSettingEdit(w,req,user,extra_data)
					return
				case "/panel/settings/word-filters/":
					routePanelWordFilters(w,req,user)
					return
				case "/panel/settings/word-filters/create/":
					routePanelWordFiltersCreate(w,req,user)
					return
				case "/panel/settings/word-filters/edit/":
					routePanelWordFiltersEdit(w,req,user,extra_data)
					return
				case "/panel/settings/word-filters/edit/submit/":
					routePanelWordFiltersEditSubmit(w,req,user,extra_data)
					return
				case "/panel/settings/word-filters/delete/submit/":
					routePanelWordFiltersDeleteSubmit(w,req,user,extra_data)
					return
				case "/panel/themes/":
					routePanelThemes(w,req,user)
					return
				case "/panel/themes/default/":
					routePanelThemesSetDefault(w,req,user,extra_data)
					return
				case "/panel/plugins/":
					routePanelPlugins(w,req,user)
					return
				case "/panel/plugins/activate/":
					routePanelPluginsActivate(w,req,user,extra_data)
					return
				case "/panel/plugins/deactivate/":
					routePanelPluginsDeactivate(w,req,user,extra_data)
					return
				case "/panel/plugins/install/":
					routePanelPluginsInstall(w,req,user,extra_data)
					return
				case "/panel/users/":
					routePanelUsers(w,req,user)
					return
				case "/panel/users/edit/":
					routePanelUsersEdit(w,req,user,extra_data)
					return
				case "/panel/users/edit/submit/":
					routePanelUsersEditSubmit(w,req,user,extra_data)
					return
				case "/panel/groups/":
					routePanelGroups(w,req,user)
					return
				case "/panel/groups/edit/":
					routePanelGroupsEdit(w,req,user,extra_data)
					return
				case "/panel/groups/edit/perms/":
					routePanelGroupsEditPerms(w,req,user,extra_data)
					return
				case "/panel/groups/edit/submit/":
					routePanelGroupsEditSubmit(w,req,user,extra_data)
					return
				case "/panel/groups/edit/perms/submit/":
					routePanelGroupsEditPermsSubmit(w,req,user,extra_data)
					return
				case "/panel/groups/create/":
					routePanelGroupsCreateSubmit(w,req,user)
					return
				case "/panel/backups/":
					routePanelBackups(w,req,user,extra_data)
					return
				case "/panel/logs/mod/":
					routePanelLogsMod(w,req,user)
					return
				case "/panel/debug/":
					routePanelDebug(w,req,user)
					return
				default:
					routePanel(w,req,user)
					return
			}
		case "/uploads":
			if extra_data == "" {
				NotFound(w,req)
				return
			}
			req.URL.Path += extra_data
			router.UploadHandler(w,req)
			return
		case "":
			// Stop the favicons, robots.txt file, etc. resolving to the topics list
			// TODO: Add support for favicons and robots.txt files
			switch(extra_data) {
				case "robots.txt":
					routeRobotsTxt(w,req)
					return
			}
			
			if extra_data != "" {
				NotFound(w,req)
				return
			}
			config.DefaultRoute(w,req,user)
			return
		//default: NotFound(w,req)
	}
	
	// A fallback for the routes which haven't been converted to the new router yet or plugins
	router.RLock()
	handle, ok := router.extra_routes[req.URL.Path]
	router.RUnlock()
	
	if ok {
		req.URL.Path += extra_data
		handle(w,req,user)
		return
	}
	NotFound(w,req)
}
