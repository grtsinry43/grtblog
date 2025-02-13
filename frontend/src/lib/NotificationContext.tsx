"use client"

import type React from "react"
import { createContext, useContext, useState, useCallback } from "react"
import Notification, { type NotificationType } from "@/components/Notification"

interface NotificationContextType {
    showNotification: (message: string, type: NotificationType) => void
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined)

export const useNotification = () => {
    const context = useContext(NotificationContext)
    if (!context) {
        throw new Error("useNotification must be used within a NotificationProvider")
    }
    return context
}

interface NotificationState {
    message: string
    type: NotificationType
}

export const NotificationProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [notification, setNotification] = useState<NotificationState | null>(null)

    const showNotification = useCallback((message: string, type: NotificationType) => {
        setNotification({ message, type })
    }, [])

    const hideNotification = useCallback(() => {
        setNotification(null)
    }, [])

    return (
        <NotificationContext.Provider value={{ showNotification }}>
            {children}
            {notification && (
                <Notification message={notification.message} type={notification.type} onClose={hideNotification} />
            )}
        </NotificationContext.Provider>
    )
}

