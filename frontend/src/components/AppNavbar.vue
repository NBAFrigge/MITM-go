<template>
  <nav class="app-navbar">
    <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="$emit('tab-changed', tab.id)"
        :class="['nav-tab', { 'active': activeTab === tab.id }]"
    >
      <span class="tab-icon" v-html="tab.icon"></span>
      {{ tab.label }}
    </button>
  </nav>
</template>

<script>
export default {
  name: 'AppNavbar',
  props: {
    tabs: {
      type: Array,
      required: true
    },
    activeTab: {
      type: String,
      required: true
    }
  },
  emits: ['tab-changed']
}
</script>

<style scoped>
.app-navbar {
  display: flex;
  background-color: var(--bg-color-medium);
  border-bottom: 2px solid var(--border-color);
  padding: 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.nav-tab {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 1.5rem;
  background: none;
  border: none;
  color: var(--text-color-secondary);
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border-bottom: 3px solid transparent;
  position: relative;
}

.nav-tab:hover {
  background-color: var(--hover-color);
  color: var(--text-color-primary);
}

.nav-tab.active {
  color: var(--accent-color-blue);
  background-color: var(--hover-color);
  border-bottom-color: var(--accent-color-blue);
}

.tab-icon {
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tab-icon :deep(svg) {
  width: 100%;
  height: 100%;
  stroke: currentColor;
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

@media (max-width: 768px) {
  .app-navbar {
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }

  .app-navbar::-webkit-scrollbar {
    display: none;
  }

  .nav-tab {
    white-space: nowrap;
    min-width: fit-content;
  }
}
</style>