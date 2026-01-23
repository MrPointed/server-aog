package persistence

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ao-go-server/internal/model"
)

type FileDAO struct {
	basePath string
}

func NewFileDAO(basePath string) *FileDAO {
	return &FileDAO{basePath: basePath}
}

func (d *FileDAO) getFilePath(nick string) string {
	return fmt.Sprintf("%s/%s.chr", d.basePath, strings.ToLower(nick))
}

func (d *FileDAO) Exists(nick string) bool {
	_, err := os.Stat(d.getFilePath(nick))
	return err == nil
}

func (d *FileDAO) Load(nick string) (*model.Character, error) {
	data, err := ReadINI(d.getFilePath(nick))
	if err != nil {
		return nil, err
	}
	
	char := &model.Character{
		Name: nick,
		Attributes: make(map[model.Attribute]byte),
		Skills:     make(map[model.Skill]int),
	}

	init := data["INIT"]
	stats := data["STATS"]
	attrs := data["ATRIBUTOS"]
	flags := data["FLAGS"]
	skills := data["SKILLS"]

	char.Race = model.Race(toInt(init["RAZA"]))
	char.Gender = model.Gender(toInt(init["GENERO"]))
	char.Archetype = model.UserArchetype(toInt(init["CLASE"]))
	char.Head = toInt(init["HEAD"])
	char.OriginalHead = toInt(init["ORIGINALHEAD"])
	if char.OriginalHead == 0 {
		char.OriginalHead = char.Head
	}
	char.Body = toInt(init["BODY"])
	char.Heading = model.Heading(toInt(init["HEADING"]) - 1)
	if char.Heading < 0 {
		char.Heading = model.North
	}

	char.Position.Map = toInt(init["MAP"])
	posStr := init["POSITION"]
	if posStr != "" {
		parts := strings.Split(posStr, "-")
		if len(parts) == 3 {
			char.Position.X = byte(toInt(parts[1]))
			char.Position.Y = byte(toInt(parts[2]))
		}
	}

	char.Weapon = int16(toInt(init["ARMA"]))
	char.Shield = int16(toInt(init["ESCUDO"]))
	char.Helmet = int16(toInt(init["CASCO"]))

	char.Level = byte(toInt(stats["ELV"]))
	char.Exp = toInt(stats["EXP"])
	char.ExpToNext = toInt(stats["ELU"])
	if char.ExpToNext == 0 {
		char.ExpToNext = 300 // Default for lvl 1
	}
	char.MinHit = toInt(stats["MINHIT"])
	char.MaxHit = toInt(stats["MAXHIT"])
	char.Hp = toInt(stats["MINHP"])
	char.MaxHp = toInt(stats["MAXHP"])
	char.Mana = toInt(stats["MINMAN"])
	char.MaxMana = toInt(stats["MAXMAN"])
	char.Stamina = toInt(stats["MINSTA"])
	char.MaxStamina = toInt(stats["MAXSTA"])
	char.Hunger = toInt(stats["MINHAM"])
	char.Thirstiness = toInt(stats["MINAGU"])
	char.Gold = toInt(stats["ORO"])
	char.SkillPoints = toInt(stats["SKILLPTS"])

	char.Attributes[model.Strength] = byte(toInt(attrs["AT1"]))
	char.Attributes[model.Dexterity] = byte(toInt(attrs["AT2"]))
	char.Attributes[model.Intelligence] = byte(toInt(attrs["AT3"]))
	char.Attributes[model.Charisma] = byte(toInt(attrs["AT4"]))
	char.Attributes[model.Constitution] = byte(toInt(attrs["AT5"]))

	char.Dead = toInt(flags["MUERTO"]) == 1
	char.Invisible = toInt(flags["ESCONDIDO"]) == 1
	char.Paralyzed = toInt(flags["PARALIZADO"]) == 1

	// Skills
	if skills != nil {
		for i := 1; i <= 21; i++ {
			key := fmt.Sprintf("SK%d", i)
			char.Skills[model.Skill(i)] = toInt(skills[key])
		}
	}

	// Inventory
	inventory := data["INVENTORY"]
	if inventory != nil {
		for i := 0; i < model.InventorySlots; i++ {
			key := strings.ToUpper(fmt.Sprintf("Obj%d", i+1))
			val := inventory[key]
			if val != "" {
				parts := strings.Split(val, "-")
				if len(parts) >= 2 {
					char.Inventory.Slots[i].ObjectID = toInt(parts[0])
					char.Inventory.Slots[i].Amount = toInt(parts[1])
					if len(parts) >= 3 {
						char.Inventory.Slots[i].Equipped = toInt(parts[2]) == 1
					}
				}
			}
		}
	}

	// Spells
	spellsData := data["HECHIZOS"]
	if spellsData != nil {
		for i := 1; i <= 35; i++ {
			key := fmt.Sprintf("H%d", i)
			if val, ok := spellsData[key]; ok && val != "0" {
				char.Spells = append(char.Spells, toInt(val))
			}
		}
	}

	return char, nil
}

func (d *FileDAO) Get(nick string) (*model.Account, error) {
	data, err := ReadINI(d.getFilePath(nick))
	if err != nil {
		return nil, err
	}

	init := data["INIT"]
	contacto := data["CONTACTO"]
	flags := data["FLAGS"]

	acc := &model.Account{
		Nick:     nick,
		Password: init["PASSWORD"],
		Mail:     contacto["EMAIL"],
		Banned:   toInt(flags["BAN"]) == 1,
	}
	acc.Characters = append(acc.Characters, nick)

	return acc, nil
}

func (d *FileDAO) Create(nick, password, mail string) (*model.Account, error) {
	acc := &model.Account{
		Nick:     nick,
		Password: password,
		Mail:     mail,
	}
	// Note: Account creation usually happens with character creation in this simplified model
	return acc, nil
}

func (d *FileDAO) SaveAccount(acc *model.Account) error {
	// In the classic .chr system, account data is inside the .chr
	// If we have a character for this account, we save it.
	// For now, assume account name == character name as in the current simplified logic.
	return nil 
}

func (d *FileDAO) SaveCharacter(char *model.Character) error {
	// We need the account to save the password and email
	// Since our current model is simplified, we might need to load existing file first
	// to not lose password if we don't have it in char model.
	// But let's assume we can get it from somewhere or just save char stats for now.
	
	// Better: load the existing file to keep password and email.
	data, _ := ReadINI(d.getFilePath(char.Name))
	if data == nil {
		data = make(map[string]map[string]string)
	}

	if data["INIT"] == nil { data["INIT"] = make(map[string]string) }
	if data["CONTACTO"] == nil { data["CONTACTO"] = make(map[string]string) }
	if data["FLAGS"] == nil { data["FLAGS"] = make(map[string]string) }
	if data["ATRIBUTOS"] == nil { data["ATRIBUTOS"] = make(map[string]string) }
	if data["STATS"] == nil { data["STATS"] = make(map[string]string) }

	init := data["INIT"]
	init["GENERO"] = strconv.Itoa(int(char.Gender))
	init["RAZA"] = strconv.Itoa(int(char.Race))
	init["CLASE"] = strconv.Itoa(int(char.Archetype))
	init["HEAD"] = strconv.Itoa(char.Head)
	init["ORIGINALHEAD"] = strconv.Itoa(char.OriginalHead)
	init["BODY"] = strconv.Itoa(char.Body)
	init["MAP"] = strconv.Itoa(char.Position.Map)
	init["POSITION"] = fmt.Sprintf("%d-%d-%d", char.Position.Map, char.Position.X, char.Position.Y)
	init["HEADING"] = strconv.Itoa(int(char.Heading) + 1)
	init["ARMA"] = strconv.Itoa(int(char.Weapon))
	init["ESCUDO"] = strconv.Itoa(int(char.Shield))
	init["CASCO"] = strconv.Itoa(int(char.Helmet))

	flags := data["FLAGS"]
	flags["MUERTO"] = boolToIntString(char.Dead)
	flags["ESCONDIDO"] = boolToIntString(char.Invisible)
	flags["PARALIZADO"] = boolToIntString(char.Paralyzed)

	attrs := data["ATRIBUTOS"]
	attrs["AT1"] = strconv.Itoa(int(char.Attributes[model.Strength]))
	attrs["AT2"] = strconv.Itoa(int(char.Attributes[model.Dexterity]))
	attrs["AT3"] = strconv.Itoa(int(char.Attributes[model.Intelligence]))
	attrs["AT4"] = strconv.Itoa(int(char.Attributes[model.Charisma]))
	attrs["AT5"] = strconv.Itoa(int(char.Attributes[model.Constitution]))

	stats := data["STATS"]
	stats["ELV"] = strconv.Itoa(int(char.Level))
	stats["EXP"] = strconv.Itoa(char.Exp)
	stats["MINHIT"] = strconv.Itoa(char.MinHit)
	stats["MAXHIT"] = strconv.Itoa(char.MaxHit)
	stats["MINHP"] = strconv.Itoa(char.Hp)
	stats["MAXHP"] = strconv.Itoa(char.MaxHp)
	stats["MINMAN"] = strconv.Itoa(char.Mana)
	stats["MAXMAN"] = strconv.Itoa(char.MaxMana)
	stats["MINSTA"] = strconv.Itoa(char.Stamina)
	stats["MAXSTA"] = strconv.Itoa(char.MaxStamina)
	stats["MINHAM"] = strconv.Itoa(char.Hunger)
	stats["MAXHAM"] = strconv.Itoa(100)
	stats["MinAGU"] = strconv.Itoa(char.Thirstiness)
	stats["MaxAGU"] = strconv.Itoa(100)
	stats["ORO"] = strconv.Itoa(char.Gold)

	// Skills
	if data["SKILLS"] == nil { data["SKILLS"] = make(map[string]string) }
	sk := data["SKILLS"]
	for i := 1; i <= 21; i++ {
		key := fmt.Sprintf("SK%d", i)
		sk[key] = strconv.Itoa(char.Skills[model.Skill(i)])
	}

	// Inventory
	if data["INVENTORY"] == nil { data["INVENTORY"] = make(map[string]string) }
	inv := data["INVENTORY"]
	count := 0
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.Inventory.Slots[i]
		key := strings.ToUpper(fmt.Sprintf("OBJ%d", i+1))
		if slot.ObjectID > 0 {
			inv[key] = fmt.Sprintf("%d-%d", slot.ObjectID, slot.Amount)
			count++
		} else {
			delete(inv, key)
		}
	}
	inv["CANTIDADITEMS"] = strconv.Itoa(count)

	// Spells
	if data["HECHIZOS"] == nil { data["HECHIZOS"] = make(map[string]string) }
	h := data["HECHIZOS"]
	for i := 0; i < 35; i++ {
		key := fmt.Sprintf("H%d", i+1)
		if i < len(char.Spells) {
			h[key] = strconv.Itoa(char.Spells[i])
		} else {
			h[key] = "0"
		}
	}

	return d.writeINI(d.getFilePath(char.Name), data)
}

func (d *FileDAO) CreateAccountAndCharacter(nick, password, mail string, race model.Race, gender model.Gender, 
	archetype model.UserArchetype, head int, city model.City, attributes map[model.Attribute]byte) (*model.Account, *model.Character, error) {
	
	char := model.NewCharacter(nick, race, gender, archetype)
	char.Head = head
	char.Position = model.Position{X: city.X, Y: city.Y, Map: city.Map}
	char.Attributes = attributes
	char.MinHit = 1
	char.MaxHit = 2
	char.MaxHp = 20
	char.Hp = 20
	char.MaxMana = 100
	char.Mana = 100
	char.MaxStamina = 100
	char.Stamina = 100
	char.Hunger = 100
	char.Thirstiness = 100

	acc := &model.Account{
		Nick:     nick,
		Password: password,
		Mail:     mail,
	}

	// Create initial file
	data := make(map[string]map[string]string)
	data["INIT"] = make(map[string]string)
	data["INIT"]["Password"] = password
	data["CONTACTO"] = make(map[string]string)
	data["CONTACTO"]["Email"] = mail
	
	err := d.writeINI(d.getFilePath(nick), data)
	if err != nil {
		return nil, nil, err
	}

	err = d.SaveCharacter(char)
	if err != nil {
		return nil, nil, err
	}

	return acc, char, nil
}

func (d *FileDAO) writeINI(path string, data map[string]map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	// We want some order if possible, but for simplicity let's just range
	sections := []string{"INIT", "CONTACTO", "FLAGS", "ATRIBUTOS", "STATS", "SKILLS", "REP", "INVENTORY", "HECHIZOS"}
	for _, sec := range sections {
		if inner, ok := data[sec]; ok {
			fmt.Fprintf(writer, "[%s]\n", sec)
			for k, v := range inner {
				fmt.Fprintf(writer, "%s=%s\n", k, v)
			}
			fmt.Fprintln(writer)
		}
	}
	return writer.Flush()
}

func toInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func boolToIntString(b bool) string {
	if b { return "1" }
	return "0"
}

func (d *FileDAO) readINI(path string) (map[string]map[string]string, error) {
	return ReadINI(path)
}
