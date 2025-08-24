<template>
  <div class="json-node">
    <div v-if="isObject || isArray" class="json-complex">
      <div class="json-line" @click="toggleExpanded">
        <span class="json-toggle" :class="{ expanded: isExpanded }">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
            <path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z"/>
          </svg>
        </span>
        <span v-if="keyName !== null" class="json-key">"{{ keyName }}":</span>
        <span class="json-bracket">{{ isArray ? '[' : '{' }}</span>
        <span v-if="!isExpanded" class="json-collapsed">
          {{ isArray ? `${data.length} items` : `${Object.keys(data).length} keys` }}
        </span>
        <span v-if="!isExpanded" class="json-bracket">{{ isArray ? ']' : '}' }}</span>
      </div>

      <div v-if="isExpanded" class="json-children">
        <template v-for="(value, key, index) in data" :key="key">
          <div class="json-child-line">
            <JsonNode
                :data="value"
                :key-name="key"
                :path="[...path, key]"
                :global-expanded="globalExpanded"
                :expanded-state="expandedState"
                @toggle="$emit('toggle', $event)"
            />
            <span v-if="index < objectKeysLength - 1" class="json-comma">,</span>
          </div>
        </template>
      </div>

      <div v-if="isExpanded" class="json-line">
        <span class="json-bracket">{{ isArray ? ']' : '}' }}</span>
      </div>
    </div>

    <div v-else class="json-primitive-line">
      <span v-if="keyName !== null" class="json-key">"{{ keyName }}":</span>
      <span class="json-value" :class="getValueClass(data)">{{ formatValue(data) }}</span>
    </div>
  </div>
</template>

<script>
export default {
  name: 'JsonNode',
  props: {
    data: {
      required: true
    },
    keyName: {
      type: [String, Number],
      default: null
    },
    path: {
      type: Array,
      default: () => []
    },
    globalExpanded: {
      type: Boolean,
      default: true
    },
    expandedState: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ['toggle'],
  computed: {
    isObject() {
      return this.data !== null && typeof this.data === 'object' && !Array.isArray(this.data);
    },
    isArray() {
      return Array.isArray(this.data);
    },
    pathKey() {
      return JSON.stringify(this.path);
    },
    isExpanded() {
      // Controlla se c'è uno stato esplicito per questo nodo
      const state = this.expandedState[this.pathKey];
      if (state !== undefined) {
        return state;
      }
      // Altrimenti usa lo stato globale
      return this.globalExpanded;
    },
    objectKeysLength() {
      if (this.isArray) return this.data.length;
      if (this.isObject) return Object.keys(this.data).length;
      return 0;
    }
  },
  methods: {
    toggleExpanded() {
      console.log('Toggling expansion for path:', this.pathKey);
      const wasExpanded = this.isExpanded;

      // Assegnazione diretta - Vue 3 è reattivo per oggetti
      this.expandedState[this.pathKey] = !wasExpanded;

      console.log(`Node ${this.pathKey} is now ${!wasExpanded ? 'expanded' : 'collapsed'}`);
      this.$emit('toggle', this.path);
    },
    formatValue(value) {
      if (value === null) return 'null';
      if (typeof value === 'string') return `"${value}"`;
      if (typeof value === 'boolean') return value.toString();
      if (typeof value === 'number') return value.toString();
      return String(value);
    },
    getValueClass(value) {
      if (value === null) return 'json-null';
      if (typeof value === 'string') return 'json-string';
      if (typeof value === 'boolean') return 'json-boolean';
      if (typeof value === 'number') return 'json-number';
      return 'json-unknown';
    }
  }
};
</script>

<style scoped>
.json-node {
  font-family: var(--font-family);
}

.json-line {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 1px 0;
  border-radius: var(--radius-sm);
  transition: var(--transition-fast);
  font-size: var(--font-size-small);
}

.json-line:hover {
  background-color: var(--hover-color);
}

.json-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 12px;
  height: 12px;
  margin-right: var(--spacing-xs);
  transition: var(--transition-fast);
  color: var(--text-color-secondary);
  cursor: pointer;
}

.json-toggle.expanded {
  transform: rotate(90deg);
}

.json-children {
  margin-left: var(--spacing-lg);
  border-left: 1px solid var(--border-color);
  padding-left: var(--spacing-sm);
}

.json-child-line {
  display: flex;
  align-items: baseline;
}

.json-primitive-line {
  display: inline-flex;
  align-items: baseline;
  padding: 1px 0;
  font-size: 12px;
}

.json-key {
  color: var(--field-color);
  font-weight: 500;
  margin-right: var(--spacing-sm);
}

.json-bracket {
  color: var(--text-color-secondary);
  font-weight: bold;
}

.json-collapsed {
  color: var(--text-color-secondary);
  font-style: italic;
  margin: 0 var(--spacing-sm);
  font-size: var(--font-size-small);
}

.json-comma {
  color: var(--text-color-secondary);
  margin-left: 0;
}

.json-value {
  margin-left: var(--spacing-sm);
}

.json-string {
  color: var(--string-color);
}

.json-number {
  color: var(--number-color);
}

.json-boolean {
  color: var(--keyword-color);
}

.json-null {
  color: var(--text-color-secondary);
  font-style: italic;
}

.json-unknown {
  color: var(--text-color-primary);
}
</style>