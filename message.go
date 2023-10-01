package test_task

import (
	"encoding/json"
	"math/rand"
)

type ActionType int

const (
	ActionSession = iota
	ActionHit
	ActionAccess
	ActionAdView
)
const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	BidEndpoint   = "/bid"
	ProxyEndpoint = "/proxy"
)

// RequestV1 represents version 1 of external request.
type RequestV1 struct {
	ActionType ActionType `json:"action_type,omitempty"`
	Id         string     `json:"id,omitempty"`
	Balance    float64    `json:"balance,omitempty"`
}

// RequestV2 represents version 2 of external request.
type RequestV2 struct {
	ActionType ActionType `json:"action_type,omitempty"`
	Key        int32      `json:"key,omitempty"`
	Price      float32    `json:"price,omitempty"`
	Note       string     `json:"note,omitempty"`
}

// RequestV3 represents version 3 of external request.
type RequestV3 struct {
	ActionType ActionType `json:"action_type,omitempty"`
	Name       string     `json:"name,omitempty"`
	Bid        uint64     `json:"bid,omitempty"`
}

// InnerRequest represents request to interchange data between proxy and recipients.
type InnerRequest struct {
	ActionType ActionType `json:"action_type,omitempty"`
	Id         string     `json:"id,omitempty"`
	MinPrice   float64    `json:"min_price,omitempty"`
	Comment    string     `json:"comment,omitempty"`
}

// InnerResponse represents response of recipients.
type InnerResponse struct {
	RecipId int32   `json:"recip_id,omitempty"`
	Id      string  `json:"id,omitempty"`
	Message string  `json:"message,omitempty"`
	Bid     float64 `json:"bid,omitempty"`
}

// CommonProxyRequest represents request that has fields of all external request versions
type CommonProxyRequest struct {
	ActionType ActionType `json:"action_type,omitempty"`
	Id         *string    `json:"id,omitempty"`
	Balance    *float64   `json:"balance,omitempty"`
	Key        *int32     `json:"key,omitempty"`
	Price      *float32   `json:"price,omitempty"`
	Note       *string    `json:"note,omitempty"`
	Name       *string    `json:"name,omitempty"`
	Bid        *uint64    `json:"bid,omitempty"`
}

func GenerateRandomRequest(typ int) []byte {
	switch typ {
	case 0:
		req := RequestV1{
			ActionType: getRandomActionType(),
			Id:         GetRandomStr(20),
			Balance:    rand.Float64(),
		}
		b, _ := json.Marshal(req)
		return b
	case 1:
		req := RequestV2{
			ActionType: getRandomActionType(),
			Key:        rand.Int31n(9999),
			Price:      rand.Float32(),
			Note:       GetRandomStr(50),
		}
		b, _ := json.Marshal(req)
		return b
	case 2:
		req := RequestV3{
			ActionType: getRandomActionType(),
			Name:       GetRandomStr(30),
			Bid:        rand.Uint64(),
		}
		b, _ := json.Marshal(req)
		return b
	}
	return nil
}

func getRandomActionType() ActionType {
	switch rand.Intn(4) {
	case 0:
		return ActionSession
	case 1:
		return ActionHit
	case 2:
		return ActionAccess
	case 3:
		return ActionAdView
	}
	return ActionSession
}

func GetRandomBytes(l int) []byte {
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return b
}

func GetRandomStr(l int) string {
	return string(GetRandomBytes(l))
}

func GetRandomFloat(min float64) float64 {
	return min + rand.Float64()
}
