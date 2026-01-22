package service

import (
	"math/rand"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/utils"
)

type CombatFormulas struct {
	archetypeModifiers map[model.UserArchetype]*model.ArchetypeModifiers
}

func NewCombatFormulas(archetypeModifiers map[model.UserArchetype]*model.ArchetypeModifiers) *CombatFormulas {
	return &CombatFormulas{
		archetypeModifiers: archetypeModifiers,
	}
}

func (f *CombatFormulas) GetShieldEvasionPower(char *model.Character) int {
	mod := f.archetypeModifiers[char.Archetype]
	if mod == nil {
		return 0
	}
	return int(float32(char.Skills[model.Defense]) * mod.Shield) / 2
}

func (f *CombatFormulas) GetEvasionPower(char *model.Character) int {
	mod := f.archetypeModifiers[char.Archetype]
	if mod == nil {
		return 0
	}
	
	skillTactics := char.Skills[model.CombatTactics]
	agility := int(char.Attributes[model.Dexterity])
	
	lTemp := float32(skillTactics + skillTactics / 33 * agility) * mod.Evasion
	return int(lTemp + (2.5 * float32(utils.Max(int(char.Level)-12, 0))))
}

func (f *CombatFormulas) GetAttackPower(char *model.Character, weapon *model.Object) int {
	mod := f.archetypeModifiers[char.Archetype]
	if mod == nil {
		return 0
	}

	if weapon == nil {
		return f.calculateAttackPower(char, char.Skills[model.Wrestling], mod.WrestlingAttack)
	}

	if weapon.Ranged {
		return f.calculateAttackPower(char, char.Skills[model.Projectiles], mod.ProjectileAttack)
	}

	return f.calculateAttackPower(char, char.Skills[model.MeleeCombat], mod.MeleeAttack)
}

func (f *CombatFormulas) calculateAttackPower(char *model.Character, skill int, mod float32) int {
	var powerTemp int
	agility := int(char.Attributes[model.Dexterity])

	if skill < 31 {
		powerTemp = int(float32(skill) * mod)
	} else if skill < 61 {
		powerTemp = int(float32(skill+agility) * mod)
	} else if skill < 91 {
		powerTemp = int(float32(skill+2*agility) * mod)
	} else {
		powerTemp = int(float32(skill+3*agility) * mod)
	}

	return powerTemp + int(2.5*float32(utils.Max(int(char.Level)-12, 0)))
}

func (f *CombatFormulas) CalculateHitChance(attackerPower, victimEvasion int) int {
	chance := 50 + int(float32(attackerPower-victimEvasion)*0.4)
	return utils.Max(10, utils.Min(90, chance))
}

func (f *CombatFormulas) CalculateDamage(attacker *model.Character, weapon *model.Object, isNpc bool) int {
	mod := f.archetypeModifiers[attacker.Archetype]
	if mod == nil {
		return 0
	}

	var weaponDmg int
	var modClase float32

	if weapon != nil {
		weaponDmg = utils.RandomNumber(weapon.MinHit, weapon.MaxHit)
		if weapon.Ranged {
			modClase = mod.ProjectileDamage
		} else {
			modClase = mod.MeleeDamage
		}
	} else {
		weaponDmg = utils.RandomNumber(4, 9)
		modClase = mod.WrestlingDamage
	}

	userDmg := utils.RandomNumber(int(attacker.Attributes[model.Strength])/3, int(attacker.Attributes[model.Strength])/2)
	
	// Simplified formula from CalcularDano
	// CalcularDano = (3 * DanoArma + ((DanoMaxArma / 5) * MaximoInt(0, .Stats.UserAtributos(eAtributos.Fuerza) - 15)) + DanoUsuario) * ModifClase
	
	var strengthBonus int
	if weapon != nil {
		strengthBonus = (weapon.MaxHit / 5) * utils.Max(0, int(attacker.Attributes[model.Strength])-15)
	} else {
		strengthBonus = (9 / 5) * utils.Max(0, int(attacker.Attributes[model.Strength])-15)
	}

	totalDmg := float32(3*weaponDmg + strengthBonus + userDmg) * modClase
	return int(totalDmg)
}

func (f *CombatFormulas) CheckCrit(char *model.Character) bool {
	// Standard AO crit chance
	return rand.Intn(100) < 10
}

func (f *CombatFormulas) CheckStab(char *model.Character) bool {
	if char.Archetype != model.Assasin {
		return false
	}
	// Simplified stab chance
	return rand.Intn(100) < (char.Skills[model.Stab] / 5)
}