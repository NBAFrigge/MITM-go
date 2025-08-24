<template>
  <div id="app">
    <AppHeader/>

    <!-- Navbar -->
    <AppNavbar
        :tabs="tabs"
        :active-tab="activeTab"
        @tab-changed="activeTab = $event"
    />

    <main class="app-main">
      <!-- Proxy Control Panel -->
      <div v-if="activeTab === 'proxy'" class="tab-content">
        <ProxyControls
            :is-running="isRunning"
            :status-text="statusText"
            @start-proxy="startProxy"
            @stop-proxy="stopProxy"
            @clear-sessions="clearSessions"
        />

        <SearchPanel
            @search-results="handleSearchResults"
            @search-cleared="handleSearchCleared"
        />

        <div
            ref="contentArea"
            class="content-area"
            :class="{'full-width': !selectedSession, 'dragging': isDragging}"
        >
          <SessionsList
              :sessions="displayedSessions"
              :selected-session-id="selectedSessionId"
              @select-session="selectSession"
              @send-to-compare="handleSendToCompare"
              :style="{ width: selectedSession ? `${splitterPos}%` : '100%' }"
          />

          <div
              v-if="selectedSession"
              class="splitter"
              @mousedown="startDrag"
              :class="{ 'dragging': isDragging }"
          >
            <div class="splitter-handle"></div>
          </div>

          <SessionDetails
              v-if="selectedSession"
              :session="selectedSession"
              @close="deselectSession"
              @session-updated="updateSelectedSession"
              :style="{ width: `${100 - splitterPos}%` }"
          />
        </div>
      </div>

      <!-- Compare Tab -->
      <div v-if="activeTab === 'compare'" class="tab-content">
        <CompareSessions
            :sessions="sessions"
            :selected-sessions="compareSessions"
        />
      </div>

      <!-- Options Panel -->
      <div v-if="activeTab === 'options'" class="tab-content">
        <OptionsPanel
            :proxy-port="proxyPort"
            :verbose="verbose"
            :is-running="isRunning"
            @update-port="proxyPort = $event"
            @update-verbose="verbose = $event"
        />
      </div>
    </main>

    <!-- Status Bar -->
    <div class="status-bar">
      <div class="status-left">
        <span class="status-indicator" :class="{ 'running': isRunning, 'stopped': !isRunning }">
          {{ isRunning ? '●' : '○' }}
        </span>
        <span class="status-text">{{ statusText }}</span>
      </div>
      <div class="status-right">
        <span class="session-count">{{ sessions.length }} session{{ sessions.length !== 1 ? 's' : '' }}</span>
        <span class="current-tab">{{ activeTab.toUpperCase() }}</span>
      </div>
    </div>
  </div>
</template>

<script>
import AppHeader from './components/AppHeader.vue'
import AppNavbar from './components/AppNavbar.vue'
import ProxyControls from './components/ProxyControls.vue'
import SessionsList from './components/SessionsList.vue'
import SessionDetails from './components/SessionDetails.vue'
import SearchPanel from './components/SearchPanel.vue'
import OptionsPanel from './components/OptionsPanel.vue'
import CompareSessions from './components/CompareSessions.vue'
import {
  StartProxy,
  StopProxy,
  GetProxyStatus,
  GetSessions,
  GetSessionDetails,
  ClearSessions
} from '../wailsjs/go/main/App'

export default {
  name: 'App',
  components: {
    AppHeader,
    AppNavbar,
    ProxyControls,
    SessionsList,
    SessionDetails,
    SearchPanel,
    OptionsPanel,
    CompareSessions
  },
  data() {
    return {
      activeTab: 'proxy',
      tabs: [
        {
          id: 'proxy',
          label: 'Proxy',
          icon: '<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="m12 1 0 6m0 6 0 6m11-7-6 0m-6 0-6 0"/></svg>'
        },
        {
          id: 'compare',
          label: 'Compare',
          icon: '<svg viewBox="0 0 24 24"><path d="M9 11H5a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2v-6a2 2 0 0 0-2-2z"/><path d="M19 3H15a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2z"/><path d="M7 11V7a2 2 0 0 1 2-2h6"/><path d="M17 13v4a2 2 0 0 1-2 2H9"/></svg>'
        },
        {
          id: 'options',
          label: 'Options',
          icon: '<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1 1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>'
        }
      ],
      proxyPort: 8080,
      verbose: false,
      isRunning: false,
      sessions: [],
      selectedSessionId: null,
      selectedSession: null,
      statusText: 'Proxy stopped',
      pollingInterval: null,
      isDragging: false,
      splitterPos: 35,
      searchResults: null,
      isSearchActive: false,
      compareSessions: []
    }
  },

  mounted() {
    this.updateStatus()
    document.addEventListener('mousemove', this.onDrag)
    document.addEventListener('mouseup', this.stopDrag)
  },

  beforeUnmount() {
    this.stopPolling()
    document.removeEventListener('mousemove', this.onDrag)
    document.removeEventListener('mouseup', this.stopDrag)
  },

  methods: {
    async startProxy() {
      try {
        await StartProxy(this.proxyPort, this.verbose)
        await this.updateStatus()
        this.startPolling()
      } catch (error) {
        console.error('Error starting proxy:', error)
        this.$showNotification('Error starting proxy: ' + error, 'error')
      }
    },

    async stopProxy() {
      try {
        await StopProxy()
        await this.updateStatus()
        this.stopPolling()
      } catch (error) {
        console.error('Error stopping proxy:', error)
        this.$showNotification('Error stopping proxy: ' + error, 'error')
      }
    },

    async updateStatus() {
      try {
        const status = await GetProxyStatus()
        this.isRunning = status.running
        this.statusText = status.running
            ? `Running on port ${status.port}`
            : 'Stopped'
      } catch (error) {
        console.error('Error getting status:', error)
        this.statusText = 'Status unknown'
      }
    },

    async refreshSessions() {
      try {
        this.sessions = await GetSessions()
        // console.log('Sessions refreshed:', this.sessions.length)
      } catch (error) {
        console.error('Error getting sessions:', error)
      }
    },

    async selectSession(sessionId) {
      try {
        this.selectedSessionId = sessionId
        this.selectedSession = await GetSessionDetails(sessionId)
      } catch (error) {
        console.error('Error getting session details:', error)
        this.selectedSession = null
      }
    },

    updateSelectedSession(updatedSession) {
      if (!updatedSession?.id) {
        console.warn('Invalid updated session data')
        return
      }

      this.selectedSession = updatedSession

      const index = this.sessions.findIndex(s => s.id === updatedSession.id)
      if (index !== -1) {
        this.sessions[index] = {
          ...this.sessions[index],
          webSocketData: updatedSession.webSocketData,
          duration: updatedSession.duration,
          status: updatedSession.status
        }
      }
    },

    deselectSession() {
      this.selectedSessionId = null
      this.selectedSession = null
    },

    async clearSessions() {
      try {
        await ClearSessions()
        this.sessions = []
        this.selectedSession = null
        this.selectedSessionId = null
        this.$showNotification('Sessions cleared', 'success')
      } catch (error) {
        console.error('Error clearing sessions:', error)
        this.$showNotification('Error clearing sessions', 'error')
      }
    },

    startPolling() {
      this.pollingInterval = setInterval(() => {
        this.refreshSessions()
      }, 1000)
    },

    stopPolling() {
      if (this.pollingInterval) {
        clearInterval(this.pollingInterval)
        this.pollingInterval = null
      }
    },

    startDrag(e) {
      e.preventDefault()
      this.isDragging = true
      document.body.style.cursor = 'ew-resize'
      document.body.style.userSelect = 'none'
    },

    onDrag(e) {
      if (!this.isDragging || !this.selectedSession || !this.$refs.contentArea) return
      e.preventDefault()

      const contentArea = this.$refs.contentArea
      const rect = contentArea.getBoundingClientRect()
      const x = e.clientX - rect.left
      const newPos = (x / rect.width) * 100

      if (newPos >= 20 && newPos <= 80) {
        this.splitterPos = newPos
      }
    },

    stopDrag() {
      if (this.isDragging) {
        this.isDragging = false
        document.body.style.cursor = ''
        document.body.style.userSelect = ''
      }
    },

    handleSearchResults(results) {
      if (!results || results.length === 0) {
        this.searchResults = null
        this.isSearchActive = false
        return
      }
      this.searchResults = results
      this.isSearchActive = true
    },

    handleSearchCleared() {
      this.searchResults = null
      this.isSearchActive = false
    },

    handleSendToCompare(sessionId) {
      if (!this.compareSessions.includes(sessionId)) {
        this.compareSessions.push(sessionId)
      }
      this.activeTab = 'compare'
    },

    $showNotification(message, type = 'info') {
      console.log(`[${type.toUpperCase()}] ${message}`)
    }
  },

  computed: {
    displayedSessions() {
      return this.isSearchActive ? this.searchResults : this.sessions
    }
  }
}
</script>

<style scoped>
#app {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color-dark);
  color: var(--text-color-primary);
  font-family: var(--font-family-ui);
  overflow: hidden;
}

.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.tab-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
  animation: fadeIn var(--transition-fast);
}

.content-area {
  display: flex;
  flex: 1;
  overflow: hidden;
  gap: 0;
  padding: 0 var(--spacing-md) var(--spacing-md) var(--spacing-md);
  min-height: 0;
  position: relative;
}

.content-area.full-width {
  gap: 0;
}

.content-area.dragging {
  user-select: none;
  cursor: ew-resize;
}

.content-area.dragging * {
  pointer-events: none;
}

.splitter {
  width: 8px;
  background-color: var(--border-color);
  cursor: ew-resize;
  flex-shrink: 0;
  border-radius: var(--radius-sm);
  position: relative;
  z-index: 10;
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 2px;
}

.splitter:hover {
  background-color: var(--accent-color-blue);
  box-shadow: 0 0 0 1px var(--accent-color-blue);
}

.splitter.dragging {
  background-color: var(--accent-color-blue);
}

.splitter-handle {
  width: 2px;
  height: 24px;
  background-color: rgba(169, 183, 198, 0.5);
  border-radius: 1px;
  transition: var(--transition-fast);
  pointer-events: none;
}

.splitter:hover .splitter-handle,
.splitter.dragging .splitter-handle {
  background-color: white;
  height: 32px;
}

.status-bar {
  height: 24px;
  background-color: var(--bg-color-medium);
  border-top: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-md);
  font-size: var(--font-size-small);
  color: var(--text-color-secondary);
  font-family: var(--font-family);
  flex-shrink: 0;
}

.status-left,
.status-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.status-indicator {
  font-size: 8px;
  transition: var(--transition-fast);
}

.status-indicator.running {
  color: var(--status-success);
  animation: pulse 2s infinite;
}

.status-indicator.stopped {
  color: var(--text-color-secondary);
}

.status-text {
  font-weight: 500;
}

.session-count {
  color: var(--text-color-secondary);
}

.current-tab {
  color: var(--accent-color-blue);
  font-weight: 600;
  font-size: 10px;
  letter-spacing: 0.5px;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1024px) {
  .content-area {
    gap: 0;
    padding: 0 var(--spacing-sm) var(--spacing-sm) var(--spacing-sm);
  }

  .splitter {
    width: 6px;
    margin: 0 1px;
  }
}

@media (max-width: 768px) {
  .content-area {
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .content-area > div {
    width: 100% !important;
  }

  .splitter {
    display: none;
  }

  .status-bar {
    padding: 0 var(--spacing-sm);
  }

  .status-left,
  .status-right {
    gap: var(--spacing-sm);
  }

  .current-tab {
    display: none;
  }
}

@media (max-width: 480px) {
  .status-bar {
    height: 20px;
    font-size: 10px;
  }

  .session-count {
    display: none;
  }
}

@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
  .splitter-handle {
    width: 1px;
  }
}

@media (prefers-color-scheme: dark) {
  .splitter:hover {
    box-shadow: 0 0 0 1px var(--accent-color-blue), 0 0 4px rgba(104, 151, 187, 0.3);
  }
}
</style>