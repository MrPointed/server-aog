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
	return int(float32(char.Skills[model.Defense])*mod.Shield) / 2
}

func (f *CombatFormulas) GetEvasionPower(char *model.Character) int {
	mod := f.archetypeModifiers[char.Archetype]
	if mod == nil {
		return 0
	}

	skillTactics := char.Skills[model.CombatTactics]
	agility := int(char.Attributes[model.Dexterity])

	lTemp := float32(skillTactics+skillTactics/33*agility) * mod.Evasion
	return int(lTemp + (2.5 * float32(utils.Max(int(char.Level)-12, 0))))
}

func (f *CombatFormulas) GetAttackPower(char *model.Character, weapon *model.Object) int {
	mod := f.archetypeModifiers[char.Archetype]
	if mod == nil {
		return 0
	}

	var power int
	if weapon == nil {
		power = f.calculateAttackPower(char, char.Skills[model.Wrestling], mod.WrestlingAttack)
	} else if weapon.Ranged {
		power = f.calculateAttackPower(char, char.Skills[model.Projectiles], mod.ProjectileAttack) + weapon.MaxHit
	} else {
		power = f.calculateAttackPower(char, char.Skills[model.MeleeCombat], mod.MeleeAttack) + weapon.MaxHit
	}

	return power
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
	var maxWeaponDmg int
	var modClase float32

	if weapon != nil {
		weaponDmg = utils.RandomNumber(weapon.MinHit, weapon.MaxHit)
		maxWeaponDmg = weapon.MaxHit
		if weapon.Ranged {
			modClase = mod.ProjectileDamage
		} else {
			modClase = mod.MeleeDamage
		}
	} else {
		// Wrestling damage (base 4-9)
		minWrestling := 4
		maxWrestling := 9

		// TODO: Add gloves bonus if applicable (currently in ring slot in AO)

		weaponDmg = utils.RandomNumber(minWrestling, maxWrestling)
		maxWeaponDmg = maxWrestling
		modClase = mod.WrestlingDamage
	}

	userDmg := utils.RandomNumber(attacker.MinHit, attacker.MaxHit)

	// Official formula: (3 * DanoArma + ((DanoMaxArma / 5) * MaximoInt(0, Fuerza - 15)) + DanoUsuario) * ModifClase

	strength := int(attacker.Attributes[model.Strength])
	strengthBonus := (maxWeaponDmg / 5) * utils.Max(0, strength-15)

	totalDmg := float32(3*weaponDmg+strengthBonus+userDmg) * modClase
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
