<template>
  <div class="detail-section">
    <h3 class="section-title">Response</h3>

    <DetailList>
      <DetailItem label="Response Code">
        <span class="status-badge" :class="getStatusClass(session.status)">{{ session.status }}</span>
      </DetailItem>
    </DetailList>

    <CollapsiblePanel v-if="session.error" title="Error" :nested="true" :initial-expanded="true">
      <pre class="code-block error-block">{{ formatError(session.error) }}</pre>
    </CollapsiblePanel>

    <template v-if="!session.error">
      <CollapsiblePanel title="Headers" :nested="true" :initial-expanded="true">
        <pre class="code-block">{{ formatHeaders(session.responseHeaders) }}</pre>
      </CollapsiblePanel>

      <CollapsiblePanel
          v-if="hasResponseCookies"
          title="Cookies"
          :nested="true"
          :initial-expanded="true"
      >
        <pre class="code-block">{{ formatCookies(session.responseCookies) }}</pre>
      </CollapsiblePanel>

      <CollapsiblePanel
          v-if="hasResponseBody"
          title="Body"
          :nested="true"
          :initial-expanded="true"
      >
        <JsonViewer v-if="isJsonResponseBody" :json="responseBodyAsJson" />
        <pre v-else class="code-block">{{ session.responseBody }}</pre>
      </CollapsiblePanel>
    </template>
  </div>
</template>

<script>
import CollapsiblePanel from './CollapsiblePanel.vue';
import DetailList from './DetailList.vue';
import DetailItem from './DetailItem.vue';
import JsonViewer from './JsonViewer.vue';

export default {
  name: 'ResponseDetails',
  components: {
    CollapsiblePanel,
    DetailList,
    DetailItem,
    JsonViewer
  },
  props: {
    session: {
      type: Object,
      required: true
    }
  },
  computed: {
    hasResponseCookies() {
      const cookies = this.session.responseCookies;
      if (!cookies) return false;
      if (typeof cookies === 'object' && Object.keys(cookies).length === 0) return false;
      return true;
    },
    hasResponseBody() {
      const body = this.session.responseBody;
      if (!body) return false;
      if (typeof body === 'string' && body.trim() === '') return false;
      return true;
    },
    isJsonResponseBody() {
      const body = this.session.responseBody;
      if (!body) return false;
      try {
        JSON.parse(body);
        return true;
      } catch (e) {
        return false;
      }
    },
    responseBodyAsJson() {
      if (!this.isJsonResponseBody) return null;
      try {
        return JSON.parse(this.session.responseBody);
      } catch (e) {
        return null;
      }
    }
  },
  methods: {
    formatError(error) {
      if (!error) return 'No error information';

      if (typeof error === 'object') {
        return JSON.stringify(error, null, 2);
      }

      return error;
    },
    formatHeaders(headers) {
      console.log('Formatting headers:', headers);
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
    getStatusClass(status) {
      if (status >= 200 && status < 300) return 'status-2xx';
      if (status >= 300 && status < 400) return 'status-3xx';
      if (status >= 400 && status < 500) return 'status-4xx';
      if (status >= 500) return 'status-5xx';
      return '';
    }
  }
}
</script>

<style scoped>
.detail-section {
  margin-bottom: var(--spacing-md);
  font-family: var(--font-family);
}

.section-title {
  font-size: var(--font-size-small);
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0 0 var(--spacing-sm) 0;
  padding-bottom: var(--spacing-xs);
  border-bottom: 1px solid var(--border-color);
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

.error-block {
  background-color: var(--input-background);
  color: var(--status-error);
  border: 1px solid var(--status-error);
}
</style>