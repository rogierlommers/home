package message_webhook

import (
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/tinycache"
)

// curl -H 'Content-Type: application/json' -d '{"data": {"actor": "kSUTyS4CNTRL","deviceName": "github-fv-az525-182.hake-cod.ts.net","managedBy": "tag:ci","nodeID": "nkwtPXZacW11CNTRL","url": "https://login.tailscale.com/admin/machines/100.80.115.40"},"message": "Node github-fv-az525-182.hake-cod.ts.net approved","tailnet": "rogierlommers.github","timestamp": "2024-10-12T17:58:02.240764417Z","type": "nodeApproved","version": 1}' -X POST http://localhost:3000/api/message_webhook

// EXAMPLE INCOMING REQUEST FROM TAILSCALE
//
// {
//     "data": {
//         "actor": "kSUTyS4CNTRL",
//         "deviceName": "github-fv-az525-182.hake-cod.ts.net",
//         "managedBy": "tag:ci",
//         "nodeID": "nkwtPXZacW11CNTRL",
//         "url": "https://login.tailscale.com/admin/machines/100.80.115.40"
//     },
//     "message": "Node github-fv-az525-182.hake-cod.ts.net approved",
//     "tailnet": "rogierlommers.github",
//     "timestamp": "2024-10-12T17:58:02.240764417Z",
//     "type": "nodeApproved",
//     "version": 1
// }

var cache *tinycache.Cache

func Add(router *gin.Engine, cfg config.AppConfig) {
	router.POST("/api/message_webhook", addMessage)
	router.GET("/api/message_webhook/rss", displayRSS)
	cache = tinycache.NewCache(100)
}
