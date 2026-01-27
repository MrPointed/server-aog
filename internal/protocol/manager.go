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
	CP_ChangeHeading
	CP_ModifySkills
	CP_Train
	
	CP_CommerceBuy ClientPackets = 40
	CP_BankExtractItem ClientPackets = 41
	CP_CommerceSell ClientPackets = 42
	CP_BankDeposit ClientPackets = 43

	CP_ExtractGold ClientPackets = 111
	CP_DepositGold ClientPackets = 112

	CP_Online ClientPackets = 70
	CP_Quit ClientPackets = 71
	CP_GuildLeave ClientPackets = 72
	CP_Balance ClientPackets = 73
	CP_PetStay ClientPackets = 74
	CP_PetFollow ClientPackets = 75
	CP_PetRelease ClientPackets = 76
	CP_TrainList ClientPackets = 77
	CP_Rest ClientPackets = 78
	CP_Meditate ClientPackets = 79
	CP_Resurrect ClientPackets = 80
	CP_GMCommands ClientPackets = 122
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
	SP_Logged                     ServerPackets = 0
	SP_RemoveAllDialogs           ServerPackets = 1
	SP_RemoveChrDialog            ServerPackets = 2
	SP_ToggleNavigate             ServerPackets = 3
	SP_Disconnect                 ServerPackets = 4
	SP_CommerceEnd                ServerPackets = 5
	SP_BankingEnd                 ServerPackets = 6
	SP_CommerceInit               ServerPackets = 7
	SP_BankInit                   ServerPackets = 8
	SP_UserCommerceInit           ServerPackets = 9
	SP_UserCommerceEnd            ServerPackets = 10
	SP_UserOfferConfirm           ServerPackets = 11
	SP_CommerceChat               ServerPackets = 12
	SP_ShowBlacksmithForm         ServerPackets = 13
	SP_ShowCarpenterForm          ServerPackets = 14
	SP_UpdateStamina              ServerPackets = 15
	SP_UpdateMana                 ServerPackets = 16
	SP_UpdateHP                   ServerPackets = 17
	SP_UpdateGold                 ServerPackets = 18
	SP_UpdateBankGold             ServerPackets = 19
	SP_UpdateExp                  ServerPackets = 20
	SP_ChangeMap                  ServerPackets = 21
	SP_PositionUpdate             ServerPackets = 22
	SP_ChatOverHead               ServerPackets = 23
	SP_ConsoleMessage             ServerPackets = 24
	SP_GuildChat                  ServerPackets = 25
	SP_ShowMessageBox             ServerPackets = 26
	SP_UserIndexInServer          ServerPackets = 27
	SP_UserCharacterIndexInServer ServerPackets = 28
	SP_CharacterCreate            ServerPackets = 29
	SP_CharacterRemove            ServerPackets = 30
	SP_CharacterChangeNickname    ServerPackets = 31
	SP_CharacterMove              ServerPackets = 32
	SP_CharacterForceMove         ServerPackets = 33
	SP_CharacterChange            ServerPackets = 34
	SP_ObjectCreate               ServerPackets = 35
	SP_ObjectDelete               ServerPackets = 36
	SP_BlockPosition              ServerPackets = 37
	SP_PlayMidi                   ServerPackets = 38
	SP_PlayWave                   ServerPackets = 39
	SP_GuildList                  ServerPackets = 40
	SP_AreaChanged                ServerPackets = 41
	SP_TogglePause                ServerPackets = 42
	SP_ToggleRain                 ServerPackets = 43
	SP_CreateFx                   ServerPackets = 44
	SP_UpdateUserStats            ServerPackets = 45
	SP_WorkRequestTarget          ServerPackets = 46
	SP_ChangeInventorySlot        ServerPackets = 47
	SP_ChangeBankSlot             ServerPackets = 48
	SP_ChangeSpellSlot            ServerPackets = 49
	SP_Attributes                 ServerPackets = 50
	SP_BlacksmithWeapons          ServerPackets = 51
	SP_BlacksmithArmors           ServerPackets = 52
	SP_CarpenterObjects           ServerPackets = 53
	SP_RestOk                     ServerPackets = 54
	SP_ErrorMessage              ServerPackets = 55
	SP_Blind                      ServerPackets = 56
	SP_Dumb                       ServerPackets = 57
	SP_ShowSignal                 ServerPackets = 58
	SP_ChangeNpcInventorySlot     ServerPackets = 59
	SP_UpdateHungerAndThirst      ServerPackets = 60
	SP_Fame                       ServerPackets = 61
	SP_MiniStats                  ServerPackets = 62
	SP_LevelUp                    ServerPackets = 63
	SP_AddForumMessage            ServerPackets = 64
	SP_ShowForumMessage           ServerPackets = 65
	SP_SetInvisible               ServerPackets = 66
	SP_RollDice                   ServerPackets = 67
	SP_MeditateToggle             ServerPackets = 68
	SP_BlindNoMore                ServerPackets = 69
	SP_DumbNoMore                 ServerPackets = 70
	SP_SendSkills                 ServerPackets = 71
	SP_TrainerCreatureList        ServerPackets = 72
	SP_GuildNews                  ServerPackets = 73
	SP_OfferDetails               ServerPackets = 74
	SP_AllianceProposalsList      ServerPackets = 75
	SP_PeaceProposalsList         ServerPackets = 76
	SP_CharacterInfo              ServerPackets = 77
	SP_GuildLeaderInfo            ServerPackets = 78
	SP_GuildMemberInfo            ServerPackets = 79
	SP_GuildDetails               ServerPackets = 80
	SP_ShowGuildFoundationForm    ServerPackets = 81
	SP_ParalyzeOk                 ServerPackets = 82
	SP_ShowUserRequest            ServerPackets = 83
	SP_TradeOk                    ServerPackets = 84
	SP_BankOk                     ServerPackets = 85
	SP_ChangeUserTradeSlot        ServerPackets = 86
	SP_SendNight                  ServerPackets = 87
	SP_Pong                       ServerPackets = 88
	SP_UpdateTagAndStatus         ServerPackets = 89
	SP_SpawnList                  ServerPackets = 90
	SP_ShowSosForm                ServerPackets = 91
	SP_ShowMotdEditionForm        ServerPackets = 92
	SP_ShowGmPanelForm            ServerPackets = 93
	SP_UserNameList               ServerPackets = 94
	SP_ShowGuildAlign             ServerPackets = 95
	SP_ShowPartyForm              ServerPackets = 96
	SP_UpdateStrengthAndDexterity ServerPackets = 97
	SP_UpdateStrength             ServerPackets = 98
	SP_UpdateDexterity            ServerPackets = 99
	SP_AddSlots                   ServerPackets = 100
	SP_MultiMessage               ServerPackets = 101
	SP_StopWorking                ServerPackets = 102
	SP_CancelOfferItem            ServerPackets = 103
)

func WriteOutgoing(packet OutgoingPacket, id ServerPackets, buffer *network.DataBuffer) error {
	buffer.Put(byte(id))
	return packet.Write(buffer)
}

func GetOutgoingPacketID(packet OutgoingPacket) (ServerPackets, error) {
	switch packet.(type) {
	case *outgoing.LoggedPacket:
		return SP_Logged, nil
	case *outgoing.DisconnectPacket:
		return SP_Disconnect, nil
	case *outgoing.ConsoleMessagePacket:
		return SP_ConsoleMessage, nil
	case *outgoing.ErrorMessagePacket:
		return SP_ErrorMessage, nil
	case *outgoing.DiceRollPacket:
		return SP_RollDice, nil
	case *outgoing.UpdateUserStatsPacket:
		return SP_UpdateUserStats, nil
	case *outgoing.UpdateGoldPacket:
		return SP_UpdateGold, nil
	case *outgoing.UpdateBankGoldPacket:
		return SP_UpdateBankGold, nil
	case *outgoing.UpdateHungerAndThirstPacket:
		return SP_UpdateHungerAndThirst, nil
	case *outgoing.UpdateStrengthAndDexterityPacket:
		return SP_UpdateStrengthAndDexterity, nil
	case *outgoing.ChangeMapPacket:
		return SP_ChangeMap, nil
	case *outgoing.ChatOverHeadPacket:
		return SP_ChatOverHead, nil
	case *outgoing.PosUpdatePacket:
		return SP_PositionUpdate, nil
	case *outgoing.UserIndexInServerPacket:
		return SP_UserIndexInServer, nil
	case *outgoing.UserCharIndexInServerPacket:
		return SP_UserCharacterIndexInServer, nil
	case *outgoing.CharacterCreatePacket:
		return SP_CharacterCreate, nil
	case *outgoing.NpcCreatePacket:
		return SP_CharacterCreate, nil
	case *outgoing.CharacterRemovePacket:
		return SP_CharacterRemove, nil
	case *outgoing.CharacterMovePacket:
		return SP_CharacterMove, nil
	case *outgoing.CharacterChangePacket:
		return SP_CharacterChange, nil
	case *outgoing.NpcChangePacket:
		return SP_CharacterChange, nil
	case *outgoing.ObjectCreatePacket:
		return SP_ObjectCreate, nil
	case *outgoing.ObjectDeletePacket:
		return SP_ObjectDelete, nil
	case *outgoing.AreaChangedPacket:
		return SP_AreaChanged, nil
	case *outgoing.CreateFxPacket:
		return SP_CreateFx, nil
	case *outgoing.ChangeInventorySlotPacket:
		return SP_ChangeInventorySlot, nil
	case *outgoing.ChangeBankSlotPacket:
		return SP_ChangeBankSlot, nil
	case *outgoing.ChangeSpellSlotPacket:
		return SP_ChangeSpellSlot, nil
	case *outgoing.PlayWavePacket:
		return SP_PlayWave, nil
	case *outgoing.BlockPositionPacket:
		return SP_BlockPosition, nil
	case *outgoing.SkillRequestTargetPacket:
		return SP_WorkRequestTarget, nil
	case *outgoing.CommerceInitPacket:
		return SP_CommerceInit, nil
	case *outgoing.CommerceEndPacket:
		return SP_CommerceEnd, nil
	case *outgoing.BankInitPacket:
		return SP_BankInit, nil
	case *outgoing.BankingEndPacket:
		return SP_BankingEnd, nil
	case *outgoing.ChangeNpcInventorySlotPacket:
		return SP_ChangeNpcInventorySlot, nil
	case *outgoing.AttributesPacket:
		return SP_Attributes, nil
	case *outgoing.FamePacket:
		return SP_Fame, nil
	case *outgoing.SendSkillsPacket:
		return SP_SendSkills, nil
	case *outgoing.MeditateTogglePacket:
		return SP_MeditateToggle, nil
	case *outgoing.NavigateTogglePacket:
		return SP_ToggleNavigate, nil
	}
	return 0, fmt.Errorf("unknown outgoing packet type")
}

