<template>
  <div class="compare-sessions">
    <!-- Header -->
    <div class="header">
      <h2>Session Comparison</h2>
      <p>Use "Send to Compare" from session context menu to select sessions</p>
    </div>

    <!-- Sessions Display -->
    <div class="sessions-section">
      <div class="sessions-grid">
        <!-- Session 1 -->
        <div class="session-slot" :class="{ filled: session1, empty: !session1 }">
          <div class="slot-header">
            <span class="slot-label">Session 1</span>
            <button
                v-if="session1"
                @click="removeSession(1)"
                class="remove-btn"
                title="Remove session"
            >
              ×
            </button>
          </div>

          <div v-if="session1" class="session-card">
            <div class="session-badges">
              <span class="method-badge" :class="`method-${session1.method?.toLowerCase()}`">
                {{ session1.method }}
              </span>
              <span class="status-badge" :class="getStatusClass(session1.status)">
                {{ session1.status }}
              </span>
            </div>
            <div class="session-url">{{ session1.url }}</div>
            <div class="session-meta">
              <span class="session-time">{{ formatTime(session1.timestamp) }}</span>
              <span v-if="session1.duration" class="session-duration">{{ session1.duration }}ms</span>
            </div>
          </div>

          <div v-else class="empty-slot">
            <div class="empty-icon">📋</div>
            <span>No session selected</span>
            <small>Right-click a session and choose "Send to Compare"</small>
          </div>
        </div>

        <!-- VS Divider -->
        <div class="vs-divider">VS</div>

        <!-- Session 2 -->
        <div class="session-slot" :class="{ filled: session2, empty: !session2 }">
          <div class="slot-header">
            <span class="slot-label">Session 2</span>
            <button
                v-if="session2"
                @click="removeSession(2)"
                class="remove-btn"
                title="Remove session"
            >
              ×
            </button>
          </div>

          <div v-if="session2" class="session-card">
            <div class="session-badges">
              <span class="method-badge" :class="`method-${session2.method?.toLowerCase()}`">
                {{ session2.method }}
              </span>
              <span class="status-badge" :class="getStatusClass(session2.status)">
                {{ session2.status }}
              </span>
            </div>
            <div class="session-url">{{ session2.url }}</div>
            <div class="session-meta">
              <span class="session-time">{{ formatTime(session2.timestamp) }}</span>
              <span v-if="session2.duration" class="session-duration">{{ session2.duration }}ms</span>
            </div>
          </div>

          <div v-else class="empty-slot">
            <div class="empty-icon">📋</div>
            <span>No session selected</span>
            <small>Right-click a session and choose "Send to Compare"</small>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div v-if="session1 && session2" class="actions">
        <button
            @click="performComparison"
            :disabled="loading"
            class="btn-compare"
        >
          {{ loading ? 'Comparing...' : 'Compare Sessions' }}
        </button>
        <button
            @click="clearAll"
            class="btn-clear"
        >
          Clear All
        </button>
      </div>
    </div>

    <!-- Results -->
    <div v-if="comparisonResult" class="results-section">
      <div class="results-header">
        <h3>Comparison Results</h3>
        <button @click="exportResults" class="btn-export">Export</button>
      </div>

      <!-- Summary -->
      <div class="summary" :class="{ identical: !hasDifferences, different: hasDifferences }">
        <div class="summary-icon">
          {{ hasDifferences ? '!' : '✓' }}
        </div>
        <div class="summary-text">
          <span class="summary-title">
            {{ hasDifferences ? `${differenceCount} Differences Found` : 'Sessions Identical' }}
          </span>
          <span class="summary-desc">
            {{ hasDifferences ? 'Review changes below' : 'No differences detected' }}
          </span>
        </div>
      </div>

      <!-- Differences -->
      <div v-if="hasDifferences" class="differences">
        <!-- URL -->
        <div v-if="differences.url" class="diff-block">
          <div class="diff-header">URL</div>
          <div class="diff-content">
            <div class="diff-item original">
              <div class="diff-label">Session 1</div>
              <div class="diff-value">{{ differences.url.original }}</div>
            </div>
            <div class="diff-item changed">
              <div class="diff-label">Session 2</div>
              <div class="diff-value">{{ differences.url.other }}</div>
            </div>
          </div>
        </div>

        <!-- Method -->
        <div v-if="differences.method" class="diff-block">
          <div class="diff-header">Method</div>
          <div class="diff-content">
            <div class="diff-item original">
              <div class="diff-label">Session 1</div>
              <span class="method-badge" :class="`method-${differences.method.original?.toLowerCase()}`">
                {{ differences.method.original }}
              </span>
            </div>
            <div class="diff-item changed">
              <div class="diff-label">Session 2</div>
              <span class="method-badge" :class="`method-${differences.method.other?.toLowerCase()}`">
                {{ differences.method.other }}
              </span>
            </div>
          </div>
        </div>

        <!-- Content Type -->
        <div v-if="differences.contentType" class="diff-block">
          <div class="diff-header">Content-Type</div>
          <div class="diff-content">
            <div class="diff-item original">
              <div class="diff-label">Session 1</div>
              <div class="diff-value">{{ differences.contentType.original || 'Not set' }}</div>
            </div>
            <div class="diff-item changed">
              <div class="diff-label">Session 2</div>
              <div class="diff-value">{{ differences.contentType.other || 'Not set' }}</div>
            </div>
          </div>
        </div>

        <!-- Headers -->
        <div v-if="differences.headers" class="diff-block">
          <div class="diff-header">Headers</div>
          <div class="diff-content changes-content">
            <div v-if="differences.headers.added?.length" class="changes-group">
              <div class="changes-header added">
                <span class="change-type">Added</span>
                <span class="change-count">{{ differences.headers.added.length }}</span>
              </div>
              <div class="changes-list">
                <span v-for="header in differences.headers.added" :key="header" class="change-tag">
                  {{ header }}
                </span>
              </div>
            </div>

            <div v-if="differences.headers.removed?.length" class="changes-group">
              <div class="changes-header removed">
                <span class="change-type">Removed</span>
                <span class="change-count">{{ differences.headers.removed.length }}</span>
              </div>
              <div class="changes-list">
                <span v-for="header in differences.headers.removed" :key="header" class="change-tag">
                  {{ header }}
                </span>
              </div>
            </div>

            <div v-if="differences.headers.modified?.length" class="changes-group">
              <div class="changes-header modified">
                <span class="change-type">Modified</span>
                <span class="change-count">{{ differences.headers.modified.length }}</span>
              </div>
              <div class="changes-list">
                <span v-for="header in differences.headers.modified" :key="header" class="change-tag">
                  {{ header }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Cookies -->
        <div v-if="differences.cookies" class="diff-block">
          <div class="diff-header">Cookies</div>
          <div class="diff-content changes-content">
            <div v-if="differences.cookies.added?.length" class="changes-group">
              <div class="changes-header added">
                <span class="change-type">Added</span>
                <span class="change-count">{{ differences.cookies.added.length }}</span>
              </div>
              <div class="changes-list">
                <span v-for="cookie in differences.cookies.added" :key="cookie" class="change-tag">
                  {{ cookie }}
                </span>
              </div>
            </div>

            <div v-if="differences.cookies.removed?.length" class="changes-group">
              <div class="changes-header removed">
                <span class="change-type">Removed</span>
                <span class="change-count">{{ differences.cookies.removed.length }}</span>
              </div>
              <div class="changes-list">
                <span v-for="cookie in differences.cookies.removed" :key="cookie" class="change-tag">
                  {{ cookie }}
                </span>
              </div>
            </div>

            <div v-if="differences.cookies.modified?.length" class="changes-group">
              <div class="changes-header modified">
                <span class="change-type">Modified</span>
                <span class="change-count">{{ differences.cookies.modified.length }}</span>
              </div>
              <div class="changes-list">
                <span v-for="cookie in differences.cookies.modified" :key="cookie" class="change-tag">
                  {{ cookie }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Body -->
        <div v-if="differences.body" class="diff-block">
          <div class="diff-header">Request Body</div>
          <div class="diff-content body-content">
            <div class="body-section">
              <div class="diff-label">Session 1</div>
              <pre class="body-text">{{ formatBody(differences.body.original) }}</pre>
            </div>
            <div class="body-section">
              <div class="diff-label">Session 2</div>
              <pre class="body-text">{{ formatBody(differences.body.other) }}</pre>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { CompareSessions } from '../../wailsjs/go/main/App'

export default {
  name: 'CompareSessions',
  props: {
    sessions: {
      type: Array,
      required: true
    },
    selectedSessions: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      session1Id: null,
      session2Id: null,
      comparisonResult: null,
      loading: false
    }
  },
  computed: {
    session1() {
      return this.session1Id ? this.sessions.find(s => s.id === this.session1Id) : null;
    },
    session2() {
      return this.session2Id ? this.sessions.find(s => s.id === this.session2Id) : null;
    },
    differences() {
      return this.comparisonResult?.differences || {};
    },
    hasDifferences() {
      return this.comparisonResult?.differences !== "No differences found" &&
          Object.keys(this.differences).length > 0;
    },
    differenceCount() {
      return Object.keys(this.differences).length;
    }
  },
  watch: {
    selectedSessions: {
      handler(newSessions) {
        newSessions.forEach(sessionId => {
          this.addSession(sessionId);
        });
      },
      immediate: true
    }
  },
  methods: {
    addSession(sessionId) {
      if (!sessionId) return;
      if (this.session1Id === sessionId || this.session2Id === sessionId) return;
      if (!this.session1Id) {
        this.session1Id = sessionId;
      } else if (!this.session2Id) {
        this.session2Id = sessionId;
      } else {
        this.session2Id = sessionId;
      }
      this.comparisonResult = null;
    },
    removeSession(slot) {
      if (slot === 1) {
        this.session1Id = null;
      } else {
        this.session2Id = null;
      }
      this.comparisonResult = null;
    },
    clearAll() {
      this.session1Id = null;
      this.session2Id = null;
      this.comparisonResult = null;
    },
    formatTime(timestamp) {
      return new Date(timestamp).toLocaleString();
    },
    getStatusClass(status) {
      if (status >= 200 && status < 300) return 'status-success';
      if (status >= 300 && status < 400) return 'status-redirect';
      if (status >= 400 && status < 500) return 'status-client-error';
      if (status >= 500) return 'status-server-error';
      return '';
    },
    async performComparison() {
      if (!this.session1Id || !this.session2Id) return;
      this.loading = true;
      try {
        this.comparisonResult = await CompareSessions(this.session1Id, this.session2Id);
      } catch (error) {
        console.error('Comparison failed:', error);
        alert('Failed to compare sessions.');
      } finally {
        this.loading = false;
      }
    },
    formatBody(body) {
      if (!body) return 'Empty';
      try {
        return JSON.stringify(JSON.parse(body), null, 2);
      } catch {
        return body;
      }
    },
    async exportResults() {
      if (!this.comparisonResult) return;
      try {
        const dataStr = JSON.stringify(this.comparisonResult, null, 2);
        const blob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `comparison-${Date.now()}.json`;
        a.click();
        URL.revokeObjectURL(url);
      } catch (error) {
        console.error('Export failed:', error);
      }
    }
  }
}
</script>

<style scoped>
.compare-sessions {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color-dark);
  color: var(--text-color-primary);
  font-family: var(--font-family);
  overflow: hidden;
}

.header {
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-color-medium);
  flex-shrink: 0;
}

.header h2 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: var(--font-size-normal);
  font-weight: 600;
  color: var(--text-color-primary);
}

.header p {
  margin: 0;
  color: var(--text-color-secondary);
  font-size: var(--font-size-small);
}

.sessions-section {
  padding: var(--spacing-lg);
  background-color: var(--bg-color-dark);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.sessions-grid {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: var(--spacing-lg);
  align-items: start;
  margin-bottom: var(--spacing-lg);
}

.session-slot {
  min-height: 160px;
  display: flex;
  flex-direction: column;
}

.slot-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
}

.slot-label {
  font-weight: 600;
  color: var(--text-color-primary);
  font-size: var(--font-size-small);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.remove-btn {
  background: none;
  border: none;
  color: var(--text-color-secondary);
  cursor: pointer;
  font-size: 16px;
  padding: var(--spacing-xs);
  border-radius: var(--radius-sm);
  transition: var(--transition-fast);
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remove-btn:hover {
  background-color: var(--status-error);
  color: white;
}

.session-card {
  background-color: var(--bg-color-light);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  flex: 1;
  transition: var(--transition-normal);
}

.session-slot.filled .session-card {
  border-left: 3px solid var(--accent-color-blue);
}

.session-badges {
  display: flex;
  gap: 6px;
  align-items: center;
}

.method-badge {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.method-get { background-color: #3369d6; color: white; }
.method-post { background-color: #d69e2e; color: white; }
.method-put { background-color: #38a169; color: white; }
.method-delete { background-color: #e53e3e; color: white; }

.status-badge {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 600;
}

.status-success { background-color: #38a169; color: white; }
.status-redirect { background-color: #3182ce; color: white; }
.status-client-error,
.status-server-error { background-color: #e53e3e; color: white; }

.session-url {
  font-size: 12px;
  color: #cccccc;
  word-break: break-all;
  line-height: 1.4;
  background-color: #313131;
  padding: 6px;
  border-radius: 3px;
  border: 1px solid #555555;
}

.session-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 11px;
  color: #888888;
}

.empty-slot {
  background-color: #313131;
  border: 2px dashed #555555;
  border-radius: 6px;
  padding: 24px 12px;
  text-align: center;
  color: #888888;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex: 1;
  justify-content: center;
  transition: all 0.15s;
  min-height: 120px;
}

.empty-slot:hover {
  border-color: #3369d6;
  background-color: rgba(51, 105, 214, 0.05);
  color: #cccccc;
}

.empty-icon {
  font-size: 2rem;
  opacity: 0.5;
  filter: grayscale(1);
}

.empty-slot span {
  font-weight: 500;
  font-size: 12px;
}

.empty-slot small {
  font-size: 11px;
  opacity: 0.7;
}

.vs-divider {
  background-color: #555555;
  color: #cccccc;
  width: 40px;
  height: 40px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  justify-self: center;
  align-self: center;
  margin-top: 24px;
  font-size: 12px;
  border: 1px solid #666666;
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
}

.btn-compare,
.btn-clear {
  padding: 6px 12px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid #555555;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 28px;
}

.btn-compare {
  background-color: #3369d6;
  color: white;
  border-color: #3369d6;
}

.btn-compare:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-compare:not(:disabled):hover {
  background-color: #2952a3;
  border-color: #2952a3;
}

.btn-clear {
  background-color: #393939;
  color: #cccccc;
}

.btn-clear:hover {
  background-color: #4a4a4a;
  color: #ffffff;
}

.results-section {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  min-height: 0;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid #3c3c3c;
}

.results-header h3 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #cccccc;
}

.btn-export {
  background-color: #38a169;
  color: white;
  border: 1px solid #38a169;
  padding: 4px 8px;
  border-radius: 3px;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
  font-weight: 500;
}

.btn-export:hover {
  background-color: #2f855a;
  border-color: #2f855a;
}

.summary {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
  border: 1px solid;
}

.summary.identical {
  background-color: rgba(56, 161, 105, 0.1);
  border-color: #38a169;
}

.summary.different {
  background-color: rgba(214, 158, 46, 0.1);
  border-color: #d69e2e;
}

.summary-icon {
  width: 24px;
  height: 24px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 12px;
  flex-shrink: 0;
}

.identical .summary-icon {
  background-color: #38a169;
  color: white;
}

.different .summary-icon {
  background-color: #d69e2e;
  color: white;
}

.summary-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.summary-title {
  font-weight: 600;
  color: #cccccc;
  font-size: 12px;
}

.summary-desc {
  font-size: 11px;
  color: #888888;
}

.differences {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.diff-block {
  background-color: #393939;
  border: 1px solid #555555;
  border-radius: 6px;
  overflow: hidden;
}

.diff-header {
  padding: 8px 12px;
  background-color: #313131;
  border-bottom: 1px solid #555555;
  font-weight: 600;
  font-size: 11px;
  color: #cccccc;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.diff-header::before {
  content: '';
  width: 3px;
  height: 12px;
  background-color: #3369d6;
  border-radius: 1px;
}

.diff-content {
  padding: 12px;
}

.diff-content:not(.body-content):not(.changes-content) {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.diff-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.diff-label {
  font-size: 11px;
  font-weight: 600;
  color: #888888;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.diff-value {
  background-color: #313131;
  border: 1px solid #555555;
  border-radius: 3px;
  padding: 8px;
  font-size: 11px;
  word-break: break-all;
  line-height: 1.4;
  transition: all 0.15s;
}

.diff-value:hover {
  border-color: #3369d6;
}

.diff-item.original .diff-value {
  border-left: 3px solid #e53e3e;
  background-color: rgba(229, 62, 62, 0.05);
}

.diff-item.changed .diff-value {
  border-left: 3px solid #38a169;
  background-color: rgba(56, 161, 105, 0.05);
}

.changes-group {
  margin-bottom: 12px;
}

.changes-group:last-child {
  margin-bottom: 0;
}

.changes-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.change-type {
  font-weight: 600;
  font-size: 11px;
}

.change-count {
  background-color: #313131;
  color: #888888;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
  border: 1px solid #555555;
}

.changes-header.added .change-type { color: #38a169; }
.changes-header.removed .change-type { color: #e53e3e; }
.changes-header.modified .change-type { color: #d69e2e; }

.changes-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.change-tag {
  background-color: #313131;
  border: 1px solid #555555;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  transition: all 0.15s;
  cursor: default;
}

.change-tag:hover {
  background-color: #393939;
}

.body-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.body-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.body-text {
  background-color: #313131;
  border: 1px solid #555555;
  border-radius: 3px;
  padding: 8px;
  margin: 0;
  font-size: 11px;
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
  line-height: 1.4;
}

@media (max-width: 768px) {
  .sessions-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .vs-divider {
    justify-self: center;
    margin-top: 0;
  }

  .diff-content:not(.body-content):not(.changes-content) {
    grid-template-columns: 1fr;
  }

  .body-content {
    grid-template-columns: 1fr;
  }

  .actions {
    flex-direction: column;
    align-items: center;
  }

  .sessions-section,
  .results-section {
    padding: 12px;
  }
}
</style>