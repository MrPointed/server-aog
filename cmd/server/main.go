package main

import (
	"fmt"
	"io"
	"net"

	"github.com/ao-go-server/internal/actions"
	"github.com/ao-go-server/internal/config"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/incoming"
	"github.com/ao-go-server/internal/service"
)

type Server struct {
	addr           string
	packetsManager *protocol.ClientPacketsManager
	mapService     *service.MapService
	userService    *service.UserService
	loginService   *service.LoginService
}

func NewServer(addr string) *Server {
	cfg := config.NewDefaultConfig()

	objectDAO := persistence.NewObjectDAO("../../resources/data/objects.dat")
	objectService := service.NewObjectService(objectDAO)
	if err := objectService.LoadObjects(); err != nil {
		fmt.Printf("Critical error loading objects: %v\n", err)
	}

	indexManager := service.NewCharacterIndexManager()
	npcDAO := persistence.NewNpcDAO("../../resources/data/npcs.dat")
	npcService := service.NewNpcService(npcDAO, indexManager)
	if err := npcService.LoadNpcs(); err != nil {
		fmt.Printf("Critical error loading NPCs: %v\n", err)
	}

	cityDAO := persistence.NewCityDAO("../../resources/data/cities.dat")
	cityService := service.NewCityService(cityDAO)
	if err := cityService.LoadCities(); err != nil {
		fmt.Printf("Critical error loading cities: %v\n", err)
	}

	balanceDAO := persistence.NewBalanceDAO("../../resources/data/balances.dat")
	archetypeMods, _, err := balanceDAO.Load()
	if err != nil {
		fmt.Printf("Critical error loading balances: %v\n", err)
	}
	combatFormulas := service.NewCombatFormulas(archetypeMods)
	intervalService := service.NewIntervalService(cfg)

	bodyService := service.NewCharacterBodyService()
	userService := service.NewUserService(bodyService)

	mapDAO := persistence.NewMapDAO("../../resources/maps", 150)
	mapService := service.NewMapService(mapDAO, objectService, npcService)
	mapService.LoadMaps()

	executor := actions.NewActionExecutor[*service.MapService](mapService)
	executor.Start()

	areaService := service.NewAreaService(mapService, userService)
	messageService := service.NewMessageService(userService, areaService, mapService)
	trainingService := service.NewTrainingService(messageService, userService, archetypeMods)
	combatService := service.NewCombatService(messageService, objectService, mapService, combatFormulas, intervalService, trainingService)
	timedEventsService := service.NewTimedEventsService(userService, messageService)
	timedEventsService.Start()

	aiService := service.NewAIService(npcService, mapService, areaService)
	aiService.Start()

	spellDAO := persistence.NewSpellDAO("../../resources/data/hechizos.dat")
	spellService := service.NewSpellService(spellDAO, userService, messageService, objectService, intervalService, trainingService)
	if err := spellService.LoadSpells(); err != nil {
		fmt.Printf("Critical error loading spells: %v\n", err)
	}

	skillService := service.NewSkillService(mapService, objectService, messageService, userService, npcService, spellService, intervalService)

	fileDAO := persistence.NewFileDAO("../../resources/charfiles")
	loginService := service.NewLoginService(fileDAO, fileDAO, cfg, userService, mapService, bodyService, indexManager, messageService, objectService, cityService, spellService, executor)

	itemActionService := service.NewItemActionService(objectService, messageService, intervalService, bodyService)

	m := protocol.NewClientPacketsManager()
	// Register handlers
	m.RegisterHandler(protocol.CP_LoginExistingCharacter, &incoming.LoginExistingCharacterPacket{LoginService: loginService})
	m.RegisterHandler(protocol.CP_LoginNewCharacter, &incoming.LoginNewCharacterPacket{LoginService: loginService})
	m.RegisterHandler(protocol.CP_ThrowDice, &incoming.ThrowDicesPacket{})
	m.RegisterHandler(protocol.CP_Walk, &incoming.WalkPacket{MapService: mapService, AreaService: areaService, MessageService: messageService})
	m.RegisterHandler(protocol.CP_RequestPositionUpdate, &incoming.RequestPositionUpdatePacket{})
	m.RegisterHandler(protocol.CP_Talk, &incoming.TalkPacket{MessageService: messageService})
	m.RegisterHandler(protocol.CP_Yell, &incoming.YellPacket{MessageService: messageService})
	m.RegisterHandler(protocol.CP_Whisper, &incoming.WhisperPacket{UserService: userService})
	m.RegisterHandler(protocol.CP_Attack, &incoming.AttackPacket{MapService: mapService, CombatService: combatService})
	m.RegisterHandler(protocol.CP_PickUp, &incoming.PickUpPacket{MapService: mapService, MessageService: messageService})
	m.RegisterHandler(protocol.CP_Drop, &incoming.DropPacket{MapService: mapService, MessageService: messageService, ObjectService: objectService})
	m.RegisterHandler(protocol.CP_CastSpell, &incoming.CastSpellPacket{MapService: mapService, SpellService: spellService})
	m.RegisterHandler(protocol.CP_LeftClick, &incoming.LeftClickPacket{MapService: mapService, NpcService: npcService, UserService: userService, ObjectService: objectService, AreaService: areaService})
	m.RegisterHandler(protocol.CP_UseItem, &incoming.UseItemPacket{ItemActionService: itemActionService})
	m.RegisterHandler(protocol.CP_EquipItem, &incoming.EquipItemPacket{ItemActionService: itemActionService})
	m.RegisterHandler(protocol.CP_ChangeHeading, &incoming.ChangeHeadingPacket{AreaService: areaService})
	m.RegisterHandler(protocol.CP_Double_Click, &incoming.DoubleClickPacket{MapService: mapService, NpcService: npcService, UserService: userService, ObjectService: objectService, AreaService: areaService})
	m.RegisterHandler(protocol.CP_UseSkill, &incoming.UseSkillPacket{})
	m.RegisterHandler(protocol.CP_UseSkillClick, &incoming.UseSkillClickPacket{SkillService: skillService})
	m.RegisterHandler(protocol.CP_Resurrect, &incoming.ResurrectPacket{MapService: mapService, AreaService: areaService, MessageService: messageService, BodyService: bodyService})

	m.RegisterHandler(protocol.CP_CommerceEnd, &incoming.CommerceEndPacket{})
	m.RegisterHandler(protocol.CP_CommerceBuy, &incoming.CommerceBuyPacket{NpcService: npcService, ObjectService: objectService, MessageService: messageService})
	m.RegisterHandler(protocol.CP_CommerceSell, &incoming.CommerceSellPacket{NpcService: npcService, ObjectService: objectService, MessageService: messageService})

	return &Server{
		addr:           addr,
		packetsManager: m,
		mapService:     mapService,
		userService:    userService,
		loginService:   loginService,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("AO Go Server listening on %s\n", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

type connection struct {
	conn       net.Conn
	attributes map[int]byte
	user       *model.Character
}

func (c *connection) Send(packet protocol.OutgoingPacket) error {
	id, err := protocol.GetOutgoingPacketID(packet)
	if err != nil {
		return err
	}

	buf := network.NewDataBuffer(nil)
	if err := protocol.WriteOutgoing(packet, id, buf); err != nil {
		return err
	}

	return c.SendBytes(buf.Bytes())
}

func (c *connection) SendBytes(data []byte) error {
	_, err := c.conn.Write(data)
	return err
}

func (c *connection) Disconnect() {
	c.conn.Close()
}

func (c *connection) SetAttribute(attr int, value byte) {
	if c.attributes == nil {
		c.attributes = make(map[int]byte)
	}
	c.attributes[attr] = value
}

func (c *connection) GetAttribute(attr int) byte {
	if c.attributes == nil {
		return 0
	}
	return c.attributes[attr]
}

func (c *connection) GetUser() *model.Character {
	return c.user
}

func (c *connection) SetUser(user *model.Character) {
	c.user = user
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	c := &connection{conn: conn}
	defer s.loginService.OnUserDisconnect(c)

	// Buffer to accumulate data
	received := make([]byte, 0)
	tmp := make([]byte, 1024)

	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			return
		}

		received = append(received, tmp[:n]...)

		// Try to process packets
		for len(received) > 0 {
			db := network.NewDataBuffer(received)
			idByte := received[0]
			processed, err := s.packetsManager.Handle(db, c)

			if err != nil {
				fmt.Printf("Protocol error (Packet ID %d): %v\n", idByte, err)
				return // Close connection on protocol error
			}

			if processed {
				fmt.Printf("Handled packet ID: %d\n", idByte)
				// Remove processed bytes from 'received'
				consumed := db.Pos()
				received = received[consumed:]
			} else {
				// Incomplete packet, wait for more data
				break
			}
		}
	}
}

func main() {
	server := NewServer(":7666")
	if err := server.Start(); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
