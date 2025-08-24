<template>
  <div class="detail-section">
    <CollapsiblePanel title="WebSocket Connection" :initial-expanded="true">
      <DetailList>
        <DetailItem label="State">
          <span class="ws-state-badge" :class="getStateClass(webSocketData.state)">
            {{ getWebSocketStateName(webSocketData.state) }}
          </span>
        </DetailItem>
        <DetailItem label="Connected At" v-if="webSocketData.connectedAt">
          <span>{{ formatTimestamp(webSocketData.connectedAt) }}</span>
        </DetailItem>
        <DetailItem label="Disconnected At" v-if="webSocketData.disconnectedAt > webSocketData.connectedAt">
          <span>{{ formatTimestamp(webSocketData.disconnectedAt) }}</span>
        </DetailItem>
        <DetailItem label="Connection Duration">
          <span>{{ formatDuration(webSocketData.connectionDuration) }}</span>
        </DetailItem>
        <DetailItem label="Subprotocol" v-if="webSocketData.subprotocol">
          <span>{{ webSocketData.subprotocol }}</span>
        </DetailItem>
        <DetailItem label="Extensions" v-if="webSocketData.extensions && webSocketData.extensions.length">
          <span>{{ webSocketData.extensions.join(', ') }}</span>
        </DetailItem>
        <DetailItem label="Close Code" v-if="webSocketData.closeCode">
          <span>{{ webSocketData.closeCode }}</span>
        </DetailItem>
        <DetailItem label="Close Reason" v-if="webSocketData.closeReason">
          <span>{{ webSocketData.closeReason }}</span>
        </DetailItem>
      </DetailList>

      <CollapsiblePanel title="Message Statistics" :nested="true" :initial-expanded="true">
        <DetailList>
          <DetailItem label="Total Messages">
            <span>{{ webSocketData.messageStats.totalMessages }}</span>
          </DetailItem>
          <DetailItem label="Inbound Messages">
            <span>{{ webSocketData.messageStats.inboundMessages }}</span>
          </DetailItem>
          <DetailItem label="Outbound Messages">
            <span>{{ webSocketData.messageStats.outboundMessages }}</span>
          </DetailItem>
          <DetailItem label="Text Messages">
            <span>{{ webSocketData.messageStats.textMessages }}</span>
          </DetailItem>
          <DetailItem label="Binary Messages">
            <span>{{ webSocketData.messageStats.binaryMessages }}</span>
          </DetailItem>
          <DetailItem label="Total Bytes">
            <span>{{ formatBytes(webSocketData.messageStats.totalBytes) }}</span>
          </DetailItem>
          <DetailItem label="Inbound Bytes">
            <span>{{ formatBytes(webSocketData.messageStats.inboundBytes) }}</span>
          </DetailItem>
          <DetailItem label="Outbound Bytes">
            <span>{{ formatBytes(webSocketData.messageStats.outboundBytes) }}</span>
          </DetailItem>
        </DetailList>
      </CollapsiblePanel>

      <CollapsiblePanel title="Upgrade Request" :nested="true" :initial-expanded="false">
        <DetailList>
          <DetailItem label="Method">
            <span>{{ webSocketData.upgradeRequest.method }}</span>
          </DetailItem>
          <DetailItem label="URL">
            <span>{{ webSocketData.upgradeRequest.url }}</span>
          </DetailItem>
        </DetailList>
        <CollapsiblePanel title="Headers" :nested="true" :initial-expanded="false">
          <pre class="code-block">{{ formatHeaders(webSocketData.upgradeRequest.headers) }}</pre>
        </CollapsiblePanel>
        <CollapsiblePanel title="Cookies" :nested="true" :initial-expanded="false">
          <pre class="code-block">{{ formatCookies(webSocketData.upgradeRequest.cookies) }}</pre>
        </CollapsiblePanel>
        <CollapsiblePanel title="Body" :nested="true" :initial-expanded="false" v-if="webSocketData.upgradeRequest.body">
          <pre class="code-block">{{ formatBody(webSocketData.upgradeRequest.body) }}</pre>
        </CollapsiblePanel>
      </CollapsiblePanel>

      <CollapsiblePanel title="Upgrade Response" :nested="true" :initial-expanded="false">
        <DetailList>
          <DetailItem label="Status Code">
            <span class="status-badge" :class="getStatusClass(webSocketData.upgradeResponse.statusCode)">
              {{ webSocketData.upgradeResponse.statusCode }}
            </span>
          </DetailItem>
          <DetailItem label="Status">
            <span>{{ webSocketData.upgradeResponse.status }}</span>
          </DetailItem>
        </DetailList>
        <CollapsiblePanel title="Headers" :nested="true" :initial-expanded="false">
          <pre class="code-block">{{ formatHeaders(webSocketData.upgradeResponse.headers) }}</pre>
        </CollapsiblePanel>
        <CollapsiblePanel title="Cookies" :nested="true" :initial-expanded="false">
          <pre class="code-block">{{ formatCookies(webSocketData.upgradeResponse.cookies) }}</pre>
        </CollapsiblePanel>
        <CollapsiblePanel title="Body" :nested="true" :initial-expanded="false" v-if="webSocketData.upgradeResponse.body">
          <pre class="code-block">{{ formatBody(webSocketData.upgradeResponse.body) }}</pre>
        </CollapsiblePanel>
      </CollapsiblePanel>
    </CollapsiblePanel>
  </div>
</template>

<script>
import CollapsiblePanel from './CollapsiblePanel.vue';
import DetailList from './DetailList.vue';
import DetailItem from './DetailItem.vue';

export default {
  name: 'WebSocketDetails',
  components: {
    CollapsiblePanel,
    DetailList,
    DetailItem
  },
  props: {
    webSocketData: {
      type: Object,
      required: true
    }
  },
  methods: {
    getStateClass(state) {
      switch (state) {
        case 0:
          return 'state-connecting';
        case 1:
          return 'state-connected';
        case 3:
          return 'state-disconnected';
        case 4:
          return 'state-error';
        default:
          return 'state-unknown';

      }
    },
    getWebSocketStateName(state) {
      switch (state) {
        case 0:
          return 'Connecting';
        case 1:
          return 'Connected';
        case 3:
          return 'Disconnected';
        case 4:
          return 'Error';
        default:
          return 'Unknown';
      }
    },

    getStatusClass(status) {
      if (status >= 200 && status < 300) return 'status-2xx';
      if (status >= 300 && status < 400) return 'status-3xx';
      if (status >= 400 && status < 500) return 'status-4xx';
      if (status >= 500) return 'status-5xx';
      return '';
    },
    formatTimestamp(timestamp) {
      if (!timestamp) return 'N/A';
      return new Date(timestamp).toLocaleString();
    },
    formatDuration(milliseconds) {
      if (!milliseconds) return '0ms';
      if (milliseconds < 1000) return `${milliseconds}ms`;
      if (milliseconds < 60000) return `${(milliseconds / 1000).toFixed(2)}s`;
      return `${(milliseconds / 60000).toFixed(2)}m`;
    },
    formatBytes(bytes) {
      if (!bytes) return '0 B';
      const sizes = ['B', 'KB', 'MB', 'GB'];
      const i = Math.floor(Math.log(bytes) / Math.log(1024));
      return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
    },
    formatHeaders(headers) {
      if (!headers || (!headers.order && !headers.entries)) return 'No headers';
      if (headers.order && headers.entries) {
        const result = headers.order
            .filter(key => headers.entries[key] !== undefined)
            .map(key => `${key}: ${headers.entries[key]}`)
            .join('\n');
        return result || 'No headers found';
      }
      return JSON.stringify(headers, null, 2);
    },
    formatCookies(cookies) {
      if (!cookies) return 'No cookies';
      if (typeof cookies === 'object' && Object.keys(cookies).length === 0) {
        return 'No cookies';
      }
      if (typeof cookies === 'object' && cookies !== null) {
        const cookieEntries = Object.entries(cookies)
            .map(([key, value]) => `${key}: ${value}`)
            .join('\n');
        return cookieEntries || 'No cookies found';
      }
      return JSON.stringify(cookies, null, 2);
    },
    formatBody(body) {
      if (!body) return 'No body';
      try {
        const parsed = JSON.parse(body);
        return JSON.stringify(parsed, null, 2);
      } catch (e) {
        return body;
      }
    }
  }
}
</script>

<style scoped>
.detail-section {
  margin-bottom: 0;
  font-family: var(--font-family);
}

.ws-state-badge {
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-weight: 600;
  font-size: var(--font-size-small);
  color: white;
}

.state-connected {
  background-color: var(--status-success);
}

.state-disconnected {
  background-color: var(--text-color-secondary);
}

.state-connecting {
  background-color: var(--accent-color-blue);
}

.state-error {
  background-color: var(--status-error);
}

.state-unknown {
  background-color: var(--text-color-secondary);
}

.status-badge {
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-weight: 600;
  font-size: var(--font-size-small);
  color: white;
}

.status-2xx {
  background-color: var(--status-success);
}

.status-3xx {
  background-color: var(--accent-color-blue);
}

.status-4xx {
  background-color: var(--status-error);
}

.status-5xx {
  background-color: var(--status-error);
}

.code-block {
  background-color: var(--input-background);
  color: var(--text-color-primary);
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  overflow-x: auto;
  font-family: var(--font-family);
  font-size: var(--font-size-small);
  margin: 0;
  white-space: pre-wrap;
  border: 1px solid var(--border-color);
}
</style>