package model

type PrivilegeLevel byte

const (
	PrivilegeUser       PrivilegeLevel = 0
	PrivilegeCounselor  PrivilegeLevel = 1
	PrivilegeSemiGod    PrivilegeLevel = 2
	PrivilegeGod        PrivilegeLevel = 3
	PrivilegeAdmin      PrivilegeLevel = 4
	PrivilegeRoleMaster PrivilegeLevel = 5
)

func (p PrivilegeLevel) IsGM() bool {
	return p >= PrivilegeCounselor
}

func (p PrivilegeLevel) IsGod() bool {
	return p >= PrivilegeGod
}
