package duckz

import (
	"errors"
	"testing"
	"time"

	myERR "gofiber/errors"
	myMDL "gofiber/model"
	cfgMDL "gofiber/model/config"
	dzMDL "gofiber/model/duckz"
	dzREPO "gofiber/repository/duckz"
)

type fakeState struct {
	transactions []myMDL.Transaction
	deposits     []dzMDL.CoinDeposit
	purchases    []dzMDL.StorePurchase
	entitlements []dzMDL.Entitlement
}

func (s fakeState) clone() fakeState {
	return fakeState{
		transactions: append([]myMDL.Transaction(nil), s.transactions...),
		deposits:     append([]dzMDL.CoinDeposit(nil), s.deposits...),
		purchases:    append([]dzMDL.StorePurchase(nil), s.purchases...),
		entitlements: append([]dzMDL.Entitlement(nil), s.entitlements...),
	}
}

func (s *fakeState) balance(playerID string) int {
	bal := 0
	for _, t := range s.transactions {
		if t.PlayerID == playerID && !t.IsDeleted {
			bal += t.Credit - t.Debit
		}
	}
	return bal
}

// fakeRepo is an in-memory EconomyRepository; WithTx commits only when fn returns nil.
type fakeRepo struct {
	state    fakeState
	txSeq    uint
	loadouts map[string]string
}

func (r *fakeRepo) GetLoadoutSkin(playerID string) (string, bool, error) {
	skin, found := r.loadouts[playerID]
	return skin, found, nil
}

func (r *fakeRepo) SaveLoadoutSkin(playerID, skinKey string, now time.Time) error {
	if r.loadouts == nil {
		r.loadouts = map[string]string{}
	}
	r.loadouts[playerID] = skinKey
	return nil
}

func (r *fakeRepo) Balance(playerID string) (int, error) { return r.state.balance(playerID), nil }

func (r *fakeRepo) GetDepositByRequestID(id string) (*dzMDL.CoinDeposit, error) {
	for i := range r.state.deposits {
		if r.state.deposits[i].RequestID == id {
			d := r.state.deposits[i]
			return &d, nil
		}
	}
	return nil, myERR.NoMatch
}

func (r *fakeRepo) GetPurchaseByRequestID(id string) (*dzMDL.StorePurchase, error) {
	for i := range r.state.purchases {
		if r.state.purchases[i].RequestID == id {
			p := r.state.purchases[i]
			return &p, nil
		}
	}
	return nil, myERR.NoMatch
}

func (r *fakeRepo) ListPurchasedKeys(playerID string) ([]string, error) {
	keys := []string{}
	for _, p := range r.state.purchases {
		if p.PlayerID == playerID && !containsString(keys, p.ItemKey) {
			keys = append(keys, p.ItemKey)
		}
	}
	return keys, nil
}

func (r *fakeRepo) ListTransactions(playerID string, limit int) ([]dzMDL.AdminTransaction, error) {
	out := []dzMDL.AdminTransaction{}
	for i := len(r.state.transactions) - 1; i >= 0 && len(out) < limit; i-- {
		t := r.state.transactions[i]
		if t.PlayerID == playerID {
			out = append(out, dzMDL.AdminTransaction{ID: t.ID, Platform: t.Platform.String, Debit: t.Debit, Credit: t.Credit})
		}
	}
	return out, nil
}

func (r *fakeRepo) ListEntitlementSkus(playerID string) ([]string, error) {
	out := []string{}
	for _, e := range r.state.entitlements {
		if e.PlayerID == playerID {
			out = append(out, e.Sku)
		}
	}
	return out, nil
}

func (r *fakeRepo) CreateEntitlement(row *dzMDL.Entitlement) error {
	for _, e := range r.state.entitlements {
		if (e.PlayerID == row.PlayerID && e.Sku == row.Sku) || e.PlatformTxnID == row.PlatformTxnID {
			return myERR.ErrDuckzDuplicate
		}
	}
	r.state.entitlements = append(r.state.entitlements, *row)
	return nil
}

func (r *fakeRepo) ListAdjustments(playerID string, limit int) ([]dzMDL.CoinAdjustment, error) {
	out := []dzMDL.CoinAdjustment{}
	for i := len(r.state.transactions) - 1; i >= 0 && len(out) < limit; i-- {
		t := r.state.transactions[i]
		if t.PlayerID == playerID && t.Platform.String == dzMDL.TransactionPlatformAdminAdjust {
			out = append(out, dzMDL.CoinAdjustment{ID: t.ID, AdminID: t.PlatformAccount.String, Amount: t.Credit - t.Debit})
		}
	}
	return out, nil
}

func (r *fakeRepo) WithTx(playerID string, fn func(tx dzREPO.EconomyTx) error) error {
	staged := &fakeTx{repo: r, state: r.state.clone()}
	if err := fn(staged); err != nil {
		return err
	}
	r.state = staged.state
	return nil
}

type fakeTx struct {
	repo  *fakeRepo
	state fakeState
}

func (t *fakeTx) Balance(playerID string) (int, error) { return t.state.balance(playerID), nil }

func (t *fakeTx) SumDepositCoinsSince(playerID string, since time.Time) (int, error) {
	sum := 0
	for _, d := range t.state.deposits {
		if d.PlayerID == playerID && !d.CreatedAt.Before(since) {
			sum += d.Coins
		}
	}
	return sum, nil
}

func (t *fakeTx) CountPurchases(playerID, itemKey string) (int64, error) {
	var n int64
	for _, p := range t.state.purchases {
		if p.PlayerID == playerID && p.ItemKey == itemKey {
			n++
		}
	}
	return n, nil
}

func (t *fakeTx) CreateTransaction(tr *myMDL.Transaction) error {
	t.repo.txSeq++
	tr.ID = t.repo.txSeq
	t.state.transactions = append(t.state.transactions, *tr)
	return nil
}

func (t *fakeTx) CreateDeposit(row *dzMDL.CoinDeposit) error {
	for _, d := range t.state.deposits {
		if d.RequestID == row.RequestID {
			return myERR.ErrDuckzDuplicate
		}
	}
	t.state.deposits = append(t.state.deposits, *row)
	return nil
}

func (t *fakeTx) CreatePurchase(row *dzMDL.StorePurchase) error {
	for _, p := range t.state.purchases {
		if p.RequestID == row.RequestID {
			return myERR.ErrDuckzDuplicate
		}
	}
	t.state.purchases = append(t.state.purchases, *row)
	return nil
}

type fakeVerifier struct {
	err   error
	calls int
}

func (v *fakeVerifier) Verify(playerID, platform, sku string) error {
	v.calls++
	return v.err
}

const player = "p1"

func testEconomy() cfgMDL.DuckzEconomy {
	return cfgMDL.DuckzEconomy{
		CoinsPerBar:       100,
		MaxBarsPerDeposit: 5,
		MaxCoinsPerHour:   1000,
		StoreItems: []cfgMDL.DuckzStoreItem{
			{Key: "skin_gold", Price: 300, Durable: true},
			{Key: "snack", Price: 50, Durable: false},
			{Key: "skin_hero", Durable: true},
			{Key: "hero_knife", Durable: false},
		},
		IAPSkus: []cfgMDL.DuckzIAPSku{
			{Sku: "skin_royal"},
			{Sku: "hero_bundle", Items: []string{"skin_hero", "hero_knife"}},
		},
	}
}

func newTestService() (*fakeRepo, *fakeVerifier, EconomyService) {
	repo := &fakeRepo{}
	ver := &fakeVerifier{}
	return repo, ver, NewEconomyService(repo, testEconomy(), ver)
}

func mustDeposit(t *testing.T, svc EconomyService, reqID string, bars int, now time.Time) *DepositResult {
	t.Helper()
	res, err := svc.Deposit(player, DepositRequest{RequestID: reqID, Bars: bars}, now)
	if err != nil {
		t.Fatalf("deposit %s: %v", reqID, err)
	}
	return res
}

func TestStoreListsEconomyPrices(t *testing.T) {
	_, _, svc := newTestService()
	got := svc.Store()
	if len(got.Items) != 2 || got.Items[0].Key != "skin_gold" || got.Items[0].Price != 300 || !got.Items[0].Durable {
		t.Fatalf("store = %+v", got.Items)
	}
	if got.Items[1].Key != "snack" || got.Items[1].Price != 50 || got.Items[1].Durable {
		t.Fatalf("snack = %+v", got.Items[1])
	}
}

func TestDepositCreditsServerComputedCoins(t *testing.T) {
	_, _, svc := newTestService()
	res := mustDeposit(t, svc, "d1", 3, time.Now())
	if res.Credited != 300 || res.Coins != 300 {
		t.Fatalf("got credited=%d coins=%d, want 300/300", res.Credited, res.Coins)
	}
}

func TestDepositRejectsInvalidBars(t *testing.T) {
	_, _, svc := newTestService()
	for _, bars := range []int{0, -1, 6} {
		if _, err := svc.Deposit(player, DepositRequest{RequestID: "x", Bars: bars}, time.Now()); !errors.Is(err, myERR.ErrDuckzInvalidBars) {
			t.Fatalf("bars=%d err=%v, want ErrDuckzInvalidBars", bars, err)
		}
	}
}

func TestDepositDuplicateRequestIDDoesNotDoubleCredit(t *testing.T) {
	repo, _, svc := newTestService()
	now := time.Now()
	mustDeposit(t, svc, "d1", 2, now)
	res := mustDeposit(t, svc, "d1", 2, now)
	if !res.Idempotent || res.Coins != 200 {
		t.Fatalf("got idempotent=%v coins=%d, want true/200", res.Idempotent, res.Coins)
	}
	if len(repo.state.transactions) != 1 {
		t.Fatalf("transactions=%d, want 1", len(repo.state.transactions))
	}
}

func TestDepositRequestIDFromOtherPlayerRejected(t *testing.T) {
	_, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 1, time.Now())
	if _, err := svc.Deposit("p2", DepositRequest{RequestID: "d1", Bars: 1}, time.Now()); !errors.Is(err, myERR.ErrDuckzRequestIDOtherPlayer) {
		t.Fatalf("err=%v, want ErrDuckzRequestIDOtherPlayer", err)
	}
}

func TestDepositHourlyCapEnforced(t *testing.T) {
	repo, _, svc := newTestService()
	now := time.Now()
	mustDeposit(t, svc, "d1", 5, now)
	mustDeposit(t, svc, "d2", 4, now)
	if _, err := svc.Deposit(player, DepositRequest{RequestID: "d3", Bars: 2}, now); !errors.Is(err, myERR.ErrDuckzDepositCapReached) {
		t.Fatalf("err=%v, want ErrDuckzDepositCapReached", err)
	}
	if bal := repo.state.balance(player); bal != 900 {
		t.Fatalf("balance=%d, want 900 (capped deposit must roll back)", bal)
	}
	mustDeposit(t, svc, "d4", 2, now.Add(61*time.Minute))
}

func TestBuyInsufficientCoins(t *testing.T) {
	repo, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 2, time.Now())
	res, err := svc.Buy(player, BuyRequest{RequestID: "b1", ItemKey: "skin_gold"}, time.Now())
	if !errors.Is(err, myERR.ErrDuckzInsufficientCoins) {
		t.Fatalf("err=%v, want ErrDuckzInsufficientCoins", err)
	}
	if res == nil || res.Coins != 200 {
		t.Fatalf("result=%+v, want coins=200", res)
	}
	if len(repo.state.purchases) != 0 || repo.state.balance(player) != 200 {
		t.Fatalf("purchase must not be recorded or charged")
	}
}

func TestBuyDebitsAndRecordsPurchase(t *testing.T) {
	repo, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 5, time.Now())
	res, err := svc.Buy(player, BuyRequest{RequestID: "b1", ItemKey: "skin_gold"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if res.Coins != 200 || res.AlreadyOwned {
		t.Fatalf("result=%+v, want coins=200 owned=false", res)
	}
	if len(repo.state.purchases) != 1 || repo.state.purchases[0].Price != 300 {
		t.Fatalf("purchases=%+v", repo.state.purchases)
	}
	keys, _ := repo.ListPurchasedKeys(player)
	if len(keys) != 1 || keys[0] != "skin_gold" {
		t.Fatalf("keys=%v", keys)
	}
}

func TestBuyDuplicateRequestIDDoesNotDoubleCharge(t *testing.T) {
	repo, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 5, time.Now())
	for i := 0; i < 2; i++ {
		if _, err := svc.Buy(player, BuyRequest{RequestID: "b1", ItemKey: "snack"}, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	if bal := repo.state.balance(player); bal != 450 {
		t.Fatalf("balance=%d, want 450", bal)
	}
}

func TestBuyDurableItemCannotBeBoughtTwice(t *testing.T) {
	repo, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 5, time.Now())
	mustDeposit(t, svc, "d2", 5, time.Now())
	if _, err := svc.Buy(player, BuyRequest{RequestID: "b1", ItemKey: "skin_gold"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	res, err := svc.Buy(player, BuyRequest{RequestID: "b2", ItemKey: "skin_gold"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !res.AlreadyOwned || res.Coins != 700 {
		t.Fatalf("result=%+v, want already_owned with coins=700", res)
	}
	if len(repo.state.purchases) != 1 {
		t.Fatalf("purchases=%d, want 1", len(repo.state.purchases))
	}
}

func TestBuyConsumableCanBeBoughtAgain(t *testing.T) {
	repo, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 1, time.Now())
	for _, id := range []string{"b1", "b2"} {
		if _, err := svc.Buy(player, BuyRequest{RequestID: id, ItemKey: "snack"}, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	if bal := repo.state.balance(player); bal != 0 {
		t.Fatalf("balance=%d, want 0", bal)
	}
}

func TestBuyUnknownItem(t *testing.T) {
	_, _, svc := newTestService()
	if _, err := svc.Buy(player, BuyRequest{RequestID: "b1", ItemKey: "nope"}, time.Now()); !errors.Is(err, myERR.ErrDuckzUnknownItem) {
		t.Fatalf("err=%v, want ErrDuckzUnknownItem", err)
	}
}

func TestGrantIAPVerifiesAndRecords(t *testing.T) {
	_, ver, svc := newTestService()
	res, err := svc.GrantIAP(player, "OCULUS", IAPGrantRequest{Sku: "skin_royal", PlatformTxnID: "meta-1"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if res.Idempotent || len(res.Entitlements) != 1 || ver.calls != 1 {
		t.Fatalf("result=%+v calls=%d", res, ver.calls)
	}
}

func TestGrantIAPDuplicateSucceedsWithoutInsert(t *testing.T) {
	repo, ver, svc := newTestService()
	req := IAPGrantRequest{Sku: "skin_royal", PlatformTxnID: "meta-1"}
	if _, err := svc.GrantIAP(player, "OCULUS", req, time.Now()); err != nil {
		t.Fatal(err)
	}
	res, err := svc.GrantIAP(player, "OCULUS", req, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !res.Idempotent || len(repo.state.entitlements) != 1 || ver.calls != 1 {
		t.Fatalf("result=%+v entitlements=%d calls=%d", res, len(repo.state.entitlements), ver.calls)
	}
}

func TestGrantIAPNotOwnedOnPlatform(t *testing.T) {
	repo, ver, svc := newTestService()
	ver.err = myERR.ErrDuckzEntitlementNotOwned
	if _, err := svc.GrantIAP(player, "OCULUS", IAPGrantRequest{Sku: "skin_royal", PlatformTxnID: "meta-1"}, time.Now()); !errors.Is(err, myERR.ErrDuckzEntitlementNotOwned) {
		t.Fatalf("err=%v, want ErrDuckzEntitlementNotOwned", err)
	}
	if len(repo.state.entitlements) != 0 {
		t.Fatal("entitlement must not be recorded")
	}
}

func TestGrantIAPReusedTxnIDByOtherPlayer(t *testing.T) {
	_, _, svc := newTestService()
	req := IAPGrantRequest{Sku: "skin_royal", PlatformTxnID: "meta-1"}
	if _, err := svc.GrantIAP(player, "OCULUS", req, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GrantIAP("p2", "OCULUS", req, time.Now()); !errors.Is(err, myERR.ErrDuckzPlatformTxnIDUsed) {
		t.Fatalf("err=%v, want ErrDuckzPlatformTxnIDUsed", err)
	}
}

func TestGrantIAPUnknownSku(t *testing.T) {
	_, _, svc := newTestService()
	if _, err := svc.GrantIAP(player, "OCULUS", IAPGrantRequest{Sku: "nope", PlatformTxnID: "meta-1"}, time.Now()); !errors.Is(err, myERR.ErrDuckzUnknownSku) {
		t.Fatalf("err=%v, want ErrDuckzUnknownSku", err)
	}
}

func TestWalletReportsEverything(t *testing.T) {
	_, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 5, time.Now())
	if _, err := svc.Buy(player, BuyRequest{RequestID: "b1", ItemKey: "skin_gold"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GrantIAP(player, "OCULUS", IAPGrantRequest{Sku: "skin_royal", PlatformTxnID: "meta-1"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	w, err := svc.Wallet(player)
	if err != nil {
		t.Fatal(err)
	}
	if w.Coins != 200 || len(w.Purchases) != 1 || len(w.Entitlements) != 1 {
		t.Fatalf("wallet=%+v", w)
	}
}

func TestSaveLoadoutRequiresOwnership(t *testing.T) {
	_, _, svc := newTestService()
	if _, err := svc.SaveLoadout(player, LoadoutRequest{SkinKey: "skin_gold"}, time.Now()); !errors.Is(err, myERR.ErrDuckzItemNotOwned) {
		t.Fatalf("err=%v, want ErrDuckzItemNotOwned", err)
	}
	if _, err := svc.SaveLoadout(player, LoadoutRequest{SkinKey: "nope"}, time.Now()); !errors.Is(err, myERR.ErrDuckzUnknownItem) {
		t.Fatalf("err=%v, want ErrDuckzUnknownItem", err)
	}
}

func TestSaveLoadoutOwnedItemsAndDefault(t *testing.T) {
	_, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 5, time.Now())
	if _, err := svc.Buy(player, BuyRequest{RequestID: "b1", ItemKey: "skin_gold"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GrantIAP(player, "OCULUS", IAPGrantRequest{Sku: "skin_royal", PlatformTxnID: "meta-1"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if w, _ := svc.Wallet(player); w.HasLoadout {
		t.Fatalf("has_loadout before any save")
	}

	for _, key := range []string{"skin_gold", "skin_royal", ""} {
		if _, err := svc.SaveLoadout(player, LoadoutRequest{SkinKey: key}, time.Now()); err != nil {
			t.Fatalf("save %q: %v", key, err)
		}
		w, err := svc.Wallet(player)
		if err != nil {
			t.Fatal(err)
		}
		if w.EquippedSkin != key || !w.HasLoadout {
			t.Fatalf("equipped=%q, want %q", w.EquippedSkin, key)
		}
	}
}

func TestBundleItemsAreNotSoldForCoins(t *testing.T) {
	repo, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 5, time.Now())
	for _, key := range []string{"skin_hero", "hero_knife"} {
		if _, err := svc.Buy(player, BuyRequest{RequestID: "b-" + key, ItemKey: key}, time.Now()); !errors.Is(err, myERR.ErrDuckzItemNotForSale) {
			t.Fatalf("buy %s err=%v, want ErrDuckzItemNotForSale", key, err)
		}
	}
	if len(repo.state.purchases) != 0 || repo.state.balance(player) != 500 {
		t.Fatalf("bundle items must not be recorded or charged")
	}
	for _, it := range svc.Store().Items {
		if it.Key == "skin_hero" || it.Key == "hero_knife" {
			t.Fatalf("store lists bundle item %q", it.Key)
		}
	}
}

func TestBundleSkinOwnedOnlyThroughSku(t *testing.T) {
	repo, _, svc := newTestService()
	repo.state.purchases = append(repo.state.purchases, dzMDL.StorePurchase{PlayerID: player, ItemKey: "skin_hero", RequestID: "old-coin-buy"})
	if _, err := svc.SaveLoadout(player, LoadoutRequest{SkinKey: "skin_hero"}, time.Now()); !errors.Is(err, myERR.ErrDuckzItemNotOwned) {
		t.Fatalf("err=%v, want ErrDuckzItemNotOwned before the bundle is owned", err)
	}
	if _, err := svc.GrantIAP(player, "OCULUS", IAPGrantRequest{Sku: "hero_bundle", PlatformTxnID: "meta-2"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveLoadout(player, LoadoutRequest{SkinKey: "skin_hero"}, time.Now()); err != nil {
		t.Fatalf("save after bundle: %v", err)
	}
}

const admin = "admin1"

func TestAdminAdjustGrantsAndListsHistory(t *testing.T) {
	repo, _, svc := newTestService()
	res, err := svc.AdminAdjust(admin, AdminAdjustRequest{PlayerID: player, Amount: 500})
	if err != nil {
		t.Fatal(err)
	}
	if res.Coins != 500 {
		t.Fatalf("res=%+v", res)
	}
	tr := repo.state.transactions[0]
	if tr.Credit != 500 || tr.Platform.String != dzMDL.TransactionPlatformAdminAdjust || tr.PlatformAccount.String != admin {
		t.Fatalf("transaction=%+v", tr)
	}
	w, err := svc.AdminWallet(player)
	if err != nil {
		t.Fatal(err)
	}
	if w.Coins != 500 || len(w.Adjustments) != 1 || w.Adjustments[0].AdminID != admin || w.Adjustments[0].Amount != 500 {
		t.Fatalf("admin wallet=%+v", w)
	}
}

func TestAdminAdjustRemovesButNotBelowZero(t *testing.T) {
	repo, _, svc := newTestService()
	mustDeposit(t, svc, "d1", 3, time.Now())
	res, err := svc.AdminAdjust(admin, AdminAdjustRequest{PlayerID: player, Amount: -100})
	if err != nil || res.Coins != 200 {
		t.Fatalf("res=%+v err=%v, want coins=200", res, err)
	}
	res, err = svc.AdminAdjust(admin, AdminAdjustRequest{PlayerID: player, Amount: -201})
	if !errors.Is(err, myERR.ErrDuckzInsufficientCoins) || res.Coins != 200 {
		t.Fatalf("res=%+v err=%v, want ErrDuckzInsufficientCoins with coins=200", res, err)
	}
	if len(repo.state.transactions) != 2 {
		t.Fatalf("transactions=%d, want 2", len(repo.state.transactions))
	}
}

func TestAdminAdjustValidates(t *testing.T) {
	_, _, svc := newTestService()
	cases := []struct {
		req  AdminAdjustRequest
		want error
	}{
		{AdminAdjustRequest{PlayerID: player, Amount: 0}, myERR.ErrDuckzInvalidAmount},
		{AdminAdjustRequest{PlayerID: player, Amount: 1000001}, myERR.ErrDuckzInvalidAmount},
		{AdminAdjustRequest{PlayerID: "", Amount: 10}, myERR.ErrDuckzPlayerNotFound},
	}
	for _, c := range cases {
		if _, err := svc.AdminAdjust(admin, c.req); !errors.Is(err, c.want) {
			t.Fatalf("req=%+v err=%v, want %v", c.req, err, c.want)
		}
	}
}
