package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// NonceManager manages nonces for accounts
type NonceManager struct {
	nonces map[string]string
	mu     sync.Mutex
}

func NewNonceManager() *NonceManager {
	return &NonceManager{
		nonces: make(map[string]string),
	}
}

var nonceManager = NewNonceManager()

func MetamaskNonceHandler(ctx iris.Context, db *gorm.DB) {
	var req struct {
		EthAddress string `json:"ethAddress"`
	}

	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	if req.EthAddress == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	nonce, err := nonceManager.generateNonce(req.EthAddress)
	if err != nil {
		fmt.Printf("generateNonce err:%v\n", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(iris.Map{"error": "Server error"})
		return
	}
	_ = ctx.JSON(iris.Map{"nonce": nonce})
}

func (nm *NonceManager) generateNonce(account string) (string, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	i, err := rand.Int(rand.Reader, big.NewInt(99999999))
	if err != nil {
		return "", err
	}
	nonce := fmt.Sprintf("%d", i)
	nm.nonces[account] = nonce
	return nonce, nil
}

func (nm *NonceManager) verifyNonce(account, nonce string) bool {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if storedNonce, ok := nm.nonces[account]; ok && storedNonce == nonce {
		delete(nm.nonces, account) // Invalidate the nonce
		return true
	}
	return false
}
