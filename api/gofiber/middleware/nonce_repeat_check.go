package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	mySTO "gofiber/cache"
	myCFG "gofiber/config"
	myMDL "gofiber/model"
	"hash"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	skipNonceHeaderKey = "X-Skip-Nonce"
	nonceHeaderKey     = "X-Nonce"
	nonceSignatureKey  = "X-Signature"

	nonceCacheDBIndex  = 2
	nonceCacheDuration = 2 * time.Hour
)

func formHMACVerifyKey(sb myMDL.SessionBase) hash.Hash {
	hmacTemp := hmac.New(sha256.New, []byte(myCFG.RetrieveNonceSecret()))
	hmacTemp.Write([]byte(sb.ComposedID()))

	// log.Printf("formHMACVerifyKey secret: %v, playerID: %v", myCFG.RetrieveNonceSecret(), sb.ComposedID())

	return hmac.New(sha256.New, hmacTemp.Sum(nil))
}

func verifyHMAC(c *fiber.Ctx, sb myMDL.SessionBase, nonce string, signature string) bool {
	// Create the hash with the secret (hmac(sharedString).sum(playerUUID))
	h := formHMACVerifyKey(sb)

	// Write components to the hash in the exact same order as the client
	h.Write([]byte(nonce))
	h.Write(c.Body()) // This is the raw request content
	//h.Write([]byte(sb.ComposedID())) // The player UUID, deprecated

	// log.Printf("NonceRepeatCheck nonce:%v, body:%v", nonce, string(c.Body()))

	// Calculate expected signature
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// log.Printf("expectedSignature: %v, signature: %v", expectedSignature, signature)

	// Constant-time comparison
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func formCacheNonceKey(sb myMDL.SessionBase, nonce string) string {
	return fmt.Sprintf("nonce:%s:%s", sb.ComposedIDSafe(), nonce)
}

func NonceRepeatCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. BYPASS CHECK (Fastest)
		skipPhrase := c.Get(skipNonceHeaderKey)
		if len(skipPhrase) > 0 && skipPhrase == myCFG.RetrieveNonceSkipPhrase() {
			return c.Next()
		}

		// 2. PRESENCE CHECK
		nonce := c.Get(nonceHeaderKey)
		signature := c.Get(nonceSignatureKey)
		if nonce == "" || signature == "" {
			resp := myMDL.Default400Resp.Copy()
			resp.SetErrorAndMessage("missing security headers")
			return c.Status(resp.Status).JSON(resp)
		}

		// 3. HMAC AUTHENTICATION (CPU-only, no DB)
		ses := RetrieveSession(c)
		if !verifyHMAC(c, ses.SessionBase, nonce, signature) {
			resp := myMDL.Default401Resp.Copy()
			resp.SetErrorAndMessage("invalid signature")
			return c.Status(resp.Status).JSON(resp)
		}

		// 4. REPLAY CHECK (Redis - I/O)
		redisKey := formCacheNonceKey(ses.SessionBase, nonce)
		// Store just the nonce key with a simple value "1"
		isNew, err := mySTO.CreateIfNotExistWithDBIndex(nonceCacheDBIndex, redisKey, []byte("1"), nonceCacheDuration)
		if err != nil {
			resp := myMDL.Default500Resp.Copy()
			resp.SetErrorAndMessage("security service error")
			return c.Status(resp.Status).JSON(resp)
		}
		if !isNew {
			resp := myMDL.Default403Resp.Copy()
			resp.SetErrorAndMessage("duplicate request (replay)")
			return c.Status(resp.Status).JSON(resp)
		}

		return c.Next()
	}
}
