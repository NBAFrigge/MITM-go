<template>
  <div class="session-preview">
    <div class="preview-header">
      <span class="method-badge" :class="`method-${session.method?.toLowerCase()}`">
        {{ session.method }}
      </span>
      <span class="status-badge" :class="getStatusClass(session.status)">
        {{ session.status }}
      </span>
    </div>

    <div class="preview-url">
      {{ session.url }}
    </div>

    <div class="preview-meta">
      <span class="timestamp">{{ formatTime(session.timestamp) }}</span>
      <span v-if="session.duration" class="duration">{{ session.duration }}ms</span>
    </div>

    <div v-if="hasDetails" class="preview-details">
      <div v-if="hasHeaders" class="detail-item">
        <span class="detail-label">Headers:</span>
        <span class="detail-value">{{ headerCount }}</span>
      </div>
      <div v-if="hasCookies" class="detail-item">
        <span class="detail-label">Cookies:</span>
        <span class="detail-value">{{ cookieCount }}</span>
      </div>
      <div v-if="hasBody" class="detail-item">
        <span class="detail-label">Body:</span>
        <span class="detail-value">{{ bodySize }}</span>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'SessionPreview',
  props: {
    session: {
      type: Object,
      required: true
    }
  },
  computed: {
    hasHeaders() {
      return this.session.requestHeaders && Object.keys(this.session.requestHeaders).length > 0;
    },
    hasCookies() {
      return this.session.requestCookies && Object.keys(this.session.requestCookies).length > 0;
    },
    hasBody() {
      return this.session.requestBody && this.session.requestBody.trim() !== '';
    },
    hasDetails() {
      return this.hasHeaders || this.hasCookies || this.hasBody;
    },
    headerCount() {
      return Object.keys(this.session.requestHeaders || {}).length;
    },
    cookieCount() {
      return Object.keys(this.session.requestCookies || {}).length;
    },
    bodySize() {
      if (!this.hasBody) return '0 B';
      const size = new Blob([this.session.requestBody]).size;
      return this.formatSize(size);
    }
  },
  methods: {
    formatTime(timestamp) {
      return new Date(timestamp).toLocaleString();
    },
    getStatusClass(status) {
      if (status >= 200 && status < 300) return 'status-2xx';
      if (status >= 300 && status < 400) return 'status-3xx';
      if (status >= 400 && status < 500) return 'status-4xx';
      if (status >= 500) return 'status-5xx';
      return '';
    },
    formatSize(bytes) {
      if (bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    }
  }
}
</script>

<style scoped>
.session-preview {
  background-color: #393939;
  border: 1px solid #555555;
  border-radius: 6px;
  padding: 12px;
  font-size: 11px;
  font-family: 'JetBrains Mono', 'SF Mono', Consolas, monospace;
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
}

.method-badge {
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: bold;
  text-transform: uppercase;
}

.method-get { background-color: #3369d6; color: white; }
.method-post { background-color: #38a169; color: white; }
.method-put { background-color: #d69e2e; color: white; }
.method-delete { background-color: #e53e3e; color: white; }

.status-badge {
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
}

.status-2xx { background-color: #38a169; color: white; }
.status-3xx { background-color: #3369d6; color: white; }
.status-4xx, .status-5xx { background-color: #e53e3e; color: white; }

.preview-url {
  color: #cccccc;
  word-break: break-all;
  margin-bottom: 8px;
  font-size: 11px;
}

.preview-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #888888;
  font-size: 10px;
  margin-bottom: 8px;
}

.preview-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-top: 6px;
  border-top: 1px solid #555555;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 10px;
}

.detail-label {
  color: #888888;
}

.detail-value {
  color: #cccccc;
  font-weight: 500;
}
</style>