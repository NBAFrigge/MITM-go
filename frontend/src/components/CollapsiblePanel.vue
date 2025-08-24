<template>
  <div class="collapsible-section" :class="{ 'nested': nested, 'expanded': expanded }">
    <div class="collapsible-header" @click="toggleExpanded">
      <div class="header-content">
        <div class="expand-icon" :class="{ 'expanded': expanded }">
          <svg viewBox="0 0 24 24">
            <polyline points="9,18 15,12 9,6"></polyline>
          </svg>
        </div>
        <div class="title">{{ title }}</div>
        <div v-if="badge" class="header-badge">{{ badge }}</div>
      </div>
      <div class="header-actions">
        <slot name="header-actions"></slot>
      </div>
    </div>

    <div class="collapsible-content" :class="{ 'expanded': expanded }" ref="content">
      <div class="content-wrapper">
        <slot></slot>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'CollapsiblePanel',
  props: {
    title: {
      type: String,
      required: true
    },
    nested: {
      type: Boolean,
      default: false
    },
    initialExpanded: {
      type: Boolean,
      default: true
    },
    badge: {
      type: [String, Number],
      default: null
    }
  },
  data() {
    return {
      expanded: this.initialExpanded
    }
  },
  methods: {
    toggleExpanded() {
      this.expanded = !this.expanded
      this.$emit('toggle', this.expanded)
    },

    expand() {
      this.expanded = true
      this.$emit('toggle', true)
    },

    collapse() {
      this.expanded = false
      this.$emit('toggle', false)
    }
  },
  emits: ['toggle']
}
</script>

<style scoped>
.collapsible-section {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  background-color: var(--bg-color-medium);
  transition: var(--transition-normal);
  margin-bottom: var(--spacing-sm);
}

.collapsible-section:hover {
  border-color: var(--accent-color-blue);
  box-shadow: var(--shadow-light);
}

.collapsible-section.nested {
  margin: var(--spacing-md) 0;
  background-color: var(--bg-color-light);
  border-color: var(--border-color);
}

.collapsible-section.nested:hover {
  border-color: var(--text-color-secondary);
}

.collapsible-header {
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--bg-color-light);
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  user-select: none;
  transition: var(--transition-fast);
  min-height: 40px;
}

.collapsible-header:hover {
  background-color: var(--hover-color);
}

.collapsible-header:active {
  background-color: var(--bg-color-dark);
}

.collapsible-section.nested .collapsible-header {
  background-color: var(--bg-color-medium);
  padding: var(--spacing-xs) var(--spacing-sm);
  min-height: 32px;
}

.collapsible-section.nested .collapsible-header:hover {
  background-color: var(--bg-color-light);
}

.header-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex: 1;
}

.expand-icon {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition-fast);
  color: var(--text-color-secondary);
  flex-shrink: 0;
}

.expand-icon.expanded {
  transform: rotate(90deg);
  color: var(--accent-color-blue);
}

.expand-icon svg {
  width: 12px;
  height: 12px;
  stroke: currentColor;
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.collapsible-section.nested .expand-icon {
  width: 14px;
  height: 14px;
}

.collapsible-section.nested .expand-icon svg {
  width: 10px;
  height: 10px;
}

.title {
  font-size: var(--font-size-medium);
  font-weight: 600;
  color: var(--text-color-primary);
  font-family: var(--font-family-ui);
  line-height: 1.2;
  flex: 1;
}

.collapsible-section.nested .title {
  font-size: var(--font-size-normal);
  font-weight: 500;
  color: var(--text-color-secondary);
}

.header-badge {
  background-color: var(--accent-color-blue);
  color: white;
  font-size: var(--font-size-small);
  font-weight: 600;
  font-family: var(--font-family);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  min-width: 20px;
  text-align: center;
  line-height: 1;
}

.collapsible-section.nested .header-badge {
  font-size: 10px;
  padding: 2px var(--spacing-xs);
  min-width: 16px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.collapsible-content {
  max-height: 0;
  overflow: hidden;
  transition: max-height var(--transition-normal) ease-out;
  background-color: var(--bg-color-dark);
}

.collapsible-content.expanded {
  max-height: 2000px;
  transition: max-height var(--transition-normal) ease-in;
}

.content-wrapper {
  padding: var(--spacing-md);
  border-top: 1px solid var(--border-color);
}

.collapsible-section.nested .content-wrapper {
  padding: var(--spacing-sm);
  background-color: var(--bg-color-dark);
}

/* Focus states for accessibility */
.collapsible-header:focus {
  outline: 2px solid var(--accent-color-blue);
  outline-offset: -2px;
  z-index: 1;
}

/* Animation improvements */
@keyframes expandGlow {
  0% {
    box-shadow: 0 0 0 0 rgba(104, 151, 187, 0.4);
  }
  70% {
    box-shadow: 0 0 0 4px rgba(104, 151, 187, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(104, 151, 187, 0);
  }
}

.collapsible-section.expanded {
  animation: expandGlow 0.6s ease-out;
}

/* States */
.collapsible-section:has(.content-wrapper:empty) .header-badge {
  background-color: var(--text-color-secondary);
}

.collapsible-section.expanded:has(.content-wrapper:empty) .header-badge {
  background-color: var(--status-warning);
  color: var(--bg-color-dark);
}

/* Responsive Design */
@media (max-width: 768px) {
  .collapsible-header {
    padding: var(--spacing-xs) var(--spacing-sm);
    min-height: 36px;
  }

  .header-content {
    gap: var(--spacing-xs);
  }

  .title {
    font-size: var(--font-size-normal);
  }

  .collapsible-section.nested .title {
    font-size: var(--font-size-small);
  }

  .content-wrapper {
    padding: var(--spacing-sm);
  }

  .collapsible-section.nested .content-wrapper {
    padding: var(--spacing-xs);
  }
}

@media (max-width: 480px) {
  .collapsible-header {
    padding: var(--spacing-xs);
    min-height: 32px;
  }

  .title {
    font-size: var(--font-size-small);
  }

  .header-badge {
    font-size: 10px;
    padding: 2px var(--spacing-xs);
  }

  .expand-icon {
    width: 14px;
    height: 14px;
  }

  .expand-icon svg {
    width: 10px;
    height: 10px;
  }
}

/* High contrast mode */
@media (prefers-contrast: high) {
  .collapsible-section {
    border-width: 2px;
  }

  .expand-icon.expanded {
    color: var(--accent-color-blue);
    font-weight: bold;
  }

  .header-badge {
    border: 1px solid var(--accent-color-blue);
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .collapsible-content,
  .expand-icon,
  .collapsible-section {
    transition: none;
  }

  .collapsible-section.expanded {
    animation: none;
  }
}

/* Print styles */
@media print {
  .collapsible-section {
    border: 1px solid #000;
    break-inside: avoid;
  }

  .collapsible-content {
    max-height: none !important;
    overflow: visible !important;
  }

  .expand-icon {
    display: none;
  }
}
</style>