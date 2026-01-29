package api

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type AdminAPI struct {
	mapService     service.MapService
	userService    service.UserService
	userRepo       persistence.UserRepository
	loginService   service.LoginService
	messageService service.MessageService
	npcService     service.NpcService
	aiService      service.AiService
	config         *config.Config
	globalBalance  *model.GlobalBalanceConfig
	configPath     string
	historyPath    string

	// History tracking
	mu                sync.RWMutex
	userHistoryHourly []int
	userHistoryDaily  []int
	lastHistoryUpdate time.Time
	location          *time.Location
	classDistribution map[string]int
}

func NewAdminAPI(mapService service.MapService, userService service.UserService, userRepo persistence.UserRepository, loginService service.LoginService, messageService service.MessageService, npcService service.NpcService, aiService service.AiService, cfg *config.Config, globalBalance *model.GlobalBalanceConfig, configPath string) *AdminAPI {
	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		// Fallback to FixedZone if TZ data is not available
		loc = time.FixedZone("ART", -3*60*60)
	}

	api := &AdminAPI{
		mapService:     mapService,
		userService:    userService,
		userRepo:       userRepo,
		loginService:   loginService,
		messageService: messageService,
		npcService:     npcService,
		aiService:      aiService,
		config:         cfg,
		globalBalance:  globalBalance,
		configPath:     configPath,
		historyPath:    filepath.Join(filepath.Dir(filepath.Dir(configPath)), "data", "stats_history.bin"),
		// Initialize history
		userHistoryHourly: make([]int, 24),
		userHistoryDaily:  make([]int, 0, 30),
		lastHistoryUpdate: time.Now().In(loc),
		location:          loc,
		classDistribution: make(map[string]int),
	}

	api.LoadHistory()
	api.calculateClassDistribution()
	return api
}

func (a *AdminAPI) calculateClassDistribution() {
	allChars, err := a.userRepo.GetAllCharacters()
	if err != nil {
		slog.Error("Failed to calculate class distribution", "error", err)
		return
	}

	dist := make(map[string]int)
	// Initialize all archetypes with 0
	archetypes := []model.UserArchetype{
		model.Mage, model.Cleric, model.Warrior, model.Assasin,
		model.Thief, model.Bard, model.Druid, model.Bandit,
		model.Paladin, model.Hunter, model.Worker, model.Pirate,
	}
	for _, arch := range archetypes {
		dist[archetypeToString(arch)] = 0
	}

	for _, c := range allChars {
		if c.Level >= 25 {
			typeName := archetypeToString(c.Archetype)
			dist[typeName]++
		}
	}
	a.classDistribution = dist
}

type historyData struct {
	Hourly            []int
	Daily             []int
	LastHistoryUpdate time.Time
}

func (a *AdminAPI) LoadHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.Open(a.historyPath)
	if err != nil {
		return
	}
	defer file.Close()

	var data historyData
	if err := gob.NewDecoder(file).Decode(&data); err == nil {
		a.userHistoryHourly = data.Hourly
		a.userHistoryDaily = data.Daily
		a.lastHistoryUpdate = data.LastHistoryUpdate.In(a.location)
		
		// Ensure hourly has 24 entries
		if len(a.userHistoryHourly) != 24 {
			newHourly := make([]int, 24)
			copy(newHourly, a.userHistoryHourly)
			a.userHistoryHourly = newHourly
		}
	}
}

func (a *AdminAPI) SaveHistory() {
	a.mu.RLock()
	data := historyData{
		Hourly:            a.userHistoryHourly,
		Daily:             a.userHistoryDaily,
		LastHistoryUpdate: a.lastHistoryUpdate,
	}
	a.mu.RUnlock()

	file, err := os.Create(a.historyPath)
	if err != nil {
		slog.Error("Failed to save history", "error", err)
		return
	}
	defer file.Close()

	if err := gob.NewEncoder(file).Encode(data); err != nil {
		slog.Error("Failed to encode history", "error", err)
	} else {
		slog.Info("History saved", "path", a.historyPath)
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
	mux.HandleFunc("/config/reload", a.handleConfigReload)
	
	mux.HandleFunc("/monitor/stats", a.handleMonitorStats)
	mux.HandleFunc("/monitor/charts", a.handleMonitorCharts)

	mux.HandleFunc("/event/start", a.handleEventStart)
	mux.HandleFunc("/event/stop", a.handleEventStop)
	mux.HandleFunc("/event/list", a.handleEventList)

	slog.Info("Admin API listening", "addr", addr)
	
	// Start history tracking
	go a.trackHistory()

	return http.ListenAndServe(addr, mux)
}

func (a *AdminAPI) Stop() {
	a.SaveHistory()
	a.loginService.SaveAllPlayers()
}

func (a *AdminAPI) trackHistory() {
	ticker := time.NewTicker(1 * time.Minute)
	a.recordHistory()

	for range ticker.C {
		a.recordHistory()
	}
}

func (a *AdminAPI) recordHistory() {
	now := time.Now().In(a.location)
	count := len(a.userService.GetLoggedConnections())
	
	a.mu.Lock()
	defer a.mu.Unlock()

	// Hourly slot (0-23)
	hour := now.Hour()
	a.userHistoryHourly[hour] = count

	// Daily history
	if now.Sub(a.lastHistoryUpdate) >= 24*time.Hour {
		a.userHistoryDaily = append(a.userHistoryDaily, count)
		if len(a.userHistoryDaily) > 30 {
			a.userHistoryDaily = a.userHistoryDaily[1:]
		}
		a.lastHistoryUpdate = now
		// Periodically save (offload to goroutine to avoid holding lock during IO)
		go a.SaveHistory()
	}
}

func (a *AdminAPI) handleMonitorCharts(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	hH := make([]int, len(a.userHistoryHourly))
	copy(hH, a.userHistoryHourly)
	hD := make([]int, len(a.userHistoryDaily))
	copy(hD, a.userHistoryDaily)
	a.mu.RUnlock()

	res := map[string]interface{}{
		"class_distribution": a.classDistribution,
		"history_hourly":     hH,
		"history_daily":      hD,
	}
	json.NewEncoder(w).Encode(res)
}

func archetypeToString(a model.UserArchetype) string {
	switch a {
	case model.Mage: return "Mage"
	case model.Cleric: return "Cleric"
	case model.Warrior: return "Warrior"
	case model.Assasin: return "Assasin"
	case model.Thief: return "Thief"
	case model.Bard: return "Bard"
	case model.Druid: return "Druid"
	case model.Bandit: return "Bandit"
	case model.Paladin: return "Paladin"
	case model.Hunter: return "Hunter"
	case model.Worker: return "Worker"
	case model.Pirate: return "Pirate"
	default: return "Unknown"
	}
}

func (a *AdminAPI) handleConfigReload(w http.ResponseWriter, r *http.Request) {	newCfg, err := config.Load(a.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reloading config: %v", err), http.StatusInternalServerError)
		return
	}

	// Update the existing config struct in place so all references see the new values
	*a.config = *newCfg
	
	fmt.Fprintf(w, "Configuration reloaded from %s", a.configPath)
}

func (a *AdminAPI) handleEventList(w http.ResponseWriter, r *http.Request) {
	events := []map[string]string{
		{"name": "weekend_xp", "description": "Double XP for all players"},
		{"name": "gold_rush", "description": "Double gold drops from NPCs"},
	}
	json.NewEncoder(w).Encode(events)
}

func (a *AdminAPI) handleEventStart(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	switch name {
	case "weekend_xp":
		a.config.XpMultiplier = 2.0
		fmt.Fprintf(w, "Event %s started: XP Multiplier set to 2.0", name)
	case "gold_rush":
		a.config.GoldMultiplier = 2.0
		fmt.Fprintf(w, "Event %s started: Gold Multiplier set to 2.0", name)
	default:
		fmt.Fprintf(w, "Event %s started (no specific logic defined)", name)
	}
}

func (a *AdminAPI) handleEventStop(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	switch name {
	case "weekend_xp":
		a.config.XpMultiplier = 1.0
		fmt.Fprintf(w, "Event %s stopped: XP Multiplier restored to 1.0", name)
	case "gold_rush":
		a.config.GoldMultiplier = 1.0
		fmt.Fprintf(w, "Event %s stopped: Gold Multiplier restored to 1.0", name)
	default:
		fmt.Fprintf(w, "Event %s stopped", name)
	}
}

func (a *AdminAPI) handleConfigList(w http.ResponseWriter, r *http.Request) {
	keys := []map[string]interface{}{
		{"key": "version", "description": "Server version (read-only)", "type": "string", "value": a.config.Version},
		{"key": "max_users", "description": "Maximum concurrent users", "type": "int", "value": a.config.MaxConcurrentUsers},
		{"key": "creation_enabled", "description": "Allow new character creation", "type": "bool", "value": a.config.CharacterCreationEnabled},
		{"key": "admins_only", "description": "Restrict access to admins only", "type": "bool", "value": a.config.RestrictedToAdmins},
		{"key": "xp_multiplier", "description": "Global XP Multiplier", "type": "float", "value": a.config.XpMultiplier},
		{"key": "gold_multiplier", "description": "Global Gold Multiplier", "type": "float", "value": a.config.GoldMultiplier},
		{"key": "interval_attack", "description": "Interval between attacks (ms)", "type": "int64", "value": a.globalBalance.IntervalAttack},
		{"key": "interval_spell", "description": "Interval between spells (ms)", "type": "int64", "value": a.globalBalance.IntervalSpell},
		{"key": "interval_item", "description": "Interval to use items (ms)", "type": "int64", "value": a.globalBalance.IntervalItem},
		{"key": "interval_work", "description": "Interval to work (ms)", "type": "int64", "value": a.globalBalance.IntervalWork},
		{"key": "interval_magic_hit", "description": "Interval magic-hit (ms)", "type": "int64", "value": a.globalBalance.IntervalMagicHit},
		{"key": "interval_start_meditating", "description": "Delay to start meditating (ms)", "type": "int64", "value": a.globalBalance.IntervalStartMeditating},
		{"key": "interval_meditation", "description": "Interval between meditation regens (ms)", "type": "int64", "value": a.globalBalance.IntervalMeditation},
		{"key": "md5_enabled", "description": "Enable MD5 client validation", "type": "bool", "value": a.config.MD5Enabled},
		{"key": "check_critical_files", "description": "Check critical files integrity", "type": "bool", "value": a.config.CheckCriticalFiles},
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
	case "xp_multiplier":
		val = a.config.XpMultiplier
	case "gold_multiplier":
		val = a.config.GoldMultiplier
	case "interval_attack":
		val = a.globalBalance.IntervalAttack
	case "interval_spell":
		val = a.globalBalance.IntervalSpell
	case "interval_item":
		val = a.globalBalance.IntervalItem
	case "interval_work":
		val = a.globalBalance.IntervalWork
	case "interval_magic_hit":
		val = a.globalBalance.IntervalMagicHit
	case "interval_start_meditating":
		val = a.globalBalance.IntervalStartMeditating
	case "interval_meditation":
		val = a.globalBalance.IntervalMeditation
	case "md5_enabled":
		val = a.config.MD5Enabled
	case "check_critical_files":
		val = a.config.CheckCriticalFiles
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
	char := a.userService.GetCharacterByName(nick)
	if char != nil {
		conn := a.userService.GetConnection(char)
		if conn != nil {
			conn.Send(&outgoing.ErrorMessagePacket{Message: "Has sido baneado del servidor."})
			time.Sleep(100 * time.Millisecond)
		}
	}
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
	var npcsToRemove []*model.WorldNPC
	m.Modify(func(m *model.Map) {
		npcsToRemove = make([]*model.WorldNPC, 0, len(m.GetNpcs()))
		for _, npc := range m.GetNpcs() {
			npcsToRemove = append(npcsToRemove, npc)
		}
	})

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
		char := a.userService.GetCharacterByName(name)
		if char != nil {
			conn := a.userService.GetConnection(char)
			if conn != nil {
				conn.Send(&outgoing.ErrorMessagePacket{Message: "Has sido expulsado del servidor."})
				time.Sleep(100 * time.Millisecond)
			}
		}
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
	case "xp_multiplier":
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			a.config.XpMultiplier = f
		}
	case "gold_multiplier":
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			a.config.GoldMultiplier = f
		}
	case "interval_attack":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.globalBalance.IntervalAttack = i
		}
	case "interval_spell":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.globalBalance.IntervalSpell = i
		}
	case "interval_item":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.globalBalance.IntervalItem = i
		}
	case "interval_work":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.globalBalance.IntervalWork = i
		}
	case "interval_magic_hit":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.globalBalance.IntervalMagicHit = i
		}
	case "interval_start_meditating":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.globalBalance.IntervalStartMeditating = i
		}
	case "interval_meditation":
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			a.globalBalance.IntervalMeditation = i
		}
	case "md5_enabled":
		a.config.MD5Enabled = val == "true"
	case "check_critical_files":
		a.config.CheckCriticalFiles = val == "true"
	default:
		http.Error(w, "Unknown or read-only config key", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Config %s set to %s", key, val)
}

func (a *AdminAPI) handleWorldList(w http.ResponseWriter, r *http.Request) {
	loadedIDs := a.mapService.GetLoadedMaps()
	type MapInfo struct {
		ID    int `json:"id"`
		Users int `json:"users"`
		NPCs  int `json:"npcs"`
	}
	var list []MapInfo
	
	for _, id := range loadedIDs {
		users := 0
		a.mapService.ForEachCharacter(id, func(c *model.Character) { users++ })
		npcs := 0
		a.mapService.ForEachNpc(id, func(n *model.WorldNPC) { npcs++ })
		
		list = append(list, MapInfo{ID: id, Users: users, NPCs: npcs})
	}
	
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	
	json.NewEncoder(w).Encode(list)
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

func (a *AdminAPI) handleMonitorStats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := map[string]interface{}{
		"system": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"heap_alloc": m.HeapAlloc,
			"heap_sys":   m.HeapSys,
		},
		"connections": len(a.userService.GetLoggedConnections()),
	}

	// Network Stats
	conns := a.userService.GetLoggedConnections()
	type UserNetStat struct {
		User     string `json:"user"`
		BytesIn  uint64 `json:"bytes_in"`
		BytesOut uint64 `json:"bytes_out"`
		PktsIn   uint64 `json:"pkts_in"`
		PktsOut  uint64 `json:"pkts_out"`
		Age      int64  `json:"age_seconds"`
	}

	var netUsers []UserNetStat
	var totalIn, totalOut uint64

	for _, c := range conns {
		in, out, pIn, pOut, start := c.GetStats()
		totalIn += in
		totalOut += out

		name := "Unknown"
		if u := c.GetUser(); u != nil {
			name = u.Name
		}

		netUsers = append(netUsers, UserNetStat{
			User:     name,
			BytesIn:  in,
			BytesOut: out,
			PktsIn:   pIn,
			PktsOut:  pOut,
			Age:      int64(time.Since(start).Seconds()),
		})
	}

	stats["network"] = map[string]interface{}{
		"connections": netUsers,
		"total_in":    totalIn,
		"total_out":   totalOut,
	}

	// Map stats
	loadedMaps := a.mapService.GetLoadedMaps()
	mapStats := make([]map[string]interface{}, 0)
	for _, mapID := range loadedMaps {
		count := 0
		a.mapService.ForEachCharacter(mapID, func(c *model.Character) {
			count++
		})
		if count > 0 {
			mapStats = append(mapStats, map[string]interface{}{
				"id":    mapID,
				"users": count,
			})
		}
	}
	
	// Sort maps by users descending
	sort.Slice(mapStats, func(i, j int) bool {
		return mapStats[i]["users"].(int) > mapStats[j]["users"].(int)
	})

	// Top 10
	if len(mapStats) > 10 {
		mapStats = mapStats[:10]
	}
	stats["maps"] = mapStats

	json.NewEncoder(w).Encode(stats)
}

