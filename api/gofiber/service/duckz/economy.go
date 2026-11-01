package duckz

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	myERR "gofiber/errors"
	myMDL "gofiber/model"
	cfgMDL "gofiber/model/config"
	dzMDL "gofiber/model/duckz"
	dzREPO "gofiber/repository/duckz"
)

const (
	logPrefix            = "duckz.economy"
	maxRequestIDLen      = 64
	depositWindow        = time.Hour
	maxAdminAdjustAmount = 1000000
	adminAdjustListLimit = 50

	adminTransactionListLimit = 500
)

type AdminAdjustRequest struct {
	PlayerID string `json:"player_id"`
	Amount   int    `json:"amount"`
}

type AdminAdjustResult struct {
	Coins  int `json:"coins"`
	Amount int `json:"amount"`
}

type AdminWalletResult struct {
	WalletResult
	Adjustments []dzMDL.CoinAdjustment `json:"adjustments"`
}

type DepositRequest struct {
	RequestID string `json:"request_id"`
	Bars      int    `json:"bars"`
}

type BuyRequest struct {
	RequestID string `json:"request_id"`
	ItemKey   string `json:"item_key"`
}

type IAPGrantRequest struct {
	Sku           string `json:"sku"`
	PlatformTxnID string `json:"platform_txn_id"`
}

type WalletResult struct {
	Coins        int      `json:"coins"`
	Entitlements []string `json:"entitlements"`
	Purchases    []string `json:"purchases"`
	// EquippedSkin is the saved avatar key; "" is the default duck.
	EquippedSkin string `json:"equipped_skin"`
	// HasLoadout is false until the player first saves a loadout, so clients can upload one saved before this existed.
	HasLoadout bool `json:"has_loadout"`
}

type LoadoutRequest struct {
	SkinKey string `json:"skin_key"`
}

type LoadoutResult struct {
	SkinKey string `json:"skin_key"`
}

// StoreListResult is the coin store the server loaded from duckz_economy.json.
type StoreListResult struct {
	Items []cfgMDL.DuckzStoreItem `json:"items"`
}

type DepositResult struct {
	Coins      int  `json:"coins"`
	Credited   int  `json:"credited"`
	Idempotent bool `json:"idempotent"`
}

type BuyResult struct {
	Coins        int    `json:"coins"`
	ItemKey      string `json:"item_key"`
	AlreadyOwned bool   `json:"already_owned"`
	Idempotent   bool   `json:"idempotent"`
}

type IAPGrantResult struct {
	Entitlements []string `json:"entitlements"`
	Idempotent   bool     `json:"idempotent"`
}

// EntitlementVerifier confirms with the platform store that the player owns a durable SKU.
type EntitlementVerifier interface {
	Verify(playerID, platform, sku string) error
}

type EconomyService interface {
	Wallet(playerID string) (*WalletResult, error)
	Deposit(playerID string, req DepositRequest, now time.Time) (*DepositResult, error)
	Buy(playerID string, req BuyRequest, now time.Time) (*BuyResult, error)
	GrantIAP(playerID, platform string, req IAPGrantRequest, now time.Time) (*IAPGrantResult, error)
	// SaveLoadout stores the equipped avatar. The key must be "" or an item the player owns.
	SaveLoadout(playerID string, req LoadoutRequest, now time.Time) (*LoadoutResult, error)
	AdminWallet(playerID string) (*AdminWalletResult, error)
	AdminAdjust(adminID string, req AdminAdjustRequest) (*AdminAdjustResult, error)
	AdminTransactions(playerID string) ([]dzMDL.AdminTransaction, error)
	// Store returns coin prices from duckz_economy.json. That file is the only price list. IAP-only items are left out.
	Store() StoreListResult
}

type economyServiceImpl struct {
	repo     dzREPO.EconomyRepository
	eco      cfgMDL.DuckzEconomy
	verifier EntitlementVerifier
}

func NewEconomyService(repo dzREPO.EconomyRepository, eco cfgMDL.DuckzEconomy, verifier EntitlementVerifier) EconomyService {
	eco.BuildIndex()
	return &economyServiceImpl{repo: repo, eco: eco, verifier: verifier}
}

func validRequestID(id string) bool {
	return id != "" && len(id) <= maxRequestIDLen
}

func (s *economyServiceImpl) Wallet(playerID string) (*WalletResult, error) {
	coins, err := s.repo.Balance(playerID)
	if err != nil {
		return nil, err
	}
	ents, err := s.repo.ListEntitlementSkus(playerID)
	if err != nil {
		return nil, err
	}
	purchases, err := s.repo.ListPurchasedKeys(playerID)
	if err != nil {
		return nil, err
	}
	skin, hasLoadout, err := s.repo.GetLoadoutSkin(playerID)
	if err != nil {
		return nil, err
	}
	return &WalletResult{Coins: coins, Entitlements: ents, Purchases: purchases, EquippedSkin: skin, HasLoadout: hasLoadout}, nil
}

func (s *economyServiceImpl) SaveLoadout(playerID string, req LoadoutRequest, now time.Time) (*LoadoutResult, error) {
	key := strings.TrimSpace(req.SkinKey)
	if key != "" {
		_, isItem := s.eco.StoreItem(key)
		_, isSku := s.eco.IAPSku(key)
		if !isItem && !isSku {
			return nil, myERR.ErrDuckzUnknownItem
		}
		owned, err := s.owns(playerID, key)
		if err != nil {
			return nil, err
		}
		if !owned {
			return nil, myERR.ErrDuckzItemNotOwned
		}
	}
	if err := s.repo.SaveLoadoutSkin(playerID, key, now); err != nil {
		return nil, err
	}
	return &LoadoutResult{SkinKey: key}, nil
}

// owns is true for coin purchases and IAP skus. Items unlocked by an IAP sku count only through that sku.
func (s *economyServiceImpl) owns(playerID, key string) (bool, error) {
	ents, err := s.repo.ListEntitlementSkus(playerID)
	if err != nil {
		return false, err
	}
	if sku, iapOnly := s.eco.IAPSkuForItem(key); iapOnly {
		return containsString(ents, sku), nil
	}
	if containsString(ents, key) {
		return true, nil
	}
	purchases, err := s.repo.ListPurchasedKeys(playerID)
	if err != nil {
		return false, err
	}
	return containsString(purchases, key), nil
}

func (s *economyServiceImpl) Deposit(playerID string, req DepositRequest, now time.Time) (*DepositResult, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	if !validRequestID(req.RequestID) {
		return nil, myERR.ErrDuckzInvalidRequestID
	}
	if req.Bars <= 0 || req.Bars > s.eco.MaxBarsPerDeposit {
		return nil, myERR.ErrDuckzInvalidBars
	}

	if res, err := s.idempotentDeposit(playerID, req.RequestID); res != nil || err != nil {
		return res, err
	}

	coins := req.Bars * s.eco.CoinsPerBar
	result := &DepositResult{Credited: coins}

	err := s.repo.WithTx(playerID, func(tx dzREPO.EconomyTx) error {
		recent, err := tx.SumDepositCoinsSince(playerID, now.Add(-depositWindow))
		if err != nil {
			return err
		}
		if recent+coins > s.eco.MaxCoinsPerHour {
			return myERR.ErrDuckzDepositCapReached
		}

		tr := myMDL.Transaction{TransactionCreation: myMDL.TransactionCreation{
			PlayerID:      playerID,
			Credit:        coins,
			Platform:      sql.NullString{String: dzMDL.TransactionPlatformGoldDeposit, Valid: true},
			PlatformTxnID: sql.NullString{String: req.RequestID, Valid: true},
		}}
		if err = tx.CreateTransaction(&tr); err != nil {
			return err
		}
		trID := tr.ID
		if err = tx.CreateDeposit(&dzMDL.CoinDeposit{
			PlayerID:      playerID,
			RequestID:     req.RequestID,
			Bars:          req.Bars,
			Coins:         coins,
			TransactionID: &trID,
			CreatedAt:     now,
		}); err != nil {
			return err
		}
		result.Coins, err = tx.Balance(playerID)
		return err
	})
	if errors.Is(err, myERR.ErrDuckzDuplicate) {
		return s.idempotentDeposit(playerID, req.RequestID)
	}
	if err != nil {
		log.Printf("[%s] deposit failed player=%s bars=%d err=%v", logPrefix, playerID, req.Bars, err)
		return nil, err
	}

	log.Printf("[%s] deposit ok player=%s bars=%d coins=%d balance=%d", logPrefix, playerID, req.Bars, coins, result.Coins)
	return result, nil
}

func (s *economyServiceImpl) idempotentDeposit(playerID, requestID string) (*DepositResult, error) {
	existing, err := s.repo.GetDepositByRequestID(requestID)
	if errors.Is(err, myERR.NoMatch) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if existing.PlayerID != playerID {
		return nil, myERR.ErrDuckzRequestIDOtherPlayer
	}
	bal, err := s.repo.Balance(playerID)
	if err != nil {
		return nil, err
	}
	return &DepositResult{Coins: bal, Credited: existing.Coins, Idempotent: true}, nil
}

func (s *economyServiceImpl) Buy(playerID string, req BuyRequest, now time.Time) (*BuyResult, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.ItemKey = strings.TrimSpace(req.ItemKey)
	if !validRequestID(req.RequestID) {
		return nil, myERR.ErrDuckzInvalidRequestID
	}
	item, ok := s.eco.StoreItem(req.ItemKey)
	if !ok {
		return nil, myERR.ErrDuckzUnknownItem
	}
	if _, iapOnly := s.eco.IAPSkuForItem(item.Key); iapOnly {
		return nil, myERR.ErrDuckzItemNotForSale
	}

	if res, err := s.idempotentBuy(playerID, req); res != nil || err != nil {
		return res, err
	}

	result := &BuyResult{ItemKey: item.Key}
	err := s.repo.WithTx(playerID, func(tx dzREPO.EconomyTx) error {
		bal, err := tx.Balance(playerID)
		if err != nil {
			return err
		}

		if item.Durable {
			owned, err := tx.CountPurchases(playerID, item.Key)
			if err != nil {
				return err
			}
			if owned > 0 {
				result.AlreadyOwned = true
				result.Coins = bal
				return nil
			}
		}

		if bal < item.Price {
			result.Coins = bal
			return myERR.ErrDuckzInsufficientCoins
		}

		tr := myMDL.Transaction{TransactionCreation: myMDL.TransactionCreation{
			PlayerID:        playerID,
			Debit:           item.Price,
			Platform:        sql.NullString{String: dzMDL.TransactionPlatformStorePurchase, Valid: true},
			PlatformAccount: sql.NullString{String: item.Key, Valid: true},
			PlatformTxnID:   sql.NullString{String: req.RequestID, Valid: true},
		}}
		if err = tx.CreateTransaction(&tr); err != nil {
			return err
		}
		trID := tr.ID
		if err = tx.CreatePurchase(&dzMDL.StorePurchase{
			PlayerID:      playerID,
			ItemKey:       item.Key,
			Price:         item.Price,
			TransactionID: &trID,
			RequestID:     req.RequestID,
			CreatedAt:     now,
		}); err != nil {
			return err
		}
		result.Coins = bal - item.Price
		return nil
	})
	if errors.Is(err, myERR.ErrDuckzDuplicate) {
		return s.idempotentBuy(playerID, req)
	}
	if err != nil {
		if !errors.Is(err, myERR.ErrDuckzInsufficientCoins) {
			log.Printf("[%s] buy failed player=%s item=%s err=%v", logPrefix, playerID, item.Key, err)
		}
		return result, err
	}

	log.Printf("[%s] buy ok player=%s item=%s price=%d owned=%v balance=%d", logPrefix, playerID, item.Key, item.Price, result.AlreadyOwned, result.Coins)
	return result, nil
}

func (s *economyServiceImpl) idempotentBuy(playerID string, req BuyRequest) (*BuyResult, error) {
	existing, err := s.repo.GetPurchaseByRequestID(req.RequestID)
	if errors.Is(err, myERR.NoMatch) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if existing.PlayerID != playerID {
		return nil, myERR.ErrDuckzRequestIDOtherPlayer
	}
	bal, err := s.repo.Balance(playerID)
	if err != nil {
		return nil, err
	}
	return &BuyResult{Coins: bal, ItemKey: existing.ItemKey, Idempotent: true}, nil
}

func (s *economyServiceImpl) GrantIAP(playerID, platform string, req IAPGrantRequest, now time.Time) (*IAPGrantResult, error) {
	req.Sku = strings.TrimSpace(req.Sku)
	req.PlatformTxnID = strings.TrimSpace(req.PlatformTxnID)
	if _, ok := s.eco.IAPSku(req.Sku); !ok {
		return nil, myERR.ErrDuckzUnknownSku
	}
	if req.PlatformTxnID == "" || strings.HasPrefix(req.PlatformTxnID, "<") {
		return nil, myERR.ErrDuckzInvalidPlatformTxnID
	}

	owned, err := s.repo.ListEntitlementSkus(playerID)
	if err != nil {
		return nil, err
	}
	if containsString(owned, req.Sku) {
		return &IAPGrantResult{Entitlements: owned, Idempotent: true}, nil
	}

	if err = s.verifier.Verify(playerID, platform, req.Sku); err != nil {
		log.Printf("[%s] iap verify failed player=%s sku=%s err=%v", logPrefix, playerID, req.Sku, err)
		return nil, err
	}

	err = s.repo.CreateEntitlement(&dzMDL.Entitlement{
		PlayerID:      playerID,
		Sku:           req.Sku,
		Platform:      platform,
		PlatformTxnID: req.PlatformTxnID,
		CreatedAt:     now,
	})
	idempotent := false
	if errors.Is(err, myERR.ErrDuckzDuplicate) {
		idempotent = true
	} else if err != nil {
		return nil, err
	}

	owned, err = s.repo.ListEntitlementSkus(playerID)
	if err != nil {
		return nil, err
	}
	if idempotent && !containsString(owned, req.Sku) {
		return nil, myERR.ErrDuckzPlatformTxnIDUsed
	}

	log.Printf("[%s] iap grant ok player=%s sku=%s txn=%s idempotent=%v", logPrefix, playerID, req.Sku, req.PlatformTxnID, idempotent)
	return &IAPGrantResult{Entitlements: owned, Idempotent: idempotent}, nil
}

func (s *economyServiceImpl) AdminWallet(playerID string) (*AdminWalletResult, error) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return nil, myERR.ErrDuckzPlayerNotFound
	}
	wallet, err := s.Wallet(playerID)
	if err != nil {
		return nil, err
	}
	adjustments, err := s.repo.ListAdjustments(playerID, adminAdjustListLimit)
	if err != nil {
		return nil, err
	}
	return &AdminWalletResult{WalletResult: *wallet, Adjustments: adjustments}, nil
}

// AdminAdjust grants (positive amount) or removes (negative) duck coins. Removals can't take the balance below zero.
func (s *economyServiceImpl) AdminAdjust(adminID string, req AdminAdjustRequest) (*AdminAdjustResult, error) {
	req.PlayerID = strings.TrimSpace(req.PlayerID)
	if req.PlayerID == "" {
		return nil, myERR.ErrDuckzPlayerNotFound
	}
	if req.Amount == 0 || req.Amount > maxAdminAdjustAmount || req.Amount < -maxAdminAdjustAmount {
		return nil, myERR.ErrDuckzInvalidAmount
	}

	result := &AdminAdjustResult{Amount: req.Amount}
	err := s.repo.WithTx(req.PlayerID, func(tx dzREPO.EconomyTx) error {
		bal, err := tx.Balance(req.PlayerID)
		if err != nil {
			return err
		}
		if bal+req.Amount < 0 {
			result.Coins = bal
			return myERR.ErrDuckzInsufficientCoins
		}

		creation := myMDL.TransactionCreation{
			PlayerID:        req.PlayerID,
			Platform:        sql.NullString{String: dzMDL.TransactionPlatformAdminAdjust, Valid: true},
			PlatformAccount: sql.NullString{String: adminID, Valid: adminID != ""},
		}
		if req.Amount > 0 {
			creation.Credit = req.Amount
		} else {
			creation.Debit = -req.Amount
		}
		tr := myMDL.Transaction{TransactionCreation: creation}
		if err = tx.CreateTransaction(&tr); err != nil {
			return err
		}
		result.Coins = bal + req.Amount
		return nil
	})
	if err != nil {
		log.Printf("[%s] admin adjust failed admin=%s player=%s amount=%d err=%v", logPrefix, adminID, req.PlayerID, req.Amount, err)
		return result, err
	}

	log.Printf("[%s] admin adjust ok admin=%s player=%s amount=%d balance=%d", logPrefix, adminID, req.PlayerID, req.Amount, result.Coins)
	return result, nil
}

func (s *economyServiceImpl) AdminTransactions(playerID string) ([]dzMDL.AdminTransaction, error) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return nil, myERR.ErrDuckzPlayerNotFound
	}
	return s.repo.ListTransactions(playerID, adminTransactionListLimit)
}

func (s *economyServiceImpl) Store() StoreListResult {
	items := make([]cfgMDL.DuckzStoreItem, 0, len(s.eco.StoreItems))
	for _, it := range s.eco.StoreItems {
		if _, iapOnly := s.eco.IAPSkuForItem(it.Key); !iapOnly {
			items = append(items, it)
		}
	}
	return StoreListResult{Items: items}
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
