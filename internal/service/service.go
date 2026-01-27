package service

import (
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type MapService interface {
	LoadMaps()
	SaveCache()
	GetLoadedMaps() []int
	LoadMap(id int) error
	UnloadMap(id int)
	ReloadMap(id int) error
	GetMap(id int) *model.Map
	PutCharacterAtPos(char *model.Character, pos model.Position)
	RemoveCharacter(char *model.Character)
	ForEachCharacter(mapID int, f func(*model.Character))
	ForEachNpc(mapID int, f func(*model.WorldNPC))
	GetObjectAt(pos model.Position) *model.WorldObject
	PutObject(pos model.Position, obj *model.WorldObject)
	RemoveObject(pos model.Position)
	GetNPCAt(pos model.Position) *model.WorldNPC
	RemoveNPC(npc *model.WorldNPC)
	IsInPlayableArea(x, y int) bool
	MoveNpc(npc *model.WorldNPC, newPos model.Position) bool
	MoveCharacterTo(char *model.Character, heading model.Heading) (model.Position, bool)
	IsSafeZone(pos model.Position) bool
	IsPkMap(mapID int) bool
	IsInvalidPosition(pos model.Position) bool
	IsTileEmpty(mapID int, x, y int) bool
	IsBlocked(mapID, x, y int) bool
	SpawnNpcInMap(npcID int, mapID int) *model.WorldNPC
}

type NpcService interface {
	LoadNpcs() error
	GetNpcDef(id int) *model.NPC
	SpawnNpc(id int, pos model.Position) *model.WorldNPC
	RemoveNPC(npc *model.WorldNPC, mapService MapService)
	GetWorldNpcs() []*model.WorldNPC
	GetWorldNpcByIndex(index int16) *model.WorldNPC
	ChangeNpcHeading(npc *model.WorldNPC, heading model.Heading, areaService AreaService)
	MoveNpc(npc *model.WorldNPC, newPos model.Position, heading model.Heading, mapService MapService, areaService AreaService) bool
}

type ObjectService interface {
	LoadObjects() error
	GetObject(id int) *model.Object
}

type UserService interface {
	IsLoggedIn(conn protocol.Connection) bool
	IsUserLoggedIn(name string) bool
	GetCharacterByName(name string) *model.Character
	LogIn(conn protocol.Connection)
	LogOut(conn protocol.Connection)
	GetCharacterByIndex(index int16) *model.Character
	GetConnection(char *model.Character) protocol.Connection
	GetLoggedCharacters() []*model.Character
	GetLoggedConnections() []protocol.Connection
	KickByName(name string) bool
	KickByIP(ip string) int
	BodyService() BodyService
}

type AreaService interface {
	BroadcastToArea(pos model.Position, packet protocol.OutgoingPacket)
	BroadcastNearby(char *model.Character, packet protocol.OutgoingPacket)
	NotifyNpcMovement(npc *model.WorldNPC, oldPos model.Position)
	NotifyCharacterMovement(char *model.Character, oldPos model.Position)
	NotifyMovement(char *model.Character, oldPos model.Position)
	SendAreaInit(conn protocol.Connection)
	SendAreaState(char *model.Character)
	SendAreaObjectsOnly(char *model.Character)
	NotifyCharacterHeadingChange(char *model.Character)
	InRange(p1, p2 model.Position) bool
	GetArea(x, y byte) (int, int)
}

type MessageService interface {
	SendMessage(char *model.Character, msg string, msgType outgoing.Font)
	SendConsoleMessage(char *model.Character, msg string, font outgoing.Font)
	BroadcastMessage(msg string, msgType outgoing.Font)
	BroadcastNearby(char *model.Character, packet protocol.OutgoingPacket)
	SendToAll(packet protocol.OutgoingPacket)
	SendToArea(packet protocol.OutgoingPacket, pos model.Position)
	SendToAreaButUser(packet protocol.OutgoingPacket, pos model.Position, exclude *model.Character)
	SendToMap(packet protocol.OutgoingPacket, mapId int)
	HandleDeath(char *model.Character, msg string)
	HandleResurrection(char *model.Character)
	MapService() MapService
	UserService() UserService
	AreaService() AreaService
}

type CombatService interface {
	ResolveAttack(attacker *model.Character, target any)
	NpcAtacaUser(npc *model.WorldNPC, victim *model.Character) bool
}

type SpellService interface {
	LoadSpells() error
	GetSpell(id int) *model.Spell
	CastSpell(caster *model.Character, spellID int, target any)
	NpcLanzaSpellSobreUser(npc *model.WorldNPC, target *model.Character, spellID int) bool
	PriestHealUser(target *model.Character)
	PriestResucitateUser(target *model.Character)
}

type SkillService interface {
	HandleUseSkillClick(user *model.Character, skill model.Skill, x, y byte)
}

type LoginService interface {
	ConnectNewCharacter(conn protocol.Connection, nick, password, mail string, raceId, genderId, archetypeId, headId, cityId byte, clientHash, version string) error
	ConnectExistingCharacter(conn protocol.Connection, nick, password, version, clientHash string) error
	OnUserDisconnect(conn protocol.Connection)
	LockAccount(nick string) error
	UnlockAccount(nick string) error
	ResetPassword(nick, newPassword string) error
	TeleportPlayer(nick string, mapID, x, y int) error
	SavePlayer(nick string) error
	SaveAllPlayers()
}

type BodyService interface {
	IsValidHead(head int, race model.Race, gender model.Gender) bool
	GetBody(race model.Race, gender model.Gender) int
}

type IntervalService interface {
	CanAttack(char *model.Character) bool
	CanCastSpell(char *model.Character) bool
	CanUseItem(char *model.Character) bool
	CanWork(char *model.Character) bool
	CanNPCAttack(npc *model.WorldNPC) bool
	CanNPCCastSpell(npc *model.WorldNPC) bool
	UpdateLastAttack(char *model.Character)
	UpdateLastSpell(char *model.Character)
	UpdateLastItem(char *model.Character)
	UpdateLastWork(char *model.Character)
	UpdateNPCLastAttack(npc *model.WorldNPC)
	UpdateNPCLastSpell(npc *model.WorldNPC)
}

type TrainingService interface {
	CheckLevel(char *model.Character)
}

type ItemActionService interface {
	UseItem(char *model.Character, slotIdx int, connection protocol.Connection)
	EquipItem(char *model.Character, slotIdx int, connection protocol.Connection)
}

type CityService interface {
	LoadCities() error
	GetCity(id int) (model.City, bool)
}

type BankService interface {
	OpenBank(char *model.Character)
	CloseBank(char *model.Character)
	DepositGold(char *model.Character, amount int)
	ExtractGold(char *model.Character, amount int)
	DepositItem(char *model.Character, slotIdx int, amount int)
	ExtractItem(char *model.Character, bankSlotIdx int, amount int)
}

type GmService interface {
	HandleCommand(conn protocol.Connection, cmdID byte, buffer *network.DataBuffer) (bool, error)
}

type TimedEventsService interface {
	Start()
	Stop()
}

type AiService interface {
	Start()
	Stop()
	SetEnabled(enabled bool)
}

type ResourceManager interface {
	LoadAll()
}
