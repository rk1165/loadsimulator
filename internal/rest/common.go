package rest

import (
	"fmt"
	"log"

	"github.com/rk1165/loadsimulator/internal"
	"github.com/rk1165/loadsimulator/internal/types"
)

func resolveHeaders(apiConfig types.ApiConfig) map[string]string {
	oscar := &internal.OAuthBearer{
		ClientId:     apiConfig.ClientId,
		ClientSecret: apiConfig.ClientSecret,
		ClientScope:  apiConfig.Scope,
		TokenUrl:     internal.TokenUrl,
	}
	token, err := oscar.GenerateToken()
	if err != nil {
		log.Fatal(err)
	}
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	headers["Content-Type"] = apiConfig.ContentType
	return headers
}
