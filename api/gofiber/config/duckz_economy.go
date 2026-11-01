package config

import (
	"encoding/json"
	"fmt"
	cfgMDL "gofiber/model/config"
	"os"
)

var (
	DuckzEconomyJsonPath = "config/json/duckz_economy.json"
)

func LoadDuckzEconomyJson() (eco cfgMDL.DuckzEconomy, err error) {
	byts, err := LoadJSON(DuckzEconomyJsonPath)
	if err != nil {
		err = fmt.Errorf("load DuckzEconomy json error: %w", err)
		return
	}

	if err = json.Unmarshal(byts, &eco); err != nil {
		err = fmt.Errorf("DuckzEconomy json error: %w", err)
		return
	}

	if eco.CoinsPerBar <= 0 || eco.MaxBarsPerDeposit <= 0 || eco.MaxCoinsPerHour <= 0 {
		err = fmt.Errorf("DuckzEconomy json: coins_per_bar, max_bars_per_deposit and max_coins_per_hour must be > 0")
		return
	}
	keys := make(map[string]bool, len(eco.StoreItems))
	for _, it := range eco.StoreItems {
		if it.Key == "" || it.Price < 0 {
			err = fmt.Errorf("DuckzEconomy json: invalid store item %+v", it)
			return
		}
		keys[it.Key] = true
	}
	unlockedBy := map[string]string{}
	for _, it := range eco.IAPSkus {
		if it.Sku == "" {
			err = fmt.Errorf("DuckzEconomy json: iap sku with no sku %+v", it)
			return
		}
		for _, key := range it.Items {
			if !keys[key] {
				err = fmt.Errorf("DuckzEconomy json: iap sku %q unlocks %q, which is not in store_items", it.Sku, key)
				return
			}
			if other, dup := unlockedBy[key]; dup {
				err = fmt.Errorf("DuckzEconomy json: %q is unlocked by both %q and %q", key, other, it.Sku)
				return
			}
			unlockedBy[key] = it.Sku
		}
	}

	eco.BuildIndex()
	return
}

func LoadJSON(filename string) ([]byte, error) {
	byts, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if !json.Valid(byts) {
		return nil, fmt.Errorf("invalid JSON in file: %s", filename)
	}

	return byts, nil
}
