package service

import (
	"fmt"
	"math/rand"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type SpellService struct {
	dao            *persistence.SpellDAO
	userService    *UserService
	messageService *MessageService
	objectService  *ObjectService
	intervals      *IntervalService
	spells         map[int]*model.Spell
}

func NewSpellService(dao *persistence.SpellDAO, userService *UserService, messageService *MessageService, objectService *ObjectService, intervals *IntervalService) *SpellService {
	return &SpellService{
		dao:            dao,
		userService:    userService,
		messageService: messageService,
		objectService:  objectService,
		intervals:      intervals,
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
	fmt.Printf("CastSpell: Caster %s, SpellID %d\n", caster.Name, spellID)
	spell := s.GetSpell(spellID)
	if spell == nil {
		fmt.Printf("CastSpell: Spell %d not found\n", spellID)
		return
	}

	conn := s.userService.GetConnection(caster)

	// Check intervals
	if !s.intervals.CanCastSpell(caster) {
		return
	}

	// Validations
	if caster.Mana < spell.ManaRequired {
		fmt.Println("CastSpell: Not enough mana")
		if conn != nil {
			conn.Send(&outgoing.ConsoleMessagePacket{
				Message: "No tienes suficiente maná.",
				Font:    outgoing.INFO,
			})
		}
		return
	}

	if caster.Stamina < spell.StaminaRequired {
		fmt.Println("CastSpell: Not enough stamina")
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
	
	// Update last spell time
	s.intervals.UpdateLastSpell(caster)

	if conn != nil {
		conn.Send(outgoing.NewUpdateUserStatsPacket(caster))
	}

	// Broadcast magic words as overhead text
	s.messageService.SendToArea(&outgoing.ChatOverHeadPacket{
		Message:   spell.MagicWords,
		CharIndex: caster.CharIndex,
		R:         65,
		G:         190,
		B:         156,
	}, caster.Position)

	// Keep a console message for the caster specifically? 
	// Standard AO usually just does overhead for everyone.
	// But let's add a console msg for the caster for better feedback as previously requested.
	if conn != nil {
		conn.Send(&outgoing.ConsoleMessagePacket{
			Message: fmt.Sprintf("Lanzas: %s", spell.MagicWords),
			Font:    outgoing.TALK,
		})
	}

	// Resolve effect on target
	switch t := target.(type) {
	case *model.Character:
		fmt.Printf("CastSpell: Target is Character %s. Spell Type: %d\n", t.Name, spell.TargetType)
		if spell.TargetType == model.TargetUser || spell.TargetType == model.TargetUserAndNpc {
			s.applySpellToCharacter(caster, t, spell)
		} else {
			fmt.Println("CastSpell: Invalid target type for Character")
			if conn != nil {
				conn.Send(&outgoing.ConsoleMessagePacket{Message: "Target inválido.", Font: outgoing.INFO})
			}
		}
	case *model.WorldNPC:
		fmt.Printf("CastSpell: Target is NPC. Spell Type: %d\n", spell.TargetType)
		if spell.TargetType == model.TargetNpc || spell.TargetType == model.TargetUserAndNpc {
			s.applySpellToNPC(caster, t, spell)
		} else {
			fmt.Println("CastSpell: Invalid target type for NPC")
			if conn != nil {
				conn.Send(&outgoing.ConsoleMessagePacket{Message: "Target inválido.", Font: outgoing.INFO})
			}
		}
	case model.Position:
		fmt.Printf("CastSpell: Target is Position. Spell Type: %d\n", spell.TargetType)
		if spell.TargetType == model.TargetTerrain {
			s.applySpellToPosition(caster, t, spell)
		} else {
			fmt.Println("CastSpell: Invalid target type for Position")
			if conn != nil {
				conn.Send(&outgoing.ConsoleMessagePacket{Message: "Debes seleccionar un objetivo.", Font: outgoing.INFO})
			}
		}
	default:
		fmt.Println("CastSpell: Unknown target type")
		// Target self if no target specified or invalid, usually handled by skill service logic before calling this.
	}
}

func (s *SpellService) applySpellToCharacter(caster *model.Character, target *model.Character, spell *model.Spell) {
	// FX
	s.messageService.SendToArea(&outgoing.CreateFxPacket{
		CharIndex: target.CharIndex,
		FxID:      int16(spell.FX),
		Loops:     int16(spell.Loops),
	}, target.Position)

	// Revive
	if spell.Revive {
		if target.Dead {
			target.Dead = false
			target.Hp = target.MaxHp // Full HP or partial? Mod says usually partial or penalizes. Let's do full for simplicity now.
			target.Mana = 0
			target.Stamina = 0
			target.Body = 1 // Default body, should restore original
			target.Head = target.OriginalHead
			
			s.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: target}, target.Position)
			s.messageService.SendConsoleMessage(target, "¡Has sido revivido!", outgoing.INFO)
			s.messageService.SendConsoleMessage(caster, "Has revivido a tu objetivo.", outgoing.INFO)
			return // Don't apply other effects if dead
		} else {
			s.messageService.SendConsoleMessage(caster, "El objetivo está vivo.", outgoing.INFO)
		}
	}

	if target.Dead {
		s.messageService.SendConsoleMessage(caster, "El objetivo está muerto.", outgoing.INFO)
		return
	}

	// Heal (SubeHP = 1)
	if spell.SubeHP == 1 {
		amount := rand.Intn(spell.MaxHP-spell.MinHP+1) + spell.MinHP
		target.Hp += amount
		if target.Hp > target.MaxHp {
			target.Hp = target.MaxHp
		}
		s.messageService.SendConsoleMessage(target, fmt.Sprintf("Te han sanado %d puntos.", amount), outgoing.INFO)
		s.messageService.SendConsoleMessage(caster, fmt.Sprintf("Has sanado %d puntos.", amount), outgoing.INFO)
	}

	// Damage (SubeHP = 2)
	if spell.SubeHP == 2 {
		amount := rand.Intn(spell.MaxHP-spell.MinHP+1) + spell.MinHP
		target.Hp -= amount
		if target.Hp < 0 { target.Hp = 0 }
		
		s.messageService.SendConsoleMessage(target, fmt.Sprintf("¡%s te quitó %d puntos de vida!", caster.Name, amount), outgoing.INFO)
		s.messageService.SendConsoleMessage(caster, fmt.Sprintf("Has quitado %d puntos de vida.", amount), outgoing.INFO)

		if target.Hp <= 0 {
			s.handleCharacterDeath(target)
		}
	}

	// Paralysis / Immobilize
	if spell.Paralyzes {
		target.Paralyzed = true
		s.messageService.SendConsoleMessage(target, "¡Te han paralizado!", outgoing.INFO)
	}
	if spell.Immobilizes {
		target.Immobilized = true // Need to ensure Character struct has this
		s.messageService.SendConsoleMessage(target, "¡Te han inmovilizado!", outgoing.INFO)
	}

	// Poison
	if spell.Poison {
		target.Poisoned = true
		s.messageService.SendConsoleMessage(target, "¡Te han envenenado!", outgoing.INFO)
	}

	// Cure Poison
	if spell.CurePoison {
		if target.Poisoned {
			target.Poisoned = false
			s.messageService.SendConsoleMessage(target, "Te has curado del envenenamiento.", outgoing.INFO)
			s.messageService.SendConsoleMessage(caster, "Has curado el envenenamiento.", outgoing.INFO)
		} else {
			s.messageService.SendConsoleMessage(caster, "El objetivo no está envenenado.", outgoing.INFO)
		}
	}

	s.messageService.userService.GetConnection(target).Send(outgoing.NewUpdateUserStatsPacket(target))
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
	// FX
	s.messageService.SendToArea(&outgoing.CreateFxPacket{
		CharIndex: target.Index,
		FxID:      int16(spell.FX),
		Loops:     int16(spell.Loops),
	}, target.Position)

	// Heal
	if spell.SubeHP == 1 {
		amount := rand.Intn(spell.MaxHP-spell.MinHP+1) + spell.MinHP
		target.HP += amount
		s.messageService.SendConsoleMessage(caster, fmt.Sprintf("Has curado %d puntos a la criatura.", amount), outgoing.INFO)
	}

	// Damage
	if spell.SubeHP == 2 {
		amount := rand.Intn(spell.MaxHP-spell.MinHP+1) + spell.MinHP
		target.HP -= amount
		s.messageService.SendConsoleMessage(caster, fmt.Sprintf("Has quitado %d puntos a la criatura.", amount), outgoing.INFO)
		
		if target.HP <= 0 {
			s.handleNpcDeath(caster, target)
		}
	}
	
	// Paralysis
	if spell.Paralyzes {
		// target.Paralyzed = true // NPC struct needs this field
	}
}

func (s *SpellService) handleNpcDeath(caster *model.Character, target *model.WorldNPC) {
	s.messageService.SendToArea(&outgoing.CharacterRemovePacket{CharIndex: target.Index}, target.Position)
	
	caster.Exp += target.NPC.Exp
	s.messageService.SendConsoleMessage(caster, fmt.Sprintf("¡Has matado a la criatura! Ganaste %d exp.", target.NPC.Exp), outgoing.INFO)
	s.messageService.userService.GetConnection(caster).Send(outgoing.NewUpdateUserStatsPacket(caster))

	// Drop items
	for _, drop := range target.NPC.Drops {
		obj := s.objectService.GetObject(drop.ObjectID)
		if obj != nil {
			worldObj := &model.WorldObject{
				Object: obj,
				Amount: drop.Amount,
			}
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

func (s *SpellService) applySpellToPosition(caster *model.Character, pos model.Position, spell *model.Spell) {
	// Summon
	if spell.SummonNPC > 0 {
		// Simple summon logic
		// Need NpcService to spawn? SpellService doesn't have NpcService.
		// I should inject it. Or expose SpawnNpc in MapService? 
		// NpcService is best.
		// For now, print message "Summon not implemented fully".
		s.messageService.SendConsoleMessage(caster, "Invocación no implementada por completo.", outgoing.INFO)
	}
}