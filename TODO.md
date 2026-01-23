# TODO: Server Implementation in Go

This list outlines the tasks required to achieve a functional golang server implementation.

## 1. Core Models & Data Structures
- [x] **Character Attributes:** Implement `Strength`, `Dexterity`, `Intelligence`, `Charisma`, and `Constitution`.
- [x] **Enums:** Implement `Race`, `Gender`, and `UserArchetype`.
- [x] **World Model:** Implement `Position`, `Map`, `City`, `Heading`, `Trigger`, and `Tile`.
- [x] **User & Account:** Create `Account` and `Character` models.

## 2. Networking & Protocol (Incoming)
- [x] **`ThrowDice` Packet:** Implement the handler for attribute rolling.
- [x] **`Walk` Packet:** Handle character movement logic.
- [x] **`Talk/Yell/Whisper` Packets:** Complete the chat system logic.
- [x] **Packet Validation:** Integrated strict buffer size checks.
- [x] **`EquipItem` Packet:** Handle equipping items from inventory.
- [x] **`UseItem` Packet:** Handle item usage (potions, food, etc.).
- [x] **`Attack` Packet:** Handle physical attacks.
- [x] **`CastSpell` Packet:** Handle casting spells on targets.
- [x] **`PickUp` Packet:** Handle picking up items from the ground.
- [x] **`ModifySkills` Packet:** Allow players to assign skill points.

## 3. Networking & Protocol (Outgoing)
- [x] **State Sync Packets:** `UpdateUserStats`, `UpdateHungerAndThirst`, and `UpdateStrengthAndDexterity`.
- [x] **Dice Roll Packet:** `DiceRollPacket` for character creation.
- [x] **Spawning Handshake Sequence:** (Required for in-game entry)
    - [x] `ChangeMapPacket`: Triggers map loading.
    - [x] `UserIndexInServerPacket`: Sets the player's unique server ID.
    - [x] `CharacterCreatePacket`: Renders characters (including self).
    - [x] `AreaChangedPacket`: Syncs the initial visual area.
- [x] **Map & World Packets:** `CharacterRemove` and `CharacterMove`.
- [x] **Object Packets:** `ObjectCreate`, `ObjectDelete`.
- [x] **Inventory & Spells:** `ChangeInventorySlot` and `ChangeSpellSlot`.
- [x] **`SendSkills` Packet:** Send current skills and free skill points to client.

## 4. Services & Business Logic
- [x] **`LoginService`:** Full logic for new and existing connections.
- [x] **`MapService`:** Map loading and player positioning.
- [x] **`UserService`:** Registry of logged-in users.
- [x] **`CharacterBodyService`:** Validation of head/body combinations.
- [x] **`AreaService`:** Implement the 3x3 visual area broadcasting logic.
- [x] **`MessageService`:** Broadcasting and targeted data transmission.
- [x] **`ObjectService`:** Factory for creating and managing world objects.
- [x] **`NpcService`:** Logic for NPC management and behavior.
- [x] **`CombatService`:** Implementation of combat formulas (hit/miss, damage).
- [x] **`TimedEventsService`:** Handle periodic updates (HP/Mana regen, spell durations).
- [x] **`SpellService`:** Spell definitions and management.
- [x] **`TrainingService`:** Level up logic, stat gains, and skill points awarding.

## 5. Data Access (Persistence)
- [x] **`AccountDAO`:** FileDAO implemented (.chr files).
- [x] **`UserCharacterDAO`:** FileDAO implemented (.chr files).
- [x] **Map Loading:** Binary loader for `.map` and `.inf` files implemented.
- [x] **Object Loading:** Implement `ObjectDAO` to load `objects.dat`.
- [x] **NPC Loading:** Implement `NpcDAO` to load `npcs.dat`.
- [x] **City Loading:** Implement `CityDAO` to load `cities.dat`.
- [x] **Spell Loading:** Implement `SpellDAO` to load `hechizos.dat`.

## 6. Security & Infrastructure
- [x] **Hashing:** MD5 hashing for passwords implemented.
- [x] **Server Configuration:** `config.go` implemented.
- [x] **Action System:** Generic `ActionExecutor` implemented using Go generics.
- [ ] **Anticheat:** Implement interval validation for all critical actions (Move, Attack, Spell, Work).

## 7. Items & Inventory System
- [ ] **Item Models:** Implement `MeleeWeapon`, `RangedWeapon`, `Armor`, `Helmet`, `Shield`, `Potion`, `Food`, `Drink`, `Source` (Tree/Mine).
- [ ] **Inventory Logic:** Weight calculations and item stacking.
- [x] **Initial Inventory:** Newbie items granted upon character creation.
- [ ] **Object Interaction:** Logic for using objects on the map (Doors, Signs, Forges).

## 8. Spells & Combat System
- [x] **Spell Models:** Implement spell properties (mana cost, target type, effects).
- [x] **Combat Formulas:** Implement race/archetype modifiers for combat.
- [ ] **Health & Status:** Handle resurrection, paralysis, and poisoning (Death and basic HP already implemented).
- [x] **Spell Casting:** Logic and effects for damage/healing spells.

## 9. NPCs & AI
- [ ] **AI System:** Pathfinding and hostile/friendly behaviors.
- [ ] **Spawning Logic:** Map-based spawning from configuration files.
- [ ] **Loot System:** Handling NPC death and item drops.

## 10. Social & World Interactions
- [ ] **Guilds:** Creation, management, and chat.
- [ ] **Parties:** Formation, management, and experience sharing.
- [ ] **Bank & Shop:** Complete logic for Banker and Merchant interactions.
- [x] **Skills:** 21 skills implemented and assigned to client-compatible IDs.
- [x] **Skill Progression:** Allocation of skill points (10 initial, 5 per level).
- [ ] **World Triggers:** Map-to-map transitions (Warps).
- [ ] **Navigation:** Sailing system and water-only zones.
- [ ] **Admin Commands:** Full set of GM commands (/TELEP, /BAN, /KICK, /ITEM).
