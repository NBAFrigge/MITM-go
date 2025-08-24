<template>
  <div class="sessions-panel">
    <div class="sessions-header">
      <h2 class="sessions-title">Sessions ({{ sessions.length }})</h2>
    </div>

    <!-- Comparison Mode Banner -->
    <div v-if="comparisonMode" class="comparison-banner">
      <div class="comparison-info">
        <span class="comparison-icon">🔍</span>
        <span class="comparison-text">
          Select another session to compare with "{{ getSessionDisplayName(comparisonBaseSession) }}"
        </span>
      </div>
      <button @click="cancelComparison" class="cancel-comparison-btn">
        Cancel
      </button>
    </div>

    <div class="sessions-list">
      <SessionItem
          v-for="session in sortedSessions"
          :key="session.id"
          :session="session"
          :is-selected="selectedSessionId === session.id"
          @select="$emit('select-session', session.id)"
          @send-to-compare="handleSendToCompare"
      />
      <div v-if="sessions.length === 0" class="no-sessions">
        No sessions available
      </div>
    </div>
  </div>
</template>

<script>
import SessionItem from './SessionItem.vue'
import { CompareSessions } from '../../wailsjs/go/main/App'

export default {
  name: 'SessionsList',
  components: {
    SessionItem
  },
  props: {
    sessions: {
      type: Array,
      required: true
    },
    selectedSessionId: {
      type: [String, Number],
      default: null
    }
  },
  emits: ['select-session', 'send-to-compare'],
  data() {
    return {}
  },
  computed: {
    sortedSessions() {
      return [...this.sessions].sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
    },
    httpSessionsCount() {
      const httpCount = this.sessions.filter(session => session.type !== 'WebSocketSession').length;
      return httpCount;
    },
    webSocketSessionsCount() {
      const wsCount = this.sessions.filter(session => session.type === 'WebSocketSession').length;
      return wsCount;
    }
  },
  methods: {
    handleSendToCompare(sessionId) {
      this.$emit('send-to-compare', sessionId);
    },

    getSessionDisplayName(session) {
      if (!session) return '';
      return `${session.method || 'GET'} ${session.url || session.path || ''}`.substring(0, 50) + '...';
    }
  }
}
</script>

<style scoped>
.sessions-panel {
  width: 50%;
  min-width: 300px;
  background-color: var(--bg-color-dark);
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  font-family: var(--font-family);
}

.sessions-header {
  background-color: var(--bg-color-light);
  border-bottom: 1px solid var(--border-color);
  padding: var(--spacing-md);
}

.sessions-title {
  margin: 0 0 var(--spacing-sm) 0;
  font-weight: 600;
  text-align: center;
  font-size: var(--font-size-normal);
  color: var(--text-color-primary);
}

.sessions-stats {
  display: flex;
  justify-content: center;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: 2px var(--spacing-sm);
  background-color: var(--bg-color-dark);
  border-radius: var(--radius-sm);
  font-size: 10px;
  border: 1px solid var(--border-color);
}

.stat-label {
  color: var(--text-color-secondary);
  font-weight: 500;
}

.stat-value {
  color: var(--text-color-primary);
  font-weight: 600;
  background-color: var(--input-background);
  padding: 1px var(--spacing-xs);
  border-radius: 2px;
  min-width: 12px;
  text-align: center;
}

.comparison-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-sm) var(--spacing-sm);
  background-color: var(--accent-color-blue);
  color: white;
  border-bottom: 1px solid var(--border-color);
  font-size: var(--font-size-small);
}

.comparison-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.comparison-icon {
  font-size: 12px;
}

.comparison-text {
  font-weight: 500;
  font-size: 10px;
}

.cancel-comparison-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 10px;
  transition: var(--transition-fast);
  font-family: inherit;
}

.cancel-comparison-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.sessions-list {
  list-style: none;
  padding: 0;
  margin: 0;
  overflow-y: auto;
  flex: 1;
}

.no-sessions {
  padding: var(--spacing-xl);
  text-align: center;
  color: var(--text-color-secondary);
  font-style: italic;
  font-size: var(--font-size-small);
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-color-secondary);
  font-size: var(--font-size-small);
}

@media (max-width: 768px) {
  .sessions-panel {
    width: 100%;
    min-width: auto;
  }

  .sessions-title {
    font-size: 12px;
  }

  .sessions-header {
    padding: 8px;
  }

  .sessions-stats {
    gap: 4px;
  }

  .stat-item {
    font-size: 9px;
    padding: 1px 4px;
  }

  .comparison-banner {
    padding: 4px 6px;
    flex-direction: column;
    gap: 4px;
  }

  .comparison-text {
    font-size: 9px;
  }
}
</style>