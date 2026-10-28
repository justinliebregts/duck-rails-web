package model

type PhotonResultCode int

var (
	PhotonAuthenticationSuccessful PhotonResultCode = 1
	PhotonAuthenticationFailed     PhotonResultCode = 2
	PhotonInvalidParameters        PhotonResultCode = 3
)
