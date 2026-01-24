package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/service"
)

type AdminAPI struct {
	mapService   *service.MapService
	userService  *service.UserService
	loginService *service.LoginService
	npcService   *service.NpcService
	aiService    *service.AIService
	config       *config.Config
}

func NewAdminAPI(mapService *service.MapService, userService *service.UserService, loginService *service.LoginService, npcService *service.NpcService, aiService *service.AIService, cfg *config.Config) *AdminAPI {
	return &AdminAPI{
		mapService:   mapService,
		userService:  userService,
		loginService: loginService,
		npcService:   npcService,
		aiService:    aiService,
		config:       cfg,
	}
}

func (a *AdminAPI) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/world/list", a.handleWorldList)
	mux.HandleFunc("/world/load", a.handleWorldLoad)
	mux.HandleFunc("/world/unload", a.handleWorldUnload)
	mux.HandleFunc("/world/reload", a.handleWorldReload)
	mux.HandleFunc("/world/save", a.handleWorldSave)

	mux.HandleFunc("/conn/list", a.handleConnList)
	mux.HandleFunc("/conn/count", a.handleConnCount)
	mux.HandleFunc("/conn/kick", a.handleConnKick)
	mux.HandleFunc("/conn/ban", a.handleConnBan)
	mux.HandleFunc("/conn/unban", a.handleConnUnban)

	mux.HandleFunc("/account/lock", a.handleAccountLock)
	mux.HandleFunc("/account/unlock", a.handleAccountUnlock)
	mux.HandleFunc("/account/reset-password", a.handleAccountResetPassword)

	mux.HandleFunc("/player/teleport", a.handlePlayerTeleport)
	mux.HandleFunc("/player/save", a.handlePlayerSave)
	mux.HandleFunc("/player/info", a.handlePlayerInfo)

	mux.HandleFunc("/npc/reload", a.handleNpcReload)
	mux.HandleFunc("/npc/disable", a.handleNpcDisable)
	mux.HandleFunc("/npc/enable", a.handleNpcEnable)
	mux.HandleFunc("/npc/respawn", a.handleNpcRespawn)
	mux.HandleFunc("/npc/list", a.handleNpcList)

	mux.HandleFunc("/config/get", a.handleConfigGet)
	mux.HandleFunc("/config/set", a.handleConfigSet)
	mux.HandleFunc("/config/list", a.handleConfigList)

	fmt.Printf("Admin API listening on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func (a *AdminAPI) handleConfigList(w http.ResponseWriter, r *http.Request) {
	keys := []map[string]string{
		{"key": "version", "description": "Server version (read-only)", "type": "string"},
		{"key": "max_users", "description": "Maximum concurrent users", "type": "int"},
		{"key": "creation_enabled", "description": "Allow new character creation", "type": "bool"},
		{"key": "admins_only", "description": "Restrict access to admins only", "type": "bool"},
		{"key": "interval_attack", "description": "Interval between attacks (ms)", "type": "int64"},
		{"key": "interval_spell", "description": "Interval between spells (ms)", "type": "int64"},
	}
	json.NewEncoder(w).Encode(keys)
}

func (a *AdminAPI) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	var val interface{}

	switch key {
	case "version":
		val = a.config.Version
	case "max_users":
		val = a.config.MaxConcurrentUsers
	case "creation_enabled":
		val = a.config.CharacterCreationEnabled
	case "admins_only":
		val = a.config.RestrictedToAdmins
	case "interval_attack":
		val = a.config.IntervalAttack
	case "interval_spell":
		val = a.config.IntervalSpell
	default:
		http.Error(w, "Unknown config key", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%v", val)
}

func (a *AdminAPI) handleConnBan(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	if nick == "" {
		http.Error(w, "Missing nick", http.StatusBadRequest)
		return
	}

	if err := a.loginService.LockAccount(nick); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kick if online
	a.userService.KickByName(nick)

	fmt.Fprintf(w, "Account %s banned and kicked", nick)
}

func (a *AdminAPI) handleConnUnban(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	if nick == "" {
		http.Error(w, "Missing nick", http.StatusBadRequest)
		return
	}

	if err := a.loginService.UnlockAccount(nick); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Account %s unbanned", nick)
}


func (a *AdminAPI) handleNpcList(w http.ResponseWriter, r *http.Request) {
	npcs := a.npcService.GetWorldNpcs()
	list := make([]map[string]interface{}, 0, len(npcs))
	for _, npc := range npcs {
		list = append(list, map[string]interface{}{
			"index": npc.Index,
			"id":    npc.NPC.ID,
			"name":  npc.NPC.Name,
			"map":   npc.Position.Map,
			"x":     npc.Position.X,
			"y":     npc.Position.Y,
			"hp":    npc.HP,
		})
	}
	json.NewEncoder(w).Encode(list)
}

func (a *AdminAPI) handleNpcReload(w http.ResponseWriter, r *http.Request) {
	if err := a.npcService.LoadNpcs(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "NPCs reloaded")
}

func (a *AdminAPI) handleNpcDisable(w http.ResponseWriter, r *http.Request) {
	a.aiService.SetEnabled(false)
	fmt.Fprintf(w, "AI disabled")
}

func (a *AdminAPI) handleNpcEnable(w http.ResponseWriter, r *http.Request) {
	a.aiService.SetEnabled(true)
	fmt.Fprintf(w, "AI enabled")
}

func (a *AdminAPI) handleNpcRespawn(w http.ResponseWriter, r *http.Request) {
	mapIDStr := r.URL.Query().Get("map")
	if mapIDStr == "" {
		http.Error(w, "Missing map parameter", http.StatusBadRequest)
		return
	}

	mapID, err := strconv.Atoi(mapIDStr)
	if err != nil {
		http.Error(w, "Invalid map ID", http.StatusBadRequest)
		return
	}

	// Respawning NPCs on a map. 
	// For simplicity, we just reload the map entities if it's loaded.
	m := a.mapService.GetMap(mapID)
	if m == nil {
		http.Error(w, "Map not loaded", http.StatusNotFound)
		return
	}

	// Remove all NPCs from map first
	m.Lock()
	npcsToRemove := make([]*model.WorldNPC, 0, len(m.Npcs))
	for _, npc := range m.Npcs {
		npcsToRemove = append(npcsToRemove, npc)
	}
	m.Unlock()

	for _, npc := range npcsToRemove {
		// We use a temporary flag to avoid auto-respawn during manual respawn command
		oldRespawn := npc.Respawn
		npc.Respawn = false
		a.npcService.RemoveNPC(npc, a.mapService)
		npc.Respawn = oldRespawn
	}

	// Re-resolve entities to spawn them back from map file
	a.mapService.ReloadMap(mapID)

	fmt.Fprintf(w, "NPCs respawned on map %d", mapID)
}

func (a *AdminAPI) handleWorldSave(w http.ResponseWriter, r *http.Request) {
	a.mapService.SaveCache()
	fmt.Fprintf(w, "World cache saved")
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

func (a *AdminAPI) handlePlayerInfo(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	if nick == "" {
		http.Error(w, "Missing nick", http.StatusBadRequest)
		return
	}

	char := a.userService.GetCharacterByName(nick)
	if char == nil {
		http.Error(w, "Player not online", http.StatusNotFound)
		return
	}

	info := map[string]interface{}{
		"name":      char.Name,
		"level":     char.Level,
		"exp":       char.Exp,
		"hp":        char.Hp,
		"max_hp":    char.MaxHp,
		"mana":      char.Mana,
		"max_mana":  char.MaxMana,
		"gold":      char.Gold,
		"map":       char.Position.Map,
		"x":         char.Position.X,
		"y":         char.Position.Y,
		"archetype": char.Archetype,
	}
	json.NewEncoder(w).Encode(info)
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

func (a *AdminAPI) handleConfigSet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("value")

	switch key {
	case "max_users":
		if i, err := strconv.Atoi(val); err == nil {
			a.config.MaxConcurrentUsers = i
		}
	case "creation_enabled":
		a.config.CharacterCreationEnabled = val == "true"
	case "admins_only":
		a.config.RestrictedToAdmins = val == "true"
	case "interval_attack":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.config.IntervalAttack = i
		}
	case "interval_spell":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.config.IntervalSpell = i
		}
	default:
		http.Error(w, "Unknown or read-only config key", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Config %s set to %s", key, val)
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

