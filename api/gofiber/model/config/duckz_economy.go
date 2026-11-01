package config

// DuckzEconomy mirrors config/json/duckz_economy.json. Keys and prices must match the
// Duckz client StoreCatalog; the server copy is authoritative for prices.
type DuckzEconomy struct {
	CoinsPerBar       int                `json:"coins_per_bar"`
	MaxBarsPerDeposit int                `json:"max_bars_per_deposit"`
	MaxCoinsPerHour   int                `json:"max_coins_per_hour"`
	StoreItems        []DuckzStoreItem   `json:"store_items"`
	IAPSkus           []DuckzIAPSku      `json:"iap_skus"`
	storeByKey        map[string]DuckzStoreItem
	iapBySku          map[string]DuckzIAPSku
	skuByItem         map[string]string
}

type DuckzStoreItem struct {
	Key     string `json:"key"`
	Price   int    `json:"price"`
	Durable bool   `json:"durable"`
	Slots   int    `json:"slots,omitempty"` // bags only
}

// DuckzIAPSku is a durable Meta SKU. Items lists the store_items keys it unlocks; those are not sold for coins.
type DuckzIAPSku struct {
	Sku   string   `json:"sku"`
	Items []string `json:"items,omitempty"`
	Slots int      `json:"slots,omitempty"` // bags only
}

func (e *DuckzEconomy) BuildIndex() {
	e.storeByKey = make(map[string]DuckzStoreItem, len(e.StoreItems))
	for _, it := range e.StoreItems {
		e.storeByKey[it.Key] = it
	}
	e.iapBySku = make(map[string]DuckzIAPSku, len(e.IAPSkus))
	e.skuByItem = make(map[string]string)
	for _, it := range e.IAPSkus {
		e.iapBySku[it.Sku] = it
		for _, key := range it.Items {
			e.skuByItem[key] = it.Sku
		}
	}
}

func (e *DuckzEconomy) StoreItem(key string) (DuckzStoreItem, bool) {
	if e.storeByKey == nil {
		e.BuildIndex()
	}
	it, ok := e.storeByKey[key]
	return it, ok
}

func (e *DuckzEconomy) IAPSku(sku string) (DuckzIAPSku, bool) {
	if e.iapBySku == nil {
		e.BuildIndex()
	}
	it, ok := e.iapBySku[sku]
	return it, ok
}

// IAPSkuForItem returns the SKU that unlocks a store item, if the item is IAP-only.
func (e *DuckzEconomy) IAPSkuForItem(key string) (string, bool) {
	if e.skuByItem == nil {
		e.BuildIndex()
	}
	sku, ok := e.skuByItem[key]
	return sku, ok
}
