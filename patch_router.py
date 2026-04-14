import re

with open("internal/handler/router.go", "r", encoding="utf-8") as f:
    content = f.read()

if "NotifHandler" not in content:
    content = content.replace("KYCHandler    *KYCHandler", "KYCHandler    *KYCHandler\n\tNotifHandler  *NotificationHandler")

routes = """		chat.GET("/history/:match_id", deps.ChatHandler.GetHistory)
	}

	notifs := v1.Group("/notifications", deps.AuthMw)
	{
		notifs.GET("", deps.NotifHandler.List)
		notifs.GET("/unread-count", deps.NotifHandler.GetUnreadCount)
		notifs.PATCH("/:id/read", deps.NotifHandler.MarkAsRead)
		notifs.POST("/read-all", deps.NotifHandler.MarkAllAsRead)
	}

	interactions := v1.Group("/interactions", deps.AuthMw)"""

content = content.replace('chat.GET("/history/:match_id", deps.ChatHandler.GetHistory)\n\t}\n\n\tinteractions := v1.Group("/interactions", deps.AuthMw)', routes)

with open("internal/handler/router.go", "w", encoding="utf-8") as f:
    f.write(content)
