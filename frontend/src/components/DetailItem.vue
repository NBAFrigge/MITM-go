<template>
  <div class="detail-item">
    <strong>{{ label }}:</strong>
    <slot></slot>
  </div>
</template>

<script>
export default {
  name: 'DetailItem',
  props: {
    label: {
      type: String,
      required: true
    }
  }
}
</script>

<style scoped>
.detail-item {
  display: flex;
  padding: var(--spacing-sm) var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-color-dark);
  font-family: var(--font-family);
  font-size: var(--font-size-small);
  min-height: 24px;
  align-items: center;
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-item:nth-child(even) {
  background-color: var(--bg-color-medium);
}

.detail-item:hover {
  background-color: var(--hover-color);
}

.detail-item strong {
  color: var(--text-color-secondary);
  min-width: 120px;
  font-weight: 600;
  font-size: var(--font-size-small);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
</style>