import {useNotification} from "@/lib/NotificationContext"
import type {NotificationType} from "@/components/Notification"

export const useNotificationUtil = () => {
    const { showNotification } = useNotification()

    return (message: string, type: NotificationType = "info") => {
        showNotification(message, type)
    }
}

