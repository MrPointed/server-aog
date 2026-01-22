# TODO: Mirror Java Server Implementation in Go

This list outlines the tasks required to achieve feature parity with the existing Java server implementation.

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
- [ ] **`EquipItem` Packet:** Handle equipping items from inventory.
- [ ] **`UseItem` Packet:** Handle item usage (potions, food, etc.).
- [ ] **`Attack` Packet:** Handle physical attacks.
- [ ] **`CastSpell` Packet:** Handle casting spells on targets.
- [ ] **`PickUp` Packet:** Handle picking up items from the ground.

## 3. Networking & Protocol (Outgoing)
- [x] **State Sync Packets:** `UpdateUserStats`, `UpdateHungerAndThirst`, and `UpdateStrengthAndDexterity`.
- [x] **Dice Roll Packet:** `DiceRollPacket` for character creation.
- [x] **Spawning Handshake Sequence:** (Required for in-game entry)
    - [x] `ChangeMapPacket`: Triggers map loading.
    - [x] `UserIndexInServerPacket`: Sets the player's unique server ID.
    - [x] `CharacterCreatePacket`: Renders characters (including self).
    - [x] `AreaChangedPacket`: Syncs the initial visual area.
- [ ] **Map & World Packets:** `CharacterRemove` and `CharacterMove`.
- [ ] **Object Packets:** `ObjectCreate`, `ObjectDelete`.
- [ ] **Inventory & Spells:** `ChangeInventorySlot` and `ChangeSpellSlot`.

## 4. Services & Business Logic
- [x] **`LoginService`:** Full logic for new and existing connections.
- [x] **`MapService`:** Map loading and player positioning.
- [x] **`UserService`:** Registry of logged-in users.
- [x] **`CharacterBodyService`:** Validation of head/body combinations.
- [ ] **`AreaService`:** Implement the 3x3 visual area broadcasting logic.
- [ ] **`ObjectService`:** Factory for creating and managing world objects.
- [ ] **`NpcService`:** Logic for NPC management and behavior.
- [ ] **`CombatService`:** Implementation of combat formulas (hit/miss, damage).
- [ ] **`TimedEventsService`:** Handle periodic updates (HP/Mana regen, spell durations).

## 5. Data Access (Persistence)
- [/] **`AccountDAO`:** MemoryDAO implemented. Need persistent implementation (JSON/DB).
- [/] **`UserCharacterDAO`:** MemoryDAO implemented. Need persistent implementation.
- [x] **Map Loading:** Binary loader for `.map` and `.inf` files implemented.
- [ ] **Object Loading:** Implement `ObjectDAO` to load `objects.dat`.
- [ ] **NPC Loading:** Implement `NpcDAO` to load `npcs.dat`.
- [ ] **City Loading:** Implement `CityDAO` to load `cities.dat`.

## 6. Security & Infrastructure
- [x] **Hashing:** MD5 hashing for passwords implemented.
- [x] **Server Configuration:** `config.go` implemented.
- [x] **Action System:** Generic `ActionExecutor` implemented using Go generics.

## 7. Items & Inventory System
- [ ] **Item Models:** Implement `MeleeWeapon`, `RangedWeapon`, `Armor`, `Helmet`, `Shield`, `Potion`, `Food`, `Drink`, `Source` (Tree/Mine).
- [ ] **Inventory Logic:** Slot management, weight calculations, and item stacking.
- [ ] **Object Interaction:** Picking up, dropping, and using objects on the map (Doors, Signs, Forges).

## 8. Spells & Combat System
- [ ] **Spell Models:** Implement spell properties (mana cost, target type, effects).
- [ ] **Combat Formulas:** Implement race/archetype modifiers for combat.
- [ ] **Health & Status:** Handle death, resurrection, paralysis, and poisoning.

## 9. NPCs & AI
- [ ] **AI System:** Basic pathfinding and hostile/friendly behaviors.
- [ ] **Spawning Logic:** Map-based spawning from configuration files.
- [ ] **Loot System:** Handling NPC death and item drops.

## 10. Social & World Interactions
- [ ] **Guilds:** Basic guild creation and chat.
- [ ] **Parties:** Party formation and experience sharing.
- [ ] **Bank & Shop:** Interaction with Banker and Merchant NPCs.
- [ ] **Skills:** Implement the 21 skills (Magic, Combat, Crafting, etc.) and their progression.