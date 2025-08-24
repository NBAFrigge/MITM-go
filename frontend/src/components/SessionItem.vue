<template>
  <div
      class="session-item"
      :class="{
        'active': isSelected,
        'websocket-session': isWebSocketSession,
        'has-error': hasError,
        'has-body': hasRequestBody
      }"
      @click="handleClick"
      @contextmenu.prevent="showContextMenu"
  >
    <template v-if="isWebSocketSession">
      <div class="session-header">
        <div class="session-method websocket-method">
          WS
        </div>
        <div class="session-indicators">
          <div v-if="messageCount > 0" class="message-count" :title="`${messageCount} messages`">
            {{ messageCount }}
          </div>
        </div>
      </div>

      <div class="session-content">
        <div class="session-url" :title="session.url">{{ session.url }}</div>
        <div class="session-meta">
          <div class="websocket-state" :class="getWebSocketStateClass(webSocketState)">
            {{ getWebSocketStateName(webSocketState) }}
          </div>
          <div class="session-time">
            {{ formatTime(session.timestamp) }}
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="session-header">
        <div class="session-method" :class="'method-' + session.method.toLowerCase()">
          {{ session.method }}
        </div>
        <div class="session-indicators">
          <div v-if="hasError" class="indicator indicator-error" title="Request has error">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2L1 21h22L12 2zm0 3.99L19.53 19H4.47L12 5.99zM11 16h2v2h-2zm0-6h2v4h-2z"/>
            </svg>
          </div>
          <div v-if="hasRequestBody" class="indicator indicator-body" title="Request has body">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
              <path d="M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.89 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zm4 18H6V4h7v5h5v11z"/>
            </svg>
          </div>
          <div v-if="hasCookies" class="indicator indicator-cookies" title="Request has cookies">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
              <circle cx="12" cy="12" r="10"/>
              <circle cx="8" cy="8" r="1.5"/>
              <circle cx="16" cy="10" r="1.5"/>
              <circle cx="10" cy="16" r="1.5"/>
              <circle cx="15" cy="15" r="1.5"/>
            </svg>
          </div>
          <div v-if="responseSize" class="indicator indicator-size" :title="`Response size: ${formatSize(responseSize)}`">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
              <path d="M13 9h5.5L13 3.5V9zM6 2h8l6 6v12c0 1.1-.9 2-2 2H6c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2zm9 16v-2H6v2h9zm3-4v-2H6v2h12z"/>
            </svg>
            <span class="size-text">{{ formatSize(responseSize) }}</span>
          </div>
        </div>
      </div>

      <div class="session-content">
        <div class="session-url" :title="session.url">{{ session.url }}</div>
        <div class="session-meta">
          <div class="session-status" :class="getStatusClass(session.status)">
            {{ session.status }}
          </div>
          <div v-if="responseDuration" class="session-duration" :title="`Response time: ${responseDuration}ms`">
            {{ formatDuration(responseDuration) }}
          </div>
          <div class="session-time">
            {{ formatTime(session.timestamp) }}
          </div>
        </div>
      </div>
    </template>

    <div
        v-if="contextMenuVisible"
        class="context-menu"
        :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
        @click.stop
    >
      <div class="context-menu-item" @click="handleReplay">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" style="margin-right: 4px;">
          <path d="M4 12V6a6 6 0 1 1 6 6H4zm2-2h4a4 4 0 1 0-4-4v4z"/>
        </svg>
        Replay
      </div>
      <div class="context-menu-item" @click="handleCopyCurl">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" style="margin-right: 4px;">
          <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
        </svg>
        Copy as cURL
      </div>
      <div v-if="!isWebSocketSession" class="context-menu-divider"></div>
      <div v-if="!isWebSocketSession" class="context-menu-item" @click="handleSendToCompare">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" style="margin-right: 4px;">
          <path d="M9 11H5a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2v-6a2 2 0 0 0-2-2z"/>
          <path d="M19 3H15a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2z"/>
          <path d="M7 11V7a2 2 0 0 1 2-2h6"/>
          <path d="M17 13v4a2 2 0 0 1-2 2H9"/>
        </svg>
        Send to Compare
      </div>
    </div>
  </div>
</template>

<script>
import { Replay, GetSessionCurl } from '../../wailsjs/go/main/App'

export default {
  name: 'SessionItem',
  props: {
    session: {
      type: Object,
      required: true
    },
    isSelected: {
      type: Boolean,
      default: false
    }
  },
  emits: ['select', 'send-to-compare'],
  data() {
    return {
      contextMenuVisible: false,
      contextMenuX: 0,
      contextMenuY: 0
    }
  },
  computed: {
    isWebSocketSession() {
      return this.session?.type === 'WebSocketSession';
    },
    webSocketState() {
      if (!this.isWebSocketSession || !this.session.webSocketData) {
        return 'Unknown';
      }
      return this.session.webSocketData.state || 'Unknown';
    },
    messageCount() {
      if (!this.isWebSocketSession || !this.session.webSocketData?.messageStats) {
        return 0;
      }
      return this.session.webSocketData.messageStats.totalMessages || 0;
    },
    hasError() {
      return !!this.session.error || (this.session.status >= 400);
    },
    hasRequestBody() {
      return !!this.session.requestBody && this.session.requestBody.trim() !== '';
    },
    hasCookies() {
      const cookies = this.session.requestCookies;
      return cookies && typeof cookies === 'object' && Object.keys(cookies).length > 0;
    },
    responseSize() {
      if (this.session.responseBody) {
        return new Blob([this.session.responseBody]).size;
      }
      return null;
    },
    responseDuration() {
      return this.session.duration || null;
    }
  },
  mounted() {
    document.addEventListener('click', this.hideContextMenu)
  },
  beforeUnmount() {
    document.removeEventListener('click', this.hideContextMenu)
  },
  methods: {
    handleClick() {
      this.$emit('select', this.session.id);
    },
    formatTime(timestamp) {
      return new Date(timestamp).toLocaleTimeString()
    },
    formatSize(bytes) {
      if (bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    },
    formatDuration(ms) {
      if (ms < 1000) return `${ms}ms`;
      return `${(ms / 1000).toFixed(1)}s`;
    },
    getStatusClass(status) {
      if (status >= 200 && status < 300) return 'status-2xx'
      if (status >= 300 && status < 400) return 'status-3xx'
      if (status >= 400 && status < 500) return 'status-4xx'
      if (status >= 500) return 'status-5xx'
      return ''
    },
    getWebSocketStateClass(state) {
      switch (state) {
        case 1: return 'ws-state-connected';
        case 3: return 'ws-state-disconnected';
        case 0: return 'ws-state-connecting';
        case 4: return 'ws-state-error';
        default: return 'ws-state-unknown';
      }
    },
    getWebSocketStateName(state) {
      switch (state) {
        case 0: return 'Connecting';
        case 1: return 'Connected';
        case 3: return 'Disconnected';
        case 4: return 'Error';
        default: return 'Unknown';
      }
    },
    showContextMenu(event) {
      this.contextMenuX = event.clientX
      this.contextMenuY = event.clientY
      this.contextMenuVisible = true
    },
    hideContextMenu() {
      this.contextMenuVisible = false
    },
    async handleReplay() {
      try {
        await Replay(this.session.id)
      } catch (error) {
        console.error('Replay failed:', error)
      }
      this.hideContextMenu()
    },
    async handleCopyCurl() {
      try {
        const curlCommand = await GetSessionCurl(this.session.id)
        await navigator.clipboard.writeText(curlCommand)
      } catch (error) {
        console.error('Copy cURL failed:', error)
      }
      this.hideContextMenu()
    },
    handleSendToCompare() {
      this.$emit('send-to-compare', this.session.id);
      this.hideContextMenu();
    }
  }
}
</script>

<style scoped>
.session-item {
  padding: var(--spacing-sm) var(--spacing-sm);
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: var(--transition-fast);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  color: var(--text-color-primary);
  position: relative;
  font-family: var(--font-family);
  font-size: var(--font-size-small);
}

.session-item:hover {
  background-color: var(--hover-color);
}

.session-item.active {
  background-color: var(--selection-color);
  color: white;
}

.session-item.has-error {
  border-left: 2px solid var(--status-error);
}

.websocket-session {
  border-left: 2px solid var(--field-color);
}

.session-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-xs);
}

.session-method {
  font-weight: bold;
  padding: 1px var(--spacing-xs);
  border-radius: var(--radius-sm);
  font-size: 10px;
  text-transform: uppercase;
  flex-shrink: 0;
}

.method-get { background-color: var(--method-get); color: white; }
.method-post { background-color: var(--method-post); color: white; }
.method-put { background-color: var(--method-put); color: white; }
.method-delete { background-color: var(--method-delete); color: white; }
.method-patch { background-color: var(--method-patch); color: white; }
.method-head { background-color: var(--text-color-secondary); color: white; }
.method-options { background-color: var(--status-info); color: white; }

.websocket-method {
  background-color: var(--field-color);
  color: white;
}

.session-indicators {
  display: flex;
  gap: 2px;
  align-items: center;
}

.indicator {
  display: flex;
  align-items: center;
  gap: 1px;
  padding: 1px var(--spacing-xs);
  border-radius: 2px;
  font-size: 9px;
  font-weight: 500;
}

.indicator-error {
  background-color: var(--status-error);
  color: white;
}

.indicator-body {
  background-color: var(--accent-color-blue);
  color: white;
}

.indicator-cookies {
  background-color: var(--status-warning);
  color: white;
}

.indicator-size {
  background-color: var(--input-background);
  color: var(--text-color-secondary);
}

.message-count {
  font-size: 9px;
  color: var(--text-color-secondary);
  background-color: var(--input-background);
  padding: 1px var(--spacing-xs);
  border-radius: 2px;
  font-weight: 500;
  border: 1px solid var(--border-color);
}

.session-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.session-url {
  font-weight: 500;
  font-size: var(--font-size-small);
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
}

.session-status {
  font-size: 9px;
  border-radius: 2px;
  padding: 1px var(--spacing-xs);
  font-weight: 600;
}

.status-2xx { background-color: var(--status-success); color: white; }
.status-3xx { background-color: var(--accent-color-blue); color: white; }
.status-4xx, .status-5xx { background-color: var(--status-error); color: white; }

.websocket-state {
  font-size: 9px;
  border-radius: 2px;
  padding: 1px var(--spacing-xs);
  font-weight: 600;
}

.ws-state-connected { background-color: var(--status-success); color: white; }
.ws-state-disconnected { background-color: var(--text-color-secondary); color: white; }
.ws-state-connecting { background-color: var(--accent-color-blue); color: white; }
.ws-state-error { background-color: var(--status-error); color: white; }
.ws-state-unknown { background-color: var(--text-color-secondary); color: white; }

.session-duration {
  font-size: 9px;
  color: var(--text-color-secondary);
  background-color: var(--input-background);
  padding: 1px var(--spacing-xs);
  border-radius: 2px;
}

.session-time {
  font-size: 9px;
  color: var(--line-number-color);
  margin-left: auto;
  font-weight: 500;
}

.session-item.active .session-url,
.session-item.active .session-time,
.session-item.active .session-duration {
  color: rgba(255, 255, 255, 0.9);
}

.session-item.active .indicator-size {
  background-color: rgba(255, 255, 255, 0.2);
  color: white;
}

.context-menu {
  position: fixed;
  background: var(--bg-color-light);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-heavy);
  z-index: 1000;
  min-width: 120px;
  overflow: hidden;
  font-size: var(--font-size-small);
}

.context-menu-item {
  padding: var(--spacing-sm) var(--spacing-sm);
  cursor: pointer;
  color: var(--text-color-primary);
  transition: var(--transition-fast);
  border-bottom: 1px solid transparent;
  display: flex;
  align-items: center;
}

.context-menu-item:hover {
  background-color: var(--hover-color);
  color: var(--text-color-primary);
}

.context-menu-item:not(:last-child) {
  border-bottom-color: var(--border-color);
}

.context-menu-divider {
  height: 1px;
  background-color: var(--border-color);
  margin: 2px 0;
}

@media (max-width: 600px) {
  .session-item {
    padding: 4px 6px;
  }

  .session-header {
    flex-wrap: wrap;
  }

  .session-indicators {
    flex-wrap: wrap;
  }

  .session-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
  }

  .session-time {
    margin-left: 0;
  }

  .context-menu {
    min-width: 100px;
  }
}
</style>