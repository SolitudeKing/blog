import { defineStore } from 'pinia'
import { getLobbySetting } from '@/api/modules/setting'
import type { LobbySetting } from '@/types/setting'

interface SettingState {
  lobby: LobbySetting | null
}

export const useSettingStore = defineStore('setting', {
  state: (): SettingState => ({
    lobby: null,
  }),
  actions: {
    async loadLobby() {
      this.lobby = await getLobbySetting()
      document.documentElement.dataset.theme = this.lobby.theme
      document.documentElement.dataset.mode = this.lobby.mode
    },
  },
})

