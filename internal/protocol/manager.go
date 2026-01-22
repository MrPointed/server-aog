package protocol

import (
	"fmt"
	"github.com/ao-go-server/internal/network"
	"github.com/ao-go-server/internal/protocol/outgoing"
)

type ClientPackets int

const (
	CP_LoginExistingCharacter ClientPackets = iota
	CP_ThrowDice
	CP_LoginNewCharacter
	CP_Talk
	CP_Yell
	CP_Whisper
	CP_Walk
	CP_RequestPositionUpdate
	CP_Attack
	CP_PickUp
	CP_SafeToggle
	CP_ResuscitationSafeToggle
	CP_RequestGuildLeaderInfo
	CP_RequestAttributes
	CP_RequestFame
	CP_RequestSkills
	CP_RequestMiniStats
	CP_CommerceEnd
	CP_UserCommerceEnd
	CP_UserCommerceConfirm
	CP_CommerceChat
	CP_BankEnd
	CP_UserCommerceOk
	CP_UserCommerceReject
	CP_Drop
	CP_CastSpell
	CP_LeftClick
	CP_Double_Click
	CP_Work
	CP_UseSpellMacro
	CP_UseItem
	CP_CraftBlacksmith
	CP_CraftCarpenter
	CP_WorkLeftClick
	CP_CreateNewGuild
	CP_SpellInfo
	CP_EquipItem
)

type ClientPacketsManager struct {
	handlers map[ClientPackets]IncomingPacket
}

func NewClientPacketsManager() *ClientPacketsManager {
	return &ClientPacketsManager{
		handlers: make(map[ClientPackets]IncomingPacket),
	}
}

func (m *ClientPacketsManager) RegisterHandler(id ClientPackets, handler IncomingPacket) {
	m.handlers[id] = handler
}

func (m *ClientPacketsManager) Handle(buffer *network.DataBuffer, connection Connection) (bool, error) {
	idByte, err := buffer.Get()
	if err != nil {
		return false, err
	}

	id := ClientPackets(idByte)
	handler, ok := m.handlers[id]
	if !ok {
		return false, fmt.Errorf("unknown client packet id: %d", id)
	}

	return handler.Handle(buffer, connection)
}

type ServerPackets int

const (
	SP_Logged ServerPackets = iota
	SP_RemoveAllDialogs
	SP_RemoveChrDialog
	SP_ToggleNavigate
	SP_Disconnect
	SP_CommerceEnd
	SP_BankingEnd
	SP_CommerceInit
	SP_BankInit
	SP_UserCommerceInit
	SP_UserCommerceEnd
	SP_UserOfferConfirm
	SP_CommerceChat
	SP_ShowBlacksmithForm
	SP_ShowCarpenterForm
	SP_UpdateStamina
	SP_UpdateMana
	SP_UpdateHP
	SP_UpdateGold
	SP_UpdateBankGold
	SP_UpdateExp
	SP_ChangeMap
	SP_PositionUpdate
	SP_ChatOverHead
	SP_ConsoleMessage
	SP_GuildChat
	SP_ShowMessageBox
	SP_UserIndexInServer
	SP_UserCharacterIndexInServer
	SP_CharacterCreate
	SP_CharacterRemove
	SP_CharacterChangeNickname
	SP_CharacterMove
	SP_CharacterForceMove
	SP_CharacterChange
	SP_ObjectCreate
	SP_ObjectDelete
	SP_BlockPosition
	SP_PlayMidi
	SP_PlayWave
	SP_GuildList
	SP_AreaChanged
	SP_TogglePause
	SP_ToggleRain
	SP_CreateFx
	SP_UpdateUserStats
	SP_WorkRequestTarget
	SP_ChangeInventorySlot
	SP_ChangeBankSlot
	SP_ChangeSpellSlot
	SP_Attributes
	SP_BlacksmithWeapons
	SP_BlacksmithArmors
	SP_CarpenterObjects
	SP_RestOk
	SP_ErrorMessage
	SP_Blind
	SP_Dumb
	SP_ShowSignal
	SP_ChangeNpcInventorySlot
	SP_UpdateHungerAndThirst
	SP_Fame
	SP_MiniStats
	SP_LevelUp
	SP_AddForumMessage
	SP_ShowForumMessage
	SP_SetInvisible
	SP_RollDice
	SP_MeditateToggle
	SP_BlindNoMore
	SP_DumbNoMore
	SP_SendSkills
	SP_TrainerCreatureList
	SP_GuildNews
	SP_OfferDetails
	SP_AllianceProposalsList
	SP_PeaceProposalsList
	SP_CharacterInfo
	SP_GuildLeaderInfo
	SP_GuildMemberInfo
	SP_GuildDetails
	SP_ShowGuildFoundationForm
	SP_ParalyzeOk
	SP_ShowUserRequest
	SP_TradeOk
	SP_BankOk
	SP_ChangeUserTradeSlot
	SP_SendNight
	SP_Pong
	SP_UpdateTagAndStatus
	SP_SpawnList
	SP_ShowSosForm
	SP_ShowMotdEditionForm
	SP_ShowGmPanelForm
	SP_UserNameList
	SP_ShowGuildAlign
	SP_ShowPartyForm
	SP_UpdateStrengthAndDexterity
	SP_UpdateStrength
	SP_UpdateDexterity
	SP_AddSlots
	SP_MultiMessage
	SP_StopWorking
	SP_CancelOfferItem
)

func WriteOutgoing(packet OutgoingPacket, id ServerPackets, buffer *network.DataBuffer) error {
	buffer.Put(byte(id))
	return packet.Write(buffer)
}

func GetOutgoingPacketID(packet OutgoingPacket) (ServerPackets, error) {
	switch packet.(type) {
	case *outgoing.LoggedPacket:
		return SP_Logged, nil
	case *outgoing.ConsoleMessagePacket:
		return SP_ConsoleMessage, nil
	case *outgoing.ErrorMessagePacket:
		return SP_ErrorMessage, nil
	case *outgoing.DiceRollPacket:
		return SP_RollDice, nil
	case *outgoing.UpdateUserStatsPacket:
		return SP_UpdateUserStats, nil
	case *outgoing.UpdateHungerAndThirstPacket:
		return SP_UpdateHungerAndThirst, nil
	case *outgoing.UpdateStrengthAndDexterityPacket:
		return SP_UpdateStrengthAndDexterity, nil
	case *outgoing.ChangeMapPacket:
		return SP_ChangeMap, nil
	case *outgoing.PosUpdatePacket:
		return SP_PositionUpdate, nil
	case *outgoing.UserCharIndexInServerPacket:
		return SP_UserCharacterIndexInServer, nil
	case *outgoing.CharacterCreatePacket:
		return SP_CharacterCreate, nil
	case *outgoing.CharacterRemovePacket:
		return SP_CharacterRemove, nil
	case *outgoing.CharacterMovePacket:
		return SP_CharacterMove, nil
	case *outgoing.CharacterChangePacket:
		return SP_CharacterChange, nil
	case *outgoing.ObjectCreatePacket:
		return SP_ObjectCreate, nil
	case *outgoing.ObjectDeletePacket:
		return SP_ObjectDelete, nil
	case *outgoing.AreaChangedPacket:
		return SP_AreaChanged, nil
	case *outgoing.ChangeInventorySlotPacket:
		return SP_ChangeInventorySlot, nil
	}
	return 0, fmt.Errorf("unknown outgoing packet type")
}

