package model

import mySTO "gofiber/cache"

const (
	KeyPlayerDisplayNamePrefix = "dname/"
)

func CreateDisplayNameToCache(playerUUID string, displayName string) error {
	return mySTO.Create(KeyPlayerDisplayNamePrefix+playerUUID, []byte(displayName), 0)
}
