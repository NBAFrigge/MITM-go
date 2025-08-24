<template>
  <div class="details-panel">
    <div class="details-header">
      <div class="header-content">
        <h2>Session Details</h2>
        <div v-if="isWebSocketSession && isActiveWebSocket" class="auto-refresh-indicator">
          <div class="refresh-dot" :class="{ 'active': isAutoRefreshing }"></div>
          <span>{{ isAutoRefreshing ? 'Live' : 'WebSocket' }}</span>
        </div>
      </div>
      <button @click="$emit('close')" class="btn-close" aria-label="Close details">
        ×
      </button>
    </div>

    <div class="tab-navigation">
      <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="['tab-button', { 'active': activeTab === tab.id }]"
      >
        {{ tab.label }}
      </button>
    </div>

    <div class="session-details scrollable">
      <div v-show="activeTab === 'request'" class="tab-content">
        <RequestDetails :session="session" />
      </div>

      <div v-show="activeTab === 'response'" class="tab-content">
        <ResponseDetails :session="session" />
      </div>

      <div v-if="isWebSocketSession" v-show="activeTab === 'websocket'" class="tab-content">
        <WebSocketDetails :web-socket-data="session.webSocketData" />
      </div>

      <div v-if="isWebSocketSession" v-show="activeTab === 'messages'" class="tab-content">
        <WebSocketMessages :messages="webSocketMessages" />
      </div>

      <div v-show="activeTab === 'tls'" class="tab-content">
        <TLSInfo :session="session" />
      </div>
    </div>
  </div>
</template>

<script>
import RequestDetails from './RequestDetails.vue'
import ResponseDetails from './ResponseDetails.vue'
import WebSocketDetails from './WebSocketDetails.vue'
import WebSocketMessages from './WebSocketMessages.vue'
import TLSInfo from './TLSInfo.vue'

export default {
  name: 'SessionDetails',
  components: {
    RequestDetails,
    ResponseDetails,
    WebSocketDetails,
    WebSocketMessages,
    TLSInfo
  },
  props: {
    session: {
      type: Object,
      required: true
    }
  },
  emits: ['close', 'session-updated'],
  data() {
    return {
      activeTab: 'request',
      refreshInterval: null,
      isAutoRefreshing: false
    };
  },
  computed: {
    isWebSocketSession() {
      const isWS = this.session?.type === 'WebSocketSession';
      console.log(`Session ${this.session?.id}: type="${this.session?.type}", isWS=${isWS}`);
      return isWS;
    },
    isActiveWebSocket() {
      try {
        if (!this.isWebSocketSession || !this.session?.webSocketData) {
          console.log(`Session ${this.session?.id}: Not active WebSocket - isWS=${this.isWebSocketSession}, hasData=${!!this.session?.webSocketData}`);
          return false;
        }
        const state = this.session.webSocketData.state;
        const isActive = state === 1 || state === 0 ||
            state === 'connected' || state === 'connecting';
        console.log(`Session ${this.session?.id}: WebSocket state="${state}", isActive=${isActive}`);
        return isActive;
      } catch (error) {
        console.error('Error checking WebSocket state:', error);
        return false;
      }
    },
    tabs() {
      if (this.isWebSocketSession) {
        return [
          { id: 'websocket', label: 'WebSocket Details' },
          { id: 'messages', label: 'Messages' },
          { id: 'tls', label: 'TLS Info' }
        ];
      } else {
        return [
          { id: 'request', label: 'Request' },
          { id: 'response', label: 'Response' },
          { id: 'tls', label: 'TLS Info' }
        ];
      }
    },
    webSocketMessages() {
      try {
        if (!this.session?.webSocketData?.messages) {
          return [];
        }

        if (Array.isArray(this.session.webSocketData.messages)) {
          return this.session.webSocketData.messages;
        }

        if (this.session.webSocketData.messages.messages) {
          return this.session.webSocketData.messages.messages;
        }

        return [];
      } catch (error) {
        console.error('Error getting WebSocket messages:', error);
        return [];
      }
    }
  },
  watch: {
    session: {
      handler(newSession, oldSession) {
        if (newSession?.id !== oldSession?.id) {
          if (newSession?.type === 'WebSocketSession') {
            this.activeTab = 'websocket';
          } else {
            this.activeTab = 'request';
          }
          this.setupAutoRefresh();
        } else if (newSession?.id === oldSession?.id) {
          this.setupAutoRefresh();
        }
      },
      immediate: true
    }
  },

  mounted() {
    this.setupAutoRefresh();
  },

  beforeUnmount() {
    this.stopAutoRefresh();
  },

  methods: {
    setupAutoRefresh() {
      if (!this.session?.id) {
        return;
      }

      this.stopAutoRefresh();

      if (this.isActiveWebSocket) {
        this.startAutoRefresh();
      }
    },

    startAutoRefresh() {
      if (!this.session?.id) {
        console.warn('Cannot start auto-refresh: no session ID');
        return;
      }

      if (!window.go?.main?.App?.GetSessionDetails) {
        console.warn('Cannot start auto-refresh: Wails API not available');
        return;
      }

      this.isAutoRefreshing = true;

      this.refreshInterval = setInterval(async () => {
        try {
          if (!this.session?.id || !window.go?.main?.App?.GetSessionDetails) {
            this.stopAutoRefresh();
            return;
          }

          const updatedSession = await window.go.main.App.GetSessionDetails(this.session.id);

          if (updatedSession) {
            this.$emit('session-updated', updatedSession);

            if (updatedSession.webSocketData?.state !== 1 &&
                updatedSession.webSocketData?.state !== 0) {
              this.stopAutoRefresh();
            }
          } else {
            this.stopAutoRefresh();
          }
        } catch (error) {
          console.error('Error refreshing session details:', error);
          this.stopAutoRefresh();
        }
      }, 2000);

      console.log(`Auto-refresh started for WebSocket session ${this.session.id}`);
    },

    stopAutoRefresh() {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval);
        this.refreshInterval = null;
      }
      this.isAutoRefreshing = false;
    }
  }
}
</script>

<style scoped>
.details-panel {
  width: 50%;
  background-color: var(--bg-color-dark);
  display: flex;
  flex-direction: column;
  border-left: 1px solid var(--border-color);
  overflow: hidden;
  font-family: var(--font-family);
}

.details-header {
  padding: var(--spacing-sm) var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: var(--bg-color-light);
}

.header-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.details-header h2 {
  margin: 0;
  font-size: var(--font-size-small);
  font-weight: 600;
  color: var(--text-color-primary);
}

.auto-refresh-indicator {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: 2px var(--spacing-sm);
  background-color: var(--status-success);
  color: white;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.refresh-dot {
  width: 4px;
  height: 4px;
  background-color: white;
  border-radius: 50%;
}

.refresh-dot.active {
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.5;
    transform: scale(0.8);
  }
}

.btn-close {
  background: none;
  border: none;
  color: var(--text-color-secondary);
  cursor: pointer;
  font-size: 16px;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: var(--transition-fast);
}

.btn-close:hover {
  color: var(--status-error);
  background-color: var(--hover-color);
}

.tab-navigation {
  display: flex;
  background-color: var(--bg-color-light);
  border-bottom: 1px solid var(--border-color);
}

.tab-button {
  flex: 1;
  padding: var(--spacing-sm) var(--spacing-md);
  border: none;
  background: none;
  color: var(--text-color-secondary);
  cursor: pointer;
  font-size: var(--font-size-small);
  font-weight: 500;
  border-bottom: 2px solid transparent;
  transition: var(--transition-fast);
  font-family: inherit;
}

.tab-button:hover {
  color: var(--text-color-primary);
  background-color: var(--hover-color);
}

.tab-button.active {
  color: var(--text-color-primary);
  border-bottom-color: var(--accent-color-blue);
  background-color: var(--bg-color-dark);
}

.session-details {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: var(--spacing-md);
  overflow-y: auto;
  gap: var(--spacing-md);
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.scrollable {
  overflow-y: scroll;
}

@media (max-width: 768px) {
  .details-panel {
    width: 100%;
    border-left: none;
    border-top: 1px solid #3c3c3c;
  }

  .tab-navigation {
    flex-wrap: wrap;
  }

  .tab-button {
    min-width: 0;
    flex-basis: auto;
    font-size: 10px;
    padding: 4px 8px;
  }
}

</style>