package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ao-go-server/internal/service"
)

type AdminAPI struct {
	mapService   *service.MapService
	userService  *service.UserService
	loginService *service.LoginService
}

func NewAdminAPI(mapService *service.MapService, userService *service.UserService, loginService *service.LoginService) *AdminAPI {
	return &AdminAPI{
		mapService:   mapService,
		userService:  userService,
		loginService: loginService,
	}
}

func (a *AdminAPI) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/world/list", a.handleWorldList)
	mux.HandleFunc("/world/load", a.handleWorldLoad)
	mux.HandleFunc("/world/unload", a.handleWorldUnload)
	mux.HandleFunc("/world/reload", a.handleWorldReload)

	mux.HandleFunc("/conn/list", a.handleConnList)
	mux.HandleFunc("/conn/count", a.handleConnCount)
	mux.HandleFunc("/conn/kick", a.handleConnKick)

	mux.HandleFunc("/account/lock", a.handleAccountLock)
	mux.HandleFunc("/account/unlock", a.handleAccountUnlock)
	mux.HandleFunc("/account/reset-password", a.handleAccountResetPassword)

	mux.HandleFunc("/player/teleport", a.handlePlayerTeleport)
	mux.HandleFunc("/player/save", a.handlePlayerSave)

	fmt.Printf("Admin API listening on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func (a *AdminAPI) handleAccountLock(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	if err := a.loginService.LockAccount(nick); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Account %s locked", nick)
}

func (a *AdminAPI) handleAccountUnlock(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	if err := a.loginService.UnlockAccount(nick); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Account %s unlocked", nick)
}

func (a *AdminAPI) handleAccountResetPassword(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	newPass := r.URL.Query().Get("password")
	if newPass == "" {
		newPass = "123456" // Default if not provided
	}
	if err := a.loginService.ResetPassword(nick, newPass); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Password for %s reset to %s", nick, newPass)
}

func (a *AdminAPI) handlePlayerTeleport(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	m, _ := strconv.Atoi(r.URL.Query().Get("map"))
	x, _ := strconv.Atoi(r.URL.Query().Get("x"))
	y, _ := strconv.Atoi(r.URL.Query().Get("y"))

	if err := a.loginService.TeleportPlayer(nick, m, x, y); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Player %s teleported to %d,%d,%d", nick, m, x, y)
}

func (a *AdminAPI) handlePlayerSave(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	all := r.URL.Query().Get("all") == "true"

	if all {
		a.loginService.SaveAllPlayers()
		fmt.Fprintf(w, "All players saved")
	} else {
		if err := a.loginService.SavePlayer(nick); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Player %s saved", nick)
	}
}

func (a *AdminAPI) handleConnList(w http.ResponseWriter, r *http.Request) {
	conns := a.userService.GetLoggedConnections()
	list := make([]map[string]string, 0, len(conns))
	for _, c := range conns {
		name := "Unknown"
		if u := c.GetUser(); u != nil {
			name = u.Name
		}
		list = append(list, map[string]string{
			"addr": c.GetRemoteAddr(),
			"user": name,
		})
	}
	json.NewEncoder(w).Encode(list)
}

func (a *AdminAPI) handleConnCount(w http.ResponseWriter, r *http.Request) {
	conns := a.userService.GetLoggedConnections()
	fmt.Fprintf(w, "%d", len(conns))
}

func (a *AdminAPI) handleConnKick(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	ip := r.URL.Query().Get("ip")

	if name != "" {
		if a.userService.KickByName(name) {
			fmt.Fprintf(w, "Kicked user %s", name)
		} else {
			http.Error(w, "User not found", http.StatusNotFound)
		}
		return
	}

	if ip != "" {
		kicked := a.userService.KickByIP(ip)
		fmt.Fprintf(w, "Kicked %d connections from IP %s", kicked, ip)
		return
	}

	http.Error(w, "Missing name or ip parameter", http.StatusBadRequest)
}

func (a *AdminAPI) handleWorldList(w http.ResponseWriter, r *http.Request) {
	maps := a.mapService.GetLoadedMaps()
	json.NewEncoder(w).Encode(maps)
}

func (a *AdminAPI) handleWorldLoad(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid map ID", http.StatusBadRequest)
		return
	}

	if err := a.mapService.LoadMap(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Map %d loaded", id)
}

func (a *AdminAPI) handleWorldUnload(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid map ID", http.StatusBadRequest)
		return
	}

	a.mapService.UnloadMap(id)
	fmt.Fprintf(w, "Map %d unloaded", id)
}

func (a *AdminAPI) handleWorldReload(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid map ID", http.StatusBadRequest)
		return
	}

	if err := a.mapService.ReloadMap(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Map %d reloaded", id)
}
