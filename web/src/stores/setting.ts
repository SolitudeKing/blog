import { defineStore } from 'pinia'
import { getLobbySetting } from '@/api/modules/setting'
import { useTheme } from '@/composables/useTheme'
import type { LobbySetting } from '@/types/setting'

interface SettingState {
  lobby: LobbySetting | null
}

let lobbyRequest: Promise<LobbySetting> | null = null

export const useSettingStore = defineStore('setting', {
  state: (): SettingState => ({
    lobby: null,
  }),
  actions: {
    applyLobby(lobby: LobbySetting) {
      this.lobby = lobby
      useTheme().syncFromServer({
        theme: lobby.theme,
        mode: lobby.mode,
      })
      return lobby
    },
    async loadLobby(force = false) {
      if (this.lobby && !force) {
        return this.lobby
      }

      lobbyRequest ??= getLobbySetting()
      try {
        const lobby = await lobbyRequest
        return this.applyLobby(lobby)
      } finally {
        lobbyRequest = null
      }
    },
  },
})
