<template>
  <div class="detail-section">
    <CollapsiblePanel title="WebSocket Messages" :initial-expanded="true">
      <div class="messages-container">
        <div v-if="!messages || messages.length === 0" class="no-messages">
          No messages captured
        </div>
        <div v-else class="messages-list">
          <div class="messages-header">
            <div class="message-count">
              {{ messages.length }} message{{ messages.length !== 1 ? 's' : '' }}
            </div>
            <div class="filter-controls">
              <select v-model="directionFilter" class="filter-select">
                <option value="all">All Messages</option>
                <option value="inbound">Inbound</option>
                <option value="outbound">Outbound</option>
              </select>
              <select v-model="opcodeFilter" class="filter-select">
                <option value="all">All Types</option>
                <option value="1">Text (1)</option>
                <option value="2">Binary (2)</option>
                <option value="8">Close (8)</option>
                <option value="9">Ping (9)</option>
                <option value="10">Pong (10)</option>
              </select>
            </div>
          </div>

          <div class="messages-scroll">
            <div
                v-for="message in filteredMessages"
                :key="message.id"
                class="message-item"
                :class="getMessageClass(message)"
            >
              <div class="message-header">
                <div class="message-info">
                  <span class="message-direction" :class="getDirectionClass(message.direction)">
                    {{ message.direction.toUpperCase() }}
                  </span>
                  <span class="message-opcode">{{ getOpcodeLabel(message.opcode) }}</span>
                  <span class="message-size">{{ formatBytes(message.size) }}</span>
                  <span class="message-timestamp">{{ formatTimestamp(message.timestamp) }}</span>
                </div>
                <div class="message-flags">
                  <span v-if="message.isMasked" class="flag">MASKED</span>
                  <span v-if="message.isFragment" class="flag">FRAGMENT</span>
                </div>
              </div>
              <div class="message-payload">
                <CollapsiblePanel
                    :title="`Payload (${formatBytes(message.size)})`"
                    :nested="true"
                    :initial-expanded="false"
                >
                  <div class="payload-content">
                    <div v-if="message.payloadText" class="payload-section">
                      <div class="payload-label">Text:</div>
                      <pre class="payload-block text-payload">{{ message.payloadText }}</pre>
                    </div>
                    <div v-if="message.payload && message.payload !== message.payloadText" class="payload-section">
                      <div class="payload-label">Raw:</div>
                      <pre class="payload-block raw-payload">{{ message.payload }}</pre>
                    </div>
                  </div>
                </CollapsiblePanel>
              </div>
            </div>
          </div>
        </div>
      </div>
    </CollapsiblePanel>
  </div>
</template>

<script>
import CollapsiblePanel from './CollapsiblePanel.vue';

export default {
  name: 'WebSocketMessages',
  components: {
    CollapsiblePanel
  },
  props: {
    messages: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      directionFilter: 'all',
      opcodeFilter: 'all'
    };
  },
  computed: {
    filteredMessages() {
      let filtered = this.messages || [];

      if (this.directionFilter !== 'all') {
        filtered = filtered.filter(msg => msg.direction.toLowerCase() === this.directionFilter);
      }

      if (this.opcodeFilter !== 'all') {
        filtered = filtered.filter(msg => msg.opcode.toString() === this.opcodeFilter);
      }

      return filtered.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
    }
  },
  methods: {
    getMessageClass(message) {
      return {
        'message-inbound': message.direction.toLowerCase() === 'inbound',
        'message-outbound': message.direction.toLowerCase() === 'outbound',
        'message-control': this.isControlFrame(message.opcode)
      };
    },
    getDirectionClass(direction) {
      return {
        'direction-inbound': direction.toLowerCase() === 'inbound',
        'direction-outbound': direction.toLowerCase() === 'outbound'
      };
    },
    isControlFrame(opcode) {
      return opcode >= 8 && opcode <= 15;
    },
    getOpcodeLabel(opcode) {
      const opcodes = {
        0: 'Continuation',
        1: 'Text',
        2: 'Binary',
        8: 'Close',
        9: 'Ping',
        10: 'Pong'
      };
      return opcodes[opcode] || `Opcode ${opcode}`;
    },
    formatTimestamp(timestamp) {
      if (!timestamp) return 'N/A';
      const date = new Date(timestamp);
      return date.toLocaleTimeString() + '.' + String(date.getMilliseconds()).padStart(3, '0');
    },
    formatBytes(bytes) {
      if (!bytes) return '0 B';
      const sizes = ['B', 'KB', 'MB', 'GB'];
      const i = Math.floor(Math.log(bytes) / Math.log(1024));
      return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
    }
  }
}
</script>

<style scoped>
.detail-section {
  margin-bottom: 0;
  font-family: var(--font-family);
}

.messages-container {
  max-height: 70vh;
  display: flex;
  flex-direction: column;
}

.no-messages {
  text-align: center;
  padding: var(--spacing-lg);
  color: var(--text-color-secondary);
  font-style: italic;
  font-size: var(--font-size-small);
}

.messages-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-xs) 0;
  margin-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--border-color);
}

.message-count {
  font-weight: 600;
  color: var(--text-color-primary);
  font-size: var(--font-size-small);
}

.filter-controls {
  display: flex;
  gap: var(--spacing-xs);
}

.filter-select {
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--input-border);
  border-radius: var(--radius-sm);
  background-color: var(--input-background);
  color: var(--text-color-primary);
  font-size: var(--font-size-small);
  font-family: inherit;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition-fast);
  min-width: 120px;
}

.filter-select:hover {
  border-color: var(--accent-color-blue);
  background-color: var(--hover-color);
}

.filter-select:focus {
  outline: none;
  border-color: var(--input-focus);
  box-shadow: 0 0 0 2px rgba(104, 151, 187, 0.2);
}

.filter-select option {
  background-color: var(--bg-color-medium);
  color: var(--text-color-primary);
  padding: var(--spacing-sm);
  font-weight: 500;
}

.messages-scroll {
  flex: 1;
  overflow-y: auto;
  max-height: 60vh;
}

.message-item {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  margin-bottom: var(--spacing-xs);
  background-color: var(--bg-color-medium);
  transition: var(--transition-fast);
}

.message-item:hover {
  background-color: var(--hover-color);
  border-color: var(--accent-color-blue);
}

.message-item.message-inbound {
  border-left: 3px solid var(--accent-color-blue);
}

.message-item.message-outbound {
  border-left: 3px solid var(--status-success);
}

.message-item.message-control {
  border-left: 3px solid var(--status-warning);
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--bg-color-light);
  border-bottom: 1px solid var(--border-color);
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
}

.message-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  font-size: var(--font-size-small);
}

.message-direction {
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-weight: 600;
  font-size: 10px;
  color: white;
  min-width: 60px;
  text-align: center;
  letter-spacing: 0.5px;
}

.direction-inbound {
  background-color: var(--accent-color-blue);
}

.direction-outbound {
  background-color: var(--status-success);
}

.message-opcode {
  color: var(--text-color-primary);
  font-weight: 600;
  background-color: var(--input-background);
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: 10px;
  border: 1px solid var(--border-color);
}

.message-size {
  color: var(--number-color);
  font-weight: 600;
  font-family: var(--font-family);
  background-color: var(--input-background);
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 11px;
}

.message-timestamp {
  color: var(--text-color-primary);
  font-family: var(--font-family);
  font-weight: 600;
  background-color: var(--input-background);
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 11px;
}

.message-flags {
  display: flex;
  gap: var(--spacing-xs);
}

.flag {
  padding: 2px var(--spacing-xs);
  border-radius: var(--radius-sm);
  background-color: var(--status-warning);
  color: var(--bg-color-dark);
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.message-payload {
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--bg-color-dark);
  border-radius: 0 0 var(--radius-sm) var(--radius-sm);
}

.payload-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.payload-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.payload-label {
  font-weight: 600;
  font-size: var(--font-size-small);
  color: var(--field-color);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.payload-block {
  background-color: var(--input-background);
  color: var(--text-color-primary);
  padding: var(--spacing-md);
  border-radius: var(--radius-sm);
  overflow-x: auto;
  font-family: var(--font-family);
  font-size: var(--font-size-small);
  margin: 0;
  white-space: pre-wrap;
  border: 1px solid var(--border-color);
  max-height: 200px;
  overflow-y: auto;
  line-height: 1.4;
}

.text-payload {
  border-left: 3px solid var(--string-color);
}

.raw-payload {
  border-left: 3px solid var(--comment-color);
}

@media (max-width: 768px) {
  .messages-header {
    flex-direction: column;
    gap: 4px;
    align-items: stretch;
  }

  .message-info {
    flex-wrap: wrap;
    gap: 4px;
  }
}
</style>