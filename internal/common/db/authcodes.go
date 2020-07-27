// Authentication codes are generated by (at *AuthTable) NewAuthCode() when a new player
// joins the Minecraft server. Players can redeem their authentication code through the
// Discord bot using the "auth" command. This is done through (bot *Bot) cmdAuth().
package db

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"strings"
)

type AuthTable struct {
	db  *sql.DB
	gDB *gorm.DB
}

type AuthCode struct {
	// Pending authentication code
	AuthCode string `gorm:"column:auth_code;type:text;unique;not null"`
	// The Minecraft player associated with the pending code
	PlayerID string `gorm:"column:player_id;type:text;unique;not null"`
}

func (AuthCode) TableName() string {
	return schema + ".auth_codes"
}

// This will setup the table if it doesn't exist
func GetAuthTable(gDB *gorm.DB) AuthTable {
	gDB.AutoMigrate(&AuthCode{})

	return AuthTable{
		db:  gDB.DB(),
		gDB: gDB,
	}
}

// This will get all the pending authentication codes in the table.
func (at *AuthTable) GetAllAuthCodes() (authCodes []AuthCode, err error) {
	err = at.gDB.
		Find(&authCodes).
		Error
	return authCodes, err
}

// This will create a new authentication code for a given player UUID. If the player
// already has an authentication code then their pending one will be returned instead.
func (at *AuthTable) NewAuthCode(playerID string) (authCode string, err error) {
	// get their pending authentication code if it exists
	oldAuthCode, _ := at.GetAuthCode(playerID)

	if len(oldAuthCode) > 0 {
		return oldAuthCode, nil
	}

	newUUID := uuid.New()
	authCode = strings.Split(newUUID.String(), "-")[0]
	err = at.gDB.Create(&AuthCode{
		AuthCode: authCode,
		PlayerID: playerID,
	}).Error

	if err != nil {
		return "", err
	} else {
		return authCode, nil
	}
}

// This will get a pending authentication code of a player UUID. If it doesn't exist
// than an empty string will be returned
func (at *AuthTable) GetAuthCode(playerID string) (authCode string, err error) {
	result := AuthCode{
		AuthCode: "",
	}

	err = at.gDB.
		Find(&result, "player_id = ?", playerID).
		Error

	return result.AuthCode, err
}

// Authorize a given authentication code. It will return the player ID associated with the given
// auth code and the returned bool will be false.
func (at *AuthTable) Authorize(authCode string) (playerID string, isOK bool) {
	playerID, _ = at.GetPlayerID(authCode)

	// see if they have an authentication code
	if len(playerID) > 0 {
		isOK = true
		// remove them from the database
		go at.RemoveCode(authCode)

		return playerID, isOK
	} else {
		isOK = false
		return playerID, false
	}
}

// Get the player ID associated with the given authentication code.
func (at *AuthTable) GetPlayerID(authCode string) (playerID string, err error) {
	result := AuthCode{
		PlayerID: "",
	}

	err = at.gDB.
		Find(&result, "auth_code = ?", authCode).
		Error

	return result.PlayerID, err
}

// This will remove an authentication code given. The bool returned represents
// if the authentication code removed was removed.
func (at *AuthTable) RemoveCode(authCode string) (err error) {
	err = at.gDB.
		Where("auth_code = ? ", authCode).
		Delete(AuthCode{}).
		Error

	return err
}
