package config

import (
	"log"
	"os"
	"strings"
)

var (
	//ignoreBuildHashCheck   map[string]interface{}
	platformBuildHashCheck map[string]interface{}
)

//func RetrieveCurrentBuildHashCheckIgnored() (platformsToIgnore []string) {
//	if ignoreBuildHashCheck == nil {
//		buildHashCheckMap()
//	}
//
//	// Convert map values to []string
//	for k, _ := range ignoreBuildHashCheck {
//		platformsToIgnore = append(platformsToIgnore, k)
//	}
//
//	return
//}

//func buildHashCheckMap() {
//	ignoreBuildHashCheck = map[string]interface{}{}
//	for _, pla := range strings.Split(os.Getenv("IGNORE_HASH_CHECK"), ",") {
//		ignoreBuildHashCheck[pla] = true
//	}
//}

//func SetBuildHashCheckIgnored(platformsToIgnore []string) (er error) {
//	if er = os.Setenv("IGNORE_HASH_CHECK", strings.Join(platformsToIgnore, ",")); er == nil {
//		buildHashCheckMap()
//	}
//
//	return
//}

//func IsBuildHashCheckIgnored(platform string) bool {
//	if ignoreBuildHashCheck == nil {
//		buildHashCheckMap()
//	}
//
//	_, ok := ignoreBuildHashCheck[platform]
//	return ok
//}

func buildBuildHashCheckMap() {
	platformBuildHashCheck = map[string]interface{}{}

	st := os.Getenv("ENABLE_BUILDHASH_CHECK")
	if len(st) <= 0 {
		log.Printf("platform BuildHash Check:%v", platformBuildHashCheck)
		return
	}
	for _, pla := range strings.Split(os.Getenv("ENABLE_BUILDHASH_CHECK"), ",") {
		platformBuildHashCheck[pla] = true
	}

	log.Printf("platform BuildHash Check:%v", platformBuildHashCheck)
}

func IsPlatformBuildHashToCheck(platform string) bool {
	if platformBuildHashCheck == nil {
		buildBuildHashCheckMap()
	}

	_, ok := platformBuildHashCheck[platform]
	return ok
}

func RetrieveCurrentBuildHashCheck() (platformsToCheckBuildHash []string) {
	if platformBuildHashCheck == nil {
		buildBuildHashCheckMap()
	}

	// Convert map values to []string
	for k, _ := range platformBuildHashCheck {
		platformsToCheckBuildHash = append(platformsToCheckBuildHash, k)
	}

	return
}
