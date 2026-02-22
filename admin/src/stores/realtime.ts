import { acceptHMRUpdate, defineStore } from 'pinia'
import { ref } from 'vue'

export const useRealtimeStore = defineStore('realtimeStore', () => {
  const notificationWsConnected = ref(false)
  const realtimeWsConnected = ref(false)

  function setNotificationWsConnected(connected: boolean) {
    notificationWsConnected.value = connected
  }

  function setRealtimeWsConnected(connected: boolean) {
    realtimeWsConnected.value = connected
  }

  return {
    notificationWsConnected,
    realtimeWsConnected,
    setNotificationWsConnected,
    setRealtimeWsConnected,
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useRealtimeStore, import.meta.hot))
}
