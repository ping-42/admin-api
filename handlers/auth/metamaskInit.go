package auth

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/middleware"
	"gorm.io/gorm"
)

// The same msg needs to be sign with the private key via Metamsk
const MsgToSign = "Authenticate with Ping42 app:"

type LoginRequest struct {
	EthAddress string `json:"ethAddress"`
	Signature  string `json:"signature"`
	Nonce      string `json:"nonce"`
}

func MetamaskLoginHandler(ctx iris.Context, db *gorm.DB) {
	var req LoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	if req.Signature == "" || req.EthAddress == "" || req.Nonce == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	if !nonceManager.verifyNonce(req.EthAddress, req.Nonce) {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "Invalid nonce"})
		return
	}

	message := MsgToSign + req.Nonce

	if verifySignature(message, req.Signature, req.EthAddress) {
		// if the user do not exists will cerate new one automatically
		user, err := getOrCreateMetamaskUser(db, req.EthAddress)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(iris.Map{"error": "Failed to generate user"})
			fmt.Println(err) // TODO logger
			return
		}

		token, err := middleware.GenerateJWT(db, user)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(iris.Map{"error": "Failed to generate token"})
			return
		}
		_ = ctx.JSON(iris.Map{"token": token, "userGroupID": user.UserGroupID})
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "Invalid signature"})
	}
}

func verifySignature(message, signature, address string) bool {

	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefixedMessage))

	sig, err := hex.DecodeString(signature[2:]) // Remove "0x" prefix before decoding
	if err != nil {
		log.Printf("Failed to decode signature: %v\n", err)
		return false
	}

	// ensure the signature length is correct
	if len(sig) != 65 {
		log.Printf("Invalid signature length: %d\n", len(sig))
		return false
	}

	// adjust the recovery id (last byte)
	sig[64] -= 27

	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		log.Printf("Failed to recover public key: %v\n", err)
		return false
	}

	recoveredAddress := crypto.PubkeyToAddress(*pubKey).Hex()

	// normalize addresses to lowercase for case-insensitive comparison
	return strings.EqualFold(recoveredAddress, address)
}

// check if the user exists in the database, if no create new one
func getOrCreateMetamaskUser(db *gorm.DB, ethAddress string) (user models.User, err error) {

	if ethAddress == "" {
		err = fmt.Errorf("not supported case expected eth address")
		return
	}

	var result *gorm.DB
	result = db.Where("wallet_address = ?", ethAddress).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// user not found, create new organization & new admin user
		newOrg := models.Organization{
			ID:   uuid.New(),
			Name: "",
		}
		if err = db.Create(&newOrg).Error; err != nil {
			err = fmt.Errorf("creating new org err:%v", err)
			return
		}
		//
		newUser := models.User{
			ID:             uuid.New(),
			WalletAddress:  &ethAddress,
			UserGroupID:    2, // Admin
			OrganizationID: newOrg.ID,
			CreatedAt:      time.Now(),
			IsValidated:    true,
			IsActive:       true,
			LastLoginAt:    time.Now(),
		}
		if err = db.Create(&newUser).Error; err != nil {
			err = fmt.Errorf("creating new metamask user err:%v", err)
			return
		}
		user = newUser

		fmt.Printf("new user metamask created:%+v\n", newUser)
	} else {
		updateTx := db.Model(&models.User{}).Where("id = ?", user.ID).Update("last_login_at", time.Now())
		if updateTx.Error != nil {
			fmt.Println("Error updating user last_login_at", updateTx.Error) //TODO logger
			return
		}
		fmt.Printf("user:%+v initiating metamask login\n", user.ID)
	}

	return
}
