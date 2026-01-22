package service

import (
	"fmt"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type SpellService struct {
	dao            *persistence.SpellDAO
	userService    *UserService
	messageService *MessageService
	objectService  *ObjectService
	spells         map[int]*model.Spell
}

func NewSpellService(dao *persistence.SpellDAO, userService *UserService, messageService *MessageService, objectService *ObjectService) *SpellService {
	return &SpellService{
		dao:            dao,
		userService:    userService,
		messageService: messageService,
		objectService:  objectService,
		spells:         make(map[int]*model.Spell),
	}
}

func (s *SpellService) LoadSpells() error {
	fmt.Println("Loading spells from data file...")
	defs, err := s.dao.Load()
	if err != nil {
		return err
	}
	s.spells = defs
	fmt.Printf("Successfully loaded %d spell definitions.\n", len(s.spells))
	return nil
}

func (s *SpellService) GetSpell(id int) *model.Spell {
	return s.spells[id]
}

func (s *SpellService) CastSpell(caster *model.Character, spellID int, target any) {
	spell := s.GetSpell(spellID)
	if spell == nil {
		return
	}

	conn := s.userService.GetConnection(caster)

	// Validations
	if caster.Mana < spell.ManaRequired {
		if conn != nil {
			conn.Send(&outgoing.ConsoleMessagePacket{
				Message: "No tienes suficiente maná.",
				Font:    outgoing.INFO,
			})
		}
		return
	}

	if caster.Stamina < spell.StaminaRequired {
		if conn != nil {
			conn.Send(&outgoing.ConsoleMessagePacket{
				Message: "Estás demasiado cansado.",
				Font:    outgoing.INFO,
			})
		}
		return
	}

	// Consume resources
	caster.Mana -= spell.ManaRequired
	caster.Stamina -= spell.StaminaRequired
	if conn != nil {
		conn.Send(outgoing.NewUpdateUserStatsPacket(caster))
	}

	// Broadcast magic words
	s.messageService.SendToArea(&outgoing.ConsoleMessagePacket{
		Message: fmt.Sprintf("%s: %s", caster.Name, spell.MagicWords),
		Font:    outgoing.TALK,
	}, caster.Position)

	// Resolve effect on target
	switch t := target.(type) {
	case *model.Character:
		s.applySpellToCharacter(caster, t, spell)
	case *model.WorldNPC:
		s.applySpellToNPC(caster, t, spell)
	default:
		// Target self if no target specified or invalid
		s.applySpellToCharacter(caster, caster, spell)
	}
}

func (s *SpellService) applySpellToCharacter(caster *model.Character, target *model.Character, spell *model.Spell) {
	// Send FX
	s.messageService.SendToArea(&outgoing.CreateFxPacket{
		CharIndex: target.CharIndex,
		FxID:      int16(spell.FX),
		Loops:     int16(spell.Loops),
	}, target.Position)

	// Damage logic
	if spell.MinHP < 0 || spell.MaxHP < 0 {
		damage := -spell.MinHP // Simplified
		target.Hp -= damage
		if target.Hp < 0 { target.Hp = 0 }
		
		connTarget := s.userService.GetConnection(target)
		if connTarget != nil {
			connTarget.Send(outgoing.NewUpdateUserStatsPacket(target))
			connTarget.Send(&outgoing.ConsoleMessagePacket{
				Message: fmt.Sprintf("¡%s te ha lanzado un hechizo y te quitó %d puntos de vida!", caster.Name, damage),
				Font:    outgoing.INFO,
			})
		}

		if target.Hp <= 0 {
			s.handleCharacterDeath(target)
		}
	}

	// Basic healing logic
	if spell.MinHP > 0 {
		heal := spell.MinHP
		target.Hp += heal
		if target.Hp > target.MaxHp {
			target.Hp = target.MaxHp
		}
		
		connTarget := s.userService.GetConnection(target)
		if connTarget != nil {
			connTarget.Send(outgoing.NewUpdateUserStatsPacket(target))
			connTarget.Send(&outgoing.ConsoleMessagePacket{
				Message: fmt.Sprintf("Te han sanado %d puntos de vida.", heal),
				Font:    outgoing.INFO,
			})
		}
	}
}

func (s *SpellService) handleCharacterDeath(char *model.Character) {
	char.Dead = true
	char.Hp = 0
	char.Body = 8   // Casper
	char.Head = 500 // Casper head

	conn := s.userService.GetConnection(char)
	if conn != nil {
		conn.Send(outgoing.NewUpdateUserStatsPacket(char))
		conn.Send(&outgoing.ConsoleMessagePacket{
			Message: "¡Has muerto!",
			Font:    outgoing.INFO,
		})
	}

	// Broadcast change
	s.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: char}, char.Position)
}

func (s *SpellService) applySpellToNPC(caster *model.Character, target *model.WorldNPC, spell *model.Spell) {
	// Send FX
	s.messageService.SendToArea(&outgoing.CreateFxPacket{
		CharIndex: target.Index,
		FxID:      int16(spell.FX),
		Loops:     int16(spell.Loops),
	}, target.Position)

	// Basic damage
	if spell.MaxHP < 0 {
		damage := -spell.MaxHP
		target.HP -= damage
		
		if target.HP <= 0 {
			s.messageService.SendToArea(&outgoing.CharacterRemovePacket{CharIndex: target.Index}, target.Position)
			// Exp logic...
			caster.Exp += target.NPC.Exp
			conn := s.userService.GetConnection(caster)
			if conn != nil {
				conn.Send(outgoing.NewUpdateUserStatsPacket(caster))
				conn.Send(&outgoing.ConsoleMessagePacket{
					Message: fmt.Sprintf("¡Has matado a la criatura! Ganaste %d puntos de experiencia.", target.NPC.Exp),
					Font:    outgoing.INFO,
				})
			}

			// Drop items
			for _, drop := range target.NPC.Drops {
				obj := s.objectService.GetObject(drop.ObjectID)
				if obj != nil {
					worldObj := &model.WorldObject{
						Object: obj,
						Amount: drop.Amount,
					}
					// Only drop if tile is empty
					if s.messageService.MapService.GetObjectAt(target.Position) == nil {
						s.messageService.MapService.PutObject(target.Position, worldObj)
						s.messageService.SendToArea(&outgoing.ObjectCreatePacket{
							X:            target.Position.X,
							Y:            target.Position.Y,
							GraphicIndex: int16(obj.GraphicIndex),
						}, target.Position)
					}
				}
			}

			s.messageService.MapService.RemoveNPC(target)
		}
	}
}