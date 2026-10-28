package config

import myDOM "gofiber/domain"

func SetupHashProvider() {
	myDOM.MyHashProvider = &myDOM.BCryptHashPROV{}
}
