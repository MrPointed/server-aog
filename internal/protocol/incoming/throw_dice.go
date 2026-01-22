package incoming

import (
	"math/rand"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol"
	"github.com/ao-go-server/internal/protocol/outgoing"
	"github.com/ao-go-server/internal/model"
)

type ThrowDicesPacket struct{}

func (p *ThrowDicesPacket) Handle(buffer *network.DataBuffer, connection protocol.Connection) (bool, error) {
	// Java logic:
	// byte strength = (byte) Math.max(MIN_STRENGTH, 13 + random.nextInt(4) + random.nextInt(3));
	// byte dexterity = (byte) Math.max(MIN_DEXTERITY, 12 + random.nextInt(4) + random.nextInt(4));
	// byte intelligence = (byte) Math.max(MIN_INGELLIGENCE, 13 + random.nextInt(4) + random.nextInt(3));
	// byte charisma = (byte) Math.max(MIN_CHARISMA, 12 + random.nextInt(4) + random.nextInt(4));
	// byte constitution = (byte) Math.max(MIN_CONSTITUTION, 16 + random.nextInt(2) + random.nextInt(2));

	strength := byte(max(15, 13+rand.Intn(4)+rand.Intn(3)))
	dexterity := byte(max(15, 12+rand.Intn(4)+rand.Intn(4)))
	intelligence := byte(max(16, 13+rand.Intn(4)+rand.Intn(3)))
	charisma := byte(max(15, 12+rand.Intn(4)+rand.Intn(4)))
	constitution := byte(max(16, 16+rand.Intn(2)+rand.Intn(2)))

	connection.SetAttribute(int(model.Strength), strength)
	connection.SetAttribute(int(model.Dexterity), dexterity)
	connection.SetAttribute(int(model.Intelligence), intelligence)
	connection.SetAttribute(int(model.Charisma), charisma)
	connection.SetAttribute(int(model.Constitution), constitution)

	connection.Send(&outgoing.DiceRollPacket{
		Strength:     strength,
		Dexterity:    dexterity,
		Intelligence: intelligence,
		Charisma:     charisma,
		Constitution: constitution,
	})

	return true, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
