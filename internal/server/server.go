package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
	resourcesPath  string
}

func NewServer(addr string, resourcesPath string) *Server {
	res := resourcesPath
	cfg, err := config.Load(filepath.Join(res, "server.ini"))
	if err != nil {
		fmt.Printf("Warning: could not load server.ini: %v. Using defaults.\n", err)
		cfg = config.NewDefaultConfig()
	}

	objectDAO := persistence.NewObjectDAO(filepath.Join(res, "data/objects.dat"))
	objectService := service.NewObjectService(objectDAO)

	indexManager := service.NewCharacterIndexManager()
	npcDAO := persistence.NewNpcDAO(filepath.Join(res, "data/npcs.dat"))
	npcService := service.NewNpcService(npcDAO, indexManager)

	cityDAO := persistence.NewCityDAO(filepath.Join(res, "data/cities.dat"))
	cityService := service.NewCityService(cityDAO)

	balanceDAO := persistence.NewBalanceDAO(filepath.Join(res, "data/balances.dat"))
	archetypeMods, _, err := balanceDAO.Load()
	if err != nil {
		fmt.Printf("Critical error loading balances: %v\n", err)
	}
	combatFormulas := service.NewCombatFormulas(archetypeMods)
	intervalService := service.NewIntervalService(cfg)

	bodyService := service.NewCharacterBodyService()
	userService := service.NewUserService(bodyService)

	mapDAO := persistence.NewMapDAO(filepath.Join(res, "maps"), 150)
	if err := mapDAO.LoadProperties(filepath.Join(res, "maps.properties")); err != nil {
		fmt.Printf("Warning: could not load maps.properties: %v\n", err)
	}
	mapService := service.NewMapService(mapDAO, objectService, npcService)

	executor := actions.NewActionExecutor[*service.MapService](mapService)
	executor.Start()

	areaService := service.NewAreaService(mapService, userService)
	messageService := service.NewMessageService(userService, areaService, mapService, objectService)
	trainingService := service.NewTrainingService(messageService, userService, archetypeMods)

	spellDAO := persistence.NewSpellDAO(filepath.Join(res, "data/hechizos.dat"))
	spellService := service.NewSpellService(spellDAO, userService, npcService, messageService, objectService, intervalService, trainingService)

	resourceManager := service.NewResourceManager(objectService, npcService, mapService, spellService, cityService)
	resourceManager.LoadAll()

	combatService := service.NewCombatService(messageService, objectService, npcService, mapService, combatFormulas, intervalService, trainingService)
	timedEventsService := service.NewTimedEventsService(userService, messageService)
	timedEventsService.Start()

	aiService := service.NewAIService(npcService, mapService, areaService, userService, combatService, messageService, spellService)
	aiService.Start()

	skillService := service.NewSkillService(mapService, objectService, messageService, userService, npcService, spellService, intervalService)
	bankService := service.NewBankService(objectService, messageService, userService)

	fileDAO := persistence.NewFileDAO(filepath.Join(res, "charfiles"))
	loginService := service.NewLoginService(fileDAO, fileDAO, cfg, userService, mapService, bodyService, indexManager, messageService, objectService, cityService, spellService, executor)

	itemActionService := service.NewItemActionService(objectService, messageService, intervalService, bodyService)

	gmService := service.NewGMService(userService, mapService, messageService, executor)

	m := protocol.NewClientPacketsManager()
	// Register handlers
	m.RegisterHandler(protocol.CP_GMCommands, &incoming.GMCommandsPacket{GMService: gmService})
	m.RegisterHandler(protocol.CP_LoginExistingCharacter, &incoming.LoginExistingCharacterPacket{LoginService: loginService})
	m.RegisterHandler(protocol.CP_LoginNewCharacter, &incoming.LoginNewCharacterPacket{LoginService: loginService})
	m.RegisterHandler(protocol.CP_ThrowDice, &incoming.ThrowDicesPacket{})
	m.RegisterHandler(protocol.CP_Walk, &incoming.WalkPacket{MapService: mapService, AreaService: areaService, MessageService: messageService})
	m.RegisterHandler(protocol.CP_RequestPositionUpdate, &incoming.RequestPositionUpdatePacket{})
	m.RegisterHandler(protocol.CP_RequestAttributes, &incoming.RequestAttributesPacket{})
	m.RegisterHandler(protocol.CP_RequestSkills, &incoming.RequestSkillsPacket{})
	m.RegisterHandler(protocol.CP_Talk, &incoming.TalkPacket{MessageService: messageService})
	m.RegisterHandler(protocol.CP_Yell, &incoming.YellPacket{MessageService: messageService})
	m.RegisterHandler(protocol.CP_Whisper, &incoming.WhisperPacket{UserService: userService})
	m.RegisterHandler(protocol.CP_Attack, &incoming.AttackPacket{MapService: mapService, CombatService: combatService})
	m.RegisterHandler(protocol.CP_PickUp, &incoming.PickUpPacket{MapService: mapService, MessageService: messageService})
	m.RegisterHandler(protocol.CP_Online, &incoming.OnlinePacket{UserService: userService})
	m.RegisterHandler(protocol.CP_Meditate, &incoming.MeditatePacket{})
	m.RegisterHandler(protocol.CP_Quit, &incoming.QuitPacket{})
	m.RegisterHandler(protocol.CP_Drop, &incoming.DropPacket{MapService: mapService, MessageService: messageService, ObjectService: objectService})
	m.RegisterHandler(protocol.CP_CastSpell, &incoming.CastSpellPacket{MapService: mapService, SpellService: spellService})
	m.RegisterHandler(protocol.CP_LeftClick, &incoming.LeftClickPacket{MapService: mapService, NpcService: npcService, UserService: userService, ObjectService: objectService, AreaService: areaService})
	m.RegisterHandler(protocol.CP_UseItem, &incoming.UseItemPacket{ItemActionService: itemActionService})
	m.RegisterHandler(protocol.CP_EquipItem, &incoming.EquipItemPacket{ItemActionService: itemActionService})
	m.RegisterHandler(protocol.CP_ModifySkills, &incoming.ModifySkillsPacket{})
	m.RegisterHandler(protocol.CP_ChangeHeading, &incoming.ChangeHeadingPacket{AreaService: areaService})
	m.RegisterHandler(protocol.CP_Double_Click, &incoming.DoubleClickPacket{MapService: mapService, NpcService: npcService, UserService: userService, ObjectService: objectService, AreaService: areaService, BankService: bankService, SpellService: spellService})
	m.RegisterHandler(protocol.CP_Work, &incoming.UseSkillPacket{})
	m.RegisterHandler(protocol.CP_WorkLeftClick, &incoming.UseSkillClickPacket{SkillService: skillService})
	m.RegisterHandler(protocol.CP_Resurrect, &incoming.ResurrectPacket{MapService: mapService, AreaService: areaService, MessageService: messageService})

	m.RegisterHandler(protocol.CP_CommerceEnd, &incoming.CommerceEndPacket{})
	m.RegisterHandler(protocol.CP_CommerceBuy, &incoming.CommerceBuyPacket{NpcService: npcService, ObjectService: objectService, MessageService: messageService})
	m.RegisterHandler(protocol.CP_CommerceSell, &incoming.CommerceSellPacket{NpcService: npcService, ObjectService: objectService, MessageService: messageService})

	m.RegisterHandler(protocol.CP_BankEnd, &incoming.BankEndPacket{BankService: bankService})
	m.RegisterHandler(protocol.CP_BankExtractItem, &incoming.BankExtractItemPacket{BankService: bankService})
	m.RegisterHandler(protocol.CP_BankDeposit, &incoming.BankDepositPacket{BankService: bankService})
	m.RegisterHandler(protocol.CP_ExtractGold, &incoming.ExtractGoldPacket{BankService: bankService})
	m.RegisterHandler(protocol.CP_DepositGold, &incoming.DepositGoldPacket{BankService: bankService})

	return &Server{
		addr:           addr,
		packetsManager: m,
		mapService:     mapService,
		userService:    userService,
		loginService:   loginService,
		resourcesPath:  res,
	}
}

func (s *Server) Start() error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	if err := os.WriteFile("server.pid", []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		fmt.Printf("Warning: could not write server.pid: %v\n", err)
	}
	defer os.Remove("server.pid")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("AO Go Server listening on %s (Press Ctrl+C to stop)\n", s.addr)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go s.handleConnection(conn)
		}
	}()

	<-stop
	fmt.Println("\nShutting down server...")
	return nil
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

		for len(received) > 0 {
			db := network.NewDataBuffer(received)
			idByte := received[0]
			processed, err := s.packetsManager.Handle(db, c)

			if err != nil {
				fmt.Printf("Protocol error (Packet ID %d): %v\n", idByte, err)
				return
			}

			if processed {
				consumed := db.Pos()
				received = received[consumed:]
			} else {
				break
			}
		}
	}
}