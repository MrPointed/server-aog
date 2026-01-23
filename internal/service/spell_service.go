package service

import (
	"fmt"
	"math/rand"

	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/persistence"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type SpellService struct {
	dao             *persistence.SpellDAO
	userService     *UserService
	messageService  *MessageService
	objectService   *ObjectService
	intervals       *IntervalService
	trainingService *TrainingService
	spells          map[int]*model.Spell
}

func NewSpellService(dao *persistence.SpellDAO, userService *UserService, messageService *MessageService, objectService *ObjectService, intervals *IntervalService, trainingService *TrainingService) *SpellService {
	return &SpellService{
		dao:             dao,
		userService:     userService,
		messageService:  messageService,
		objectService:   objectService,
		intervals:       intervals,
		trainingService: trainingService,
		spells:          make(map[int]*model.Spell),
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
	s.applySpellEffectToCharacter(target, spell, caster.Name)
}

func (s *SpellService) NpcLanzaSpellSobreUser(npc *model.WorldNPC, target *model.Character, spellID int) bool {
	// Check interval
	if !s.intervals.CanNPCCastSpell(npc) {
		return false
	}

	spell := s.GetSpell(spellID)
	if spell == nil {
		return false
	}

	// Broadcast magic words as overhead text
	s.messageService.SendToArea(&outgoing.ChatOverHeadPacket{
		Message:   spell.MagicWords,
		CharIndex: npc.Index,
		R:         65,
		G:         190,
		B:         156,
	}, npc.Position)

	// Resolve effect on target
	if spell.TargetType == model.TargetUser || spell.TargetType == model.TargetUserAndNpc {
		// We need a version of applySpellToCharacter that doesn't require a 'caster' *model.Character
		// or we use a dummy caster. Let's refactor slightly.
		s.applySpellEffectToCharacter(target, spell, npc.NPC.Name)
	}

	// Update interval
	s.intervals.UpdateNPCLastSpell(npc)

	return true
}

func (s *SpellService) applySpellEffectToCharacter(target *model.Character, spell *model.Spell, casterName string) {
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
			target.Hp = target.MaxHp
			target.Mana = 0
			target.Stamina = 0
			target.Body = 1
			target.Head = target.OriginalHead

			s.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: target}, target.Position)
			s.messageService.SendConsoleMessage(target, "¡Has sido revivido!", outgoing.INFO)
			return
		}
	}

	if target.Dead {
		return
	}

	// Heal
	if spell.SubeHP == 1 {
		amount := rand.Intn(spell.MaxHP-spell.MinHP+1) + spell.MinHP
		target.Hp += amount
		if target.Hp > target.MaxHp {
			target.Hp = target.MaxHp
		}
		s.messageService.SendConsoleMessage(target, fmt.Sprintf("Te han sanado %d puntos.", amount), outgoing.FIGHT)
	}

	// Damage
	if spell.SubeHP == 2 {
		amount := rand.Intn(spell.MaxHP-spell.MinHP+1) + spell.MinHP
		target.Hp -= amount
		if target.Hp < 0 {
			target.Hp = 0
		}

		s.messageService.SendConsoleMessage(target, fmt.Sprintf("¡%s te quitó %d puntos de vida!", casterName, amount), outgoing.FIGHT)

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
		target.Immobilized = true
		s.messageService.SendConsoleMessage(target, "¡Te han inmovilizado!", outgoing.INFO)
	}

	// Poison
	if spell.Poison {
		target.Poisoned = true
		s.messageService.SendConsoleMessage(target, "¡Te han envenenado!", outgoing.INFO)
	}

	conn := s.userService.GetConnection(target)
	if conn != nil {
		conn.Send(outgoing.NewUpdateUserStatsPacket(target))
	}
}

func (s *SpellService) SacerdoteHealUser(target *model.Character) {
	// SND_CURAR_SACERDOTE = 13 (example)
	s.messageService.SendToArea(&outgoing.PlayWavePacket{
		Wave: 13,
		X:    target.Position.X,
		Y:    target.Position.Y,
	}, target.Position)

	target.Hp = target.MaxHp
	s.messageService.SendConsoleMessage(target, "El sacerdote te ha curado!!", outgoing.INFO)

	// If newbie, restore everything
	if target.Level <= 13 { // Assuming newbie level <= 13
		target.Mana = target.MaxMana
		target.Poisoned = false
		s.messageService.SendConsoleMessage(target, "El sacerdote te ha restaurado el mana completamente.", outgoing.INFO)
	}

	conn := s.userService.GetConnection(target)
	if conn != nil {
		conn.Send(outgoing.NewUpdateUserStatsPacket(target))
	}
}

func (s *SpellService) SacerdoteResucitateUser(target *model.Character) {
	// SND_RESUCITAR_SACERDOTE = 16 (example)
	s.messageService.SendToArea(&outgoing.PlayWavePacket{
		Wave: 16,
		X:    target.Position.X,
		Y:    target.Position.Y,
	}, target.Position)

	if target.Dead {
		target.Dead = false
		target.Hp = target.MaxHp / 10 // Standard AO revive gives 1/10 HP
		if target.Level <= 13 {
			target.Hp = target.MaxHp
			target.Mana = target.MaxMana
		}
		target.Body = 1
		target.Head = target.OriginalHead

		s.messageService.SendToArea(&outgoing.CharacterChangePacket{Character: target}, target.Position)
		s.messageService.SendConsoleMessage(target, "¡Has sido resucitado!", outgoing.INFO)
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
		s.messageService.SendConsoleMessage(caster, fmt.Sprintf("Has curado %d puntos a la criatura.", amount), outgoing.FIGHT)
	}

	// Damage
	if spell.SubeHP == 2 {
		amount := rand.Intn(spell.MaxHP-spell.MinHP+1) + spell.MinHP
		
		// Grant experience proportional to damage
		s.grantExperience(caster, target, amount)

		target.HP -= amount
		s.messageService.SendConsoleMessage(caster, fmt.Sprintf("Has quitado %d puntos a la criatura.", amount), outgoing.FIGHT)
		
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
	
	if target.RemainingExp > 0 {
		caster.Exp += target.RemainingExp
		s.messageService.SendConsoleMessage(caster, fmt.Sprintf("¡Has matado a la criatura! Ganaste %d exp.", target.RemainingExp), outgoing.INFO)
		target.RemainingExp = 0
		s.trainingService.CheckLevel(caster)
	}

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

func (s *SpellService) grantExperience(attacker *model.Character, victim *model.WorldNPC, damage int) {
	if victim.NPC.MaxHp == 0 || victim.NPC.Exp == 0 || victim.RemainingExp <= 0 {
		return
	}

	expToGive := int(float32(damage) * (float32(victim.NPC.Exp) / float32(victim.NPC.MaxHp)))
	
	// Ensure at least 1 exp if damage was dealt and there's exp left
	if expToGive == 0 && damage > 0 && victim.RemainingExp > 0 {
		expToGive = 1
	}

	if expToGive > victim.RemainingExp {
		expToGive = victim.RemainingExp
	}

	if expToGive > 0 {
		attacker.Exp += expToGive
		victim.RemainingExp -= expToGive
		s.messageService.SendConsoleMessage(attacker, fmt.Sprintf("Has ganado %d puntos de experiencia.", expToGive), outgoing.FIGHT)
		s.trainingService.CheckLevel(attacker)
	}
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