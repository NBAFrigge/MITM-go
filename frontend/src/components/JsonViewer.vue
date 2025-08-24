<template>
  <div class="json-viewer">
    <div class="json-toolbar">
      <button @click="expandAll" class="toolbar-btn" title="Expand All">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
          <path d="M7 14l5-5 5 5z"/>
        </svg>
        Expand All
      </button>
      <button @click="collapseAll" class="toolbar-btn" title="Collapse All">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
          <path d="M7 10l5 5 5-5z"/>
        </svg>
        Collapse All
      </button>
      <button @click="copyJson" class="toolbar-btn" title="Copy JSON">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
          <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
        </svg>
        Copy
      </button>
    </div>
    <div class="json-content" ref="jsonContent">
      <JsonNode
          :data="json"
          :path="[]"
          :key="refreshKey"
          :global-expanded="globalExpanded"
          :expanded-state="expandedState"
          @toggle="onToggle"
      />
    </div>
  </div>
</template>

<script>
import JsonNode from './JsonNode.vue';
import { reactive, ref } from 'vue';

export default {
  name: 'JsonViewer',
  components: {
    JsonNode
  },
  props: {
    json: {
      type: [Object, Array],
      required: true
    }
  },
  setup(props) {
    const globalExpanded = ref(true);
    const expandedState = reactive({});
    const refreshKey = ref(0);

    const expandAll = () => {
      globalExpanded.value = true;
      Object.keys(expandedState).forEach(key => {
        delete expandedState[key];
      });
      refreshKey.value++;
    };

    const collapseAll = () => {
      globalExpanded.value = false;
      Object.keys(expandedState).forEach(key => {
        delete expandedState[key];
      });
      refreshKey.value++;
    };

    const copyJson = async () => {
      try {
        const jsonString = JSON.stringify(props.json, null, 2);
        await navigator.clipboard.writeText(jsonString);
        console.log('JSON copied to clipboard');
      } catch (err) {
        console.error('Failed to copy JSON:', err);
        const textArea = document.createElement('textarea');
        textArea.value = JSON.stringify(props.json, null, 2);
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
      }
    };

    const onToggle = (path) => {
      console.log('Toggle received for path:', JSON.stringify(path));
    };

    return {
      globalExpanded,
      expandedState,
      refreshKey,
      expandAll,
      collapseAll,
      copyJson,
      onToggle
    };
  }
};
</script>

<style scoped>
.json-viewer {
  background-color: var(--bg-color-dark);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  font-family: var(--font-family);
}

.json-toolbar {
  display: flex;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-sm);
  background-color: var(--bg-color-medium);
  border-bottom: 1px solid var(--border-color);
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  background: none;
  border: 1px solid var(--border-color);
  color: var(--text-color-secondary);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-small);
  cursor: pointer;
  transition: var(--transition-fast);
  font-family: inherit;
  font-weight: 500;
}

.toolbar-btn:hover {
  background-color: var(--hover-color);
  color: var(--text-color-primary);
  border-color: var(--accent-color-blue);
}

.json-content {
  padding: var(--spacing-sm);
  font-size: var(--font-size-small);
  line-height: 1.4;
  color: var(--text-color-primary);
  overflow-x: auto;
}
</style>