<template>
  <div class="detail-items">
    <slot></slot>
  </div>
</template>

<script>
export default {
  name: 'DetailList'
}
</script>

<style scoped>
.detail-items {
  margin-bottom: var(--spacing-sm);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  background-color: var(--bg-color-dark);
}
</style>