package server

import (
	"encoding/json"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 设备 handler 专注设备注册与挑战验证，避免和普通会话接口互相混杂。

func (a *App) handleDeviceRegister(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		Name                 string `json:"name"`
		OS                   string `json:"os"`
		ClientVersion        string `json:"clientVersion"`
		PublicKey            string `json:"publicKey"`
		PublicKeyFingerprint string `json:"publicKeyFingerprint"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}

	device, err := a.authService.RegisterDevice(request.Context(), auth.RegisterDeviceInput{
		AccessToken:          token,
		Name:                 payload.Name,
		OS:                   payload.OS,
		ClientVersion:        payload.ClientVersion,
		PublicKey:            payload.PublicKey,
		PublicKeyFingerprint: payload.PublicKeyFingerprint,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusCreated, requestID, timestamp, map[string]any{
		"deviceId": device.ID,
		"status":   device.Status,
	})
}

func (a *App) handleDeviceChallenge(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		DeviceID string `json:"deviceId"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}

	challenge, err := a.authService.CreateDeviceChallenge(request.Context(), auth.CreateDeviceChallengeInput{
		AccessToken: token,
		DeviceID:    payload.DeviceID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"challengeId": challenge.ID,
		"challenge":   challenge.Challenge,
		"expiresIn":   challenge.ExpiresIn,
	})
}

func (a *App) handleDeviceChallengeVerify(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		ChallengeID string `json:"challengeId"`
		Signature   string `json:"signature"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}

	result, err := a.authService.VerifyDeviceChallenge(request.Context(), auth.VerifyDeviceChallengeInput{
		AccessToken: token,
		ChallengeID: payload.ChallengeID,
		Signature:   payload.Signature,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"verified": result.Verified,
	})
}
