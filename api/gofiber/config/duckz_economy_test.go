package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDuckzEconomyJson(t *testing.T) {
	if _, err := os.Stat(DuckzEconomyJsonPath); err != nil {
		wd, _ := os.Getwd()
		if err := os.Chdir(filepath.Dir(wd)); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(wd)
	}

	eco, err := LoadDuckzEconomyJson()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range eco.StoreItems {
		if got, ok := eco.StoreItem(it.Key); !ok || got.Price != it.Price {
			t.Fatalf("store item %q not indexed", it.Key)
		}
	}
	for _, it := range eco.IAPSkus {
		if _, ok := eco.IAPSku(it.Sku); !ok {
			t.Fatalf("iap sku %q not indexed", it.Sku)
		}
	}
}
