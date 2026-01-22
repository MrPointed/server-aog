package incoming

import (
	"fmt"
	"math"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/service"
)

type LeftClickPacket struct {
	MapService    *service.MapService
	NpcService    *service.NpcService
	UserService   *service.UserService
	ObjectService *service.ObjectService
	AreaService   *service.AreaService
}

func (p *LeftClickPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	if buffer.ReadableBytes() < 2 {
		return false, nil
	}

	rawX, _ := buffer.Get()
	rawY, _ := buffer.Get()
	
	x := rawX - 1
	y := rawY - 1

	user := connection.GetUser()
	if user == nil {
		return true, nil
	}

	// Vision Range Check (approximate standard AO values)
	const RANGO_VISION_X = 8
	const RANGO_VISION_Y = 6

	if math.Abs(float64(user.Position.X)-float64(x)) > RANGO_VISION_X ||
		math.Abs(float64(user.Position.Y)-float64(y)) > RANGO_VISION_Y {
		return true, nil
	}

	mapID := user.Position.Map
	gameMap := p.MapService.GetMap(mapID)
	if gameMap == nil {
		return true, nil
	}

	// Helper function to get object at pos
	getObject := func(tx, ty int) *model.WorldObject {
		if tx < 0 || tx >= model.MapWidth || ty < 0 || ty >= model.MapHeight {
			return nil
		}
		return gameMap.GetTile(tx, ty).Object
	}

	foundSomething := false
	
	// Reset targets
	user.TargetMap = mapID
	user.TargetX = int(x)
	user.TargetY = int(y)
	user.TargetObj = 0
	user.TargetNPC = 0
	user.TargetUser = 0
	user.TargetNpcType = model.NTNone

	// Check Objects
	var targetObj *model.WorldObject
	var objX, objY int

	if obj := getObject(int(x), int(y)); obj != nil {
		targetObj = obj
		objX, objY = int(x), int(y)
	} else if obj := getObject(int(x)+1, int(y)); obj != nil && obj.Object.Type == model.OTDoor {
		targetObj = obj
		objX, objY = int(x)+1, int(y)
	} else if obj := getObject(int(x)+1, int(y)+1); obj != nil && obj.Object.Type == model.OTDoor {
		targetObj = obj
		objX, objY = int(x)+1, int(y)+1
	} else if obj := getObject(int(x), int(y)+1); obj != nil && obj.Object.Type == model.OTDoor {
		targetObj = obj
		objX, objY = int(x), int(y)+1
	}

	if targetObj != nil {
		user.TargetObj = targetObj.Object.ID
		user.TargetObjMap = mapID
		user.TargetObjX = objX
		user.TargetObjY = objY
		foundSomething = true
		
		msg := targetObj.Object.Name
		if targetObj.Amount > 1 {
			msg = fmt.Sprintf("%s - %d", msg, targetObj.Amount)
		}
		connection.Send(&outgoing.ConsoleMessagePacket{
			Message: msg,
			Font:    outgoing.INFO,
		})
	}

	// Check Characters
	// Check Y+1 first (user overlap?) - VB6 does this
	var targetCharIndex int16 = 0
	var isNpc bool = false
	
	checkTileForChar := func(tx, ty int) bool {
		if tx < 0 || tx >= model.MapWidth || ty < 0 || ty >= model.MapHeight {
			return false
		}
		tile := gameMap.GetTile(tx, ty)
		if tile.Character != nil {
			targetCharIndex = tile.Character.CharIndex
			isNpc = false
			return true
		}
		if tile.NPC != nil {
			targetCharIndex = tile.NPC.Index
			isNpc = true
			return true
		}
		return false
	}

	if !checkTileForChar(int(x), int(y)+1) {
		checkTileForChar(int(x), int(y))
	}

	if targetCharIndex > 0 {
		foundSomething = true
		if isNpc {
			// Handle NPC
			worldNpc := p.NpcService.GetWorldNpcs()[targetCharIndex]
			if worldNpc != nil {
				fmt.Printf("NPC LeftClick: User %s clicked NPC %d (%s)\n", user.Name, worldNpc.NPC.ID, worldNpc.NPC.Name)
				user.TargetNPC = targetCharIndex
				user.TargetNpcType = worldNpc.NPC.Type
				user.TargetUser = 0
				
				// Calculate status based on survival skill
				// For now simplified
				status := "(Sano) " 
				ratio := float64(worldNpc.HP) / float64(worldNpc.NPC.MaxHp)
				if ratio < 0.5 {
					status = "(Herido) "
				}
				// TODO: Full survival skill logic
				
				msg := fmt.Sprintf("%s%s", status, worldNpc.NPC.Name)
				connection.Send(&outgoing.ConsoleMessagePacket{
					Message: msg,
					Font:    outgoing.INFO,
				})

				// Show NPC description if available
				if worldNpc.NPC.Description != "" {
					fmt.Printf("NPC LeftClick: Sending description: %s\n", worldNpc.NPC.Description)
					connection.Send(&outgoing.ConsoleMessagePacket{
						Message: fmt.Sprintf("%s: %s", worldNpc.NPC.Name, worldNpc.NPC.Description),
						Font:    outgoing.INFO,
					})

					// Show as overhead text
					p.AreaService.BroadcastToArea(worldNpc.Position, &outgoing.ChatOverHeadPacket{
						Message:   worldNpc.NPC.Description,
						CharIndex: worldNpc.Index,
						R:         255,
						G:         255,
						B:         255,
					})
				}
			}
		} else {
			// Handle User
			targetUser := p.UserService.GetCharacterByIndex(targetCharIndex)
			if targetUser != nil {
				// Visibility check
				if targetUser.Invisible && !user.Invisible { // simplified admin check
					// Skip if invisible and viewer is not admin (assume user is not admin for now)
				} else {
					user.TargetUser = targetCharIndex
					user.TargetNPC = 0
					
					// Build status string
					desc := targetUser.Description
					if desc == "" {
						desc = "No Description"
					}
					
					// Simplified stats
					status := "Intacto)"
					ratio := float64(targetUser.Hp) / float64(targetUser.MaxHp)
					if ratio < 1.0 { status = "Herido)" }
					if ratio < 0.5 { status = "Malherido)" }
					if ratio == 0 { status = "Muerto)" }

					// TODO: Add Class and Race names lookup
					msg := fmt.Sprintf("%s - %s (%s", targetUser.Name, desc, status)
					
					font := outgoing.CITIZEN
					// TODO: Check criminal status, etc.
					
					connection.Send(&outgoing.ConsoleMessagePacket{
						Message: msg,
						Font:    font,
					})
				}
			}
		}
	} else {
		// Found nothing (except maybe object handled above)
		if !foundSomething {
			user.TargetNPC = 0
			user.TargetUser = 0
			user.TargetObj = 0
		}
	}

	return true, nil
}