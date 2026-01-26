package incoming

import (
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/service"
)

type BankEndPacket struct {
	BankService service.BankService
}

func (p *BankEndPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	char := connection.GetUser()
	if char == nil { return true, nil }
	p.BankService.CloseBank(char)
	return true, nil
}

type BankExtractItemPacket struct {
	BankService service.BankService
}

func (p *BankExtractItemPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	slot, err := buffer.Get()
	if err != nil { return false, nil }
	amount, err := buffer.GetShort()
	if err != nil { return false, nil }

	char := connection.GetUser()
	if char == nil { return true, nil }
	p.BankService.ExtractItem(char, int(slot), int(amount))
	return true, nil
}

type BankDepositPacket struct {
	BankService service.BankService
}

func (p *BankDepositPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	slot, err := buffer.Get()
	if err != nil { return false, nil }
	amount, err := buffer.GetShort()
	if err != nil { return false, nil }

	char := connection.GetUser()
	if char == nil { return true, nil }
	p.BankService.DepositItem(char, int(slot), int(amount))
	return true, nil
}

type ExtractGoldPacket struct {
	BankService service.BankService
}

func (p *ExtractGoldPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	amount, err := buffer.GetLong()
	if err != nil { return false, nil }

	char := connection.GetUser()
	if char == nil { return true, nil }
	p.BankService.ExtractGold(char, int(amount))
	return true, nil
}

type DepositGoldPacket struct {
	BankService service.BankService
}

func (p *DepositGoldPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	amount, err := buffer.GetLong()
	if err != nil { return false, nil }

	char := connection.GetUser()
	if char == nil { return true, nil }
	p.BankService.DepositGold(char, int(amount))
	return true, nil
}
