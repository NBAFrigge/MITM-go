<template>
  <div class="detail-section">
    <h3 class="section-title">Request</h3>

    <DetailList>
      <DetailItem label="Method">
        <span>{{ session.method }}</span>
      </DetailItem>
      <DetailItem label="URL">
        <span>{{ session.url }}</span>
      </DetailItem>
    </DetailList>

    <CollapsiblePanel title="Headers" :nested="true" :initial-expanded="true">
      <pre class="code-block">{{ formatHeaders(session.requestHeaders) }}</pre>
    </CollapsiblePanel>

    <CollapsiblePanel
        v-if="hasCookies"
        title="Cookies"
        :nested="true"
        :initial-expanded="true"
    >
      <pre class="code-block">{{ formatCookies(session.requestCookies) }}</pre>
    </CollapsiblePanel>

    <CollapsiblePanel
        v-if="hasBody"
        title="Body"
        :nested="true"
        :initial-expanded="true"
    >
      <JsonViewer v-if="isJsonBody" :json="bodyAsJson" />
      <pre v-else class="code-block">{{ session.requestBody }}</pre>
    </CollapsiblePanel>
  </div>
</template>

<script>
import CollapsiblePanel from './CollapsiblePanel.vue';
import DetailList from './DetailList.vue';
import DetailItem from './DetailItem.vue';
import JsonViewer from './JsonViewer.vue';

export default {
  name: 'RequestDetails',
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
    hasCookies() {
      const cookies = this.session.requestCookies;
      if (!cookies) return false;
      if (typeof cookies === 'object' && Object.keys(cookies).length === 0) return false;
      return true;
    },
    hasBody() {
      const body = this.session.requestBody;
      if (!body) return false;
      if (typeof body === 'string' && body.trim() === '') return false;
      return true;
    },
    isJsonBody() {
      const body = this.session.requestBody;
      if (!body) return false;
      try {
        JSON.parse(body);
        return true;
      } catch (e) {
        return false;
      }
    },
    bodyAsJson() {
      if (!this.isJsonBody) return null;
      try {
        return JSON.parse(this.session.requestBody);
      } catch (e) {
        return null;
      }
    }
  },
  methods: {
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