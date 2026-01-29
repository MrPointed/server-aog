package service

import (
	"fmt"
	"github.com/ao-go-server/internal/model"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type BankServiceImpl struct {
	objectService  ObjectService
	messageService MessageService
	userService    UserService
}

func NewBankServiceImpl(objectService ObjectService, messageService MessageService, userService UserService) BankService {
	return &BankServiceImpl{
		objectService:  objectService,
		messageService: messageService,
		userService:    userService,
	}
}

func (s *BankServiceImpl) OpenBank(char *model.Character) {
	conn := s.userService.GetConnection(char)
	if conn == nil {
		return
	}

	// 1. Sync Bank Inventory
	for i := 0; i < model.InventorySlots; i++ {
		slot := char.BankInventory.Slots[i]
		obj := s.objectService.GetObject(slot.ObjectID)
		conn.Send(&outgoing.ChangeBankSlotPacket{
			Slot:   byte(i + 1),
			Object: obj,
			Amount: slot.Amount,
		})
	}

	// 2. Init Bank (includes Gold sync)
	conn.Send(&outgoing.BankInitPacket{Gold: char.BankGold})
}

func (s *BankServiceImpl) CloseBank(char *model.Character) {
	conn := s.userService.GetConnection(char)
	if conn == nil {
		return
	}
	conn.Send(&outgoing.BankingEndPacket{})
}

func (s *BankServiceImpl) DepositGold(char *model.Character, amount int) {
	if amount <= 0 {
		return
	}

	if amount > char.Gold {
		amount = char.Gold
	}

	char.Gold -= amount
	char.BankGold += amount
	char.SetStateChanged()

	conn := s.userService.GetConnection(char)
	if conn != nil {
		conn.Send(&outgoing.UpdateGoldPacket{Gold: char.Gold})
		conn.Send(&outgoing.UpdateBankGoldPacket{Gold: char.BankGold})
		s.messageService.SendConsoleMessage(char, fmt.Sprintf("Has depositado %d monedas de oro.", amount), outgoing.INFO)
	}
}

func (s *BankServiceImpl) ExtractGold(char *model.Character, amount int) {
	if amount <= 0 {
		return
	}

	if amount > char.BankGold {
		amount = char.BankGold
	}

	char.BankGold -= amount
	char.Gold += amount
	char.SetStateChanged()

	conn := s.userService.GetConnection(char)
	if conn != nil {
		conn.Send(&outgoing.UpdateGoldPacket{Gold: char.Gold})
		conn.Send(&outgoing.UpdateBankGoldPacket{Gold: char.BankGold})
		s.messageService.SendConsoleMessage(char, fmt.Sprintf("Has retirado %d monedas de oro.", amount), outgoing.INFO)
	}
}

func (s *BankServiceImpl) DepositItem(char *model.Character, slotIdx int, amount int) {
	if slotIdx < 1 || slotIdx > model.InventorySlots {
		return
	}
	idx := slotIdx - 1
	invSlot := &char.Inventory.Slots[idx]

	if invSlot.ObjectID == 0 || invSlot.Amount <= 0 {
		return
	}

	if amount <= 0 {
		return
	}

	if amount > invSlot.Amount {
		amount = invSlot.Amount
	}

	if invSlot.Equipped {
		s.messageService.SendConsoleMessage(char, "No puedes depositar un objeto equipado.", outgoing.INFO)
		return
	}

	// Add to bank
	objID := invSlot.ObjectID
	if s.addToBank(char, objID, amount) {
		// Remove from inventory
		invSlot.Amount -= amount
		if invSlot.Amount <= 0 {
			invSlot.ObjectID = 0
			invSlot.Amount = 0
		}
		char.SetStateChanged()
		
		s.syncInventorySlot(char, slotIdx)
		s.syncBank(char)
	}
}

func (s *BankServiceImpl) ExtractItem(char *model.Character, bankSlotIdx int, amount int) {
	if bankSlotIdx < 1 || bankSlotIdx > model.InventorySlots {
		return
	}
	idx := bankSlotIdx - 1
	bankSlot := &char.BankInventory.Slots[idx]

	if bankSlot.ObjectID == 0 || bankSlot.Amount <= 0 {
		return
	}

	if amount <= 0 {
		return
	}

	if amount > bankSlot.Amount {
		amount = bankSlot.Amount
	}

	// Add to inventory
	objID := bankSlot.ObjectID
	if char.Inventory.AddItem(objID, amount) {
		// Remove from bank
		bankSlot.Amount -= amount
		if bankSlot.Amount <= 0 {
			bankSlot.ObjectID = 0
			bankSlot.Amount = 0
		}
		char.SetStateChanged()
		
		s.syncBankSlot(char, bankSlotIdx)
		s.syncInventory(char)
	} else {
		s.messageService.SendConsoleMessage(char, "No tienes espacio en el inventario.", outgoing.INFO)
	}
}

func (s *BankServiceImpl) addToBank(char *model.Character, objectID int, amount int) bool {
	// Try to stack
	for i := 0; i < model.InventorySlots; i++ {
		if char.BankInventory.Slots[i].ObjectID == objectID {
			char.BankInventory.Slots[i].Amount += amount
			s.syncBankSlot(char, i+1)
			return true
		}
	}

	// Find empty slot
	for i := 0; i < model.InventorySlots; i++ {
		if char.BankInventory.Slots[i].ObjectID == 0 {
			char.BankInventory.Slots[i].ObjectID = objectID
			char.BankInventory.Slots[i].Amount = amount
			s.syncBankSlot(char, i+1)
			return true
		}
	}

	s.messageService.SendConsoleMessage(char, "No tienes más espacio en el banco.", outgoing.INFO)
	return false
}

func (s *BankServiceImpl) syncInventorySlot(char *model.Character, slotIdx int) {
	conn := s.userService.GetConnection(char)
	if conn == nil { return }
	
	idx := slotIdx - 1
	slot := char.Inventory.Slots[idx]
	obj := s.objectService.GetObject(slot.ObjectID)
	
	conn.Send(&outgoing.ChangeInventorySlotPacket{
		Slot:     byte(slotIdx),
		Object:   obj,
		Amount:   slot.Amount,
		Equipped: slot.Equipped,
	})
}

func (s *BankServiceImpl) syncInventory(char *model.Character) {
	conn := s.userService.GetConnection(char)
	if conn == nil { return }
	
	for i := 0; i < model.InventorySlots; i++ {
		s.syncInventorySlot(char, i+1)
	}
}

func (s *BankServiceImpl) syncBankSlot(char *model.Character, slotIdx int) {
	conn := s.userService.GetConnection(char)
	if conn == nil { return }
	
	idx := slotIdx - 1
	slot := char.BankInventory.Slots[idx]
	obj := s.objectService.GetObject(slot.ObjectID)
	
	conn.Send(&outgoing.ChangeBankSlotPacket{
		Slot:   byte(slotIdx),
		Object: obj,
		Amount: slot.Amount,
	})
}

func (s *BankServiceImpl) syncBank(char *model.Character) {
	for i := 0; i < model.InventorySlots; i++ {
		s.syncBankSlot(char, i+1)
	}
}
