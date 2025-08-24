<template>
  <nav class="app-navbar">
    <div class="navbar-container">
      <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="$emit('tab-changed', tab.id)"
          :class="['nav-tab', { 'active': activeTab === tab.id }]"
      >
        <span class="tab-icon" v-html="tab.icon"></span>
        <span class="tab-label">{{ tab.label }}</span>
        <div v-if="activeTab === tab.id" class="active-indicator"></div>
      </button>
    </div>
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
  justify-content: space-between;
  align-items: center;
  background-color: var(--bg-color-medium);
  border-bottom: 1px solid var(--border-color);
  padding: 0 var(--spacing-md);
  box-shadow: var(--shadow-light);
  min-height: 40px;
  flex-shrink: 0;
}

.navbar-container {
  display: flex;
  align-items: center;
  gap: 0;
}

.nav-tab {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-lg);
  background: none;
  border: none;
  color: var(--text-color-secondary);
  font-size: var(--font-size-normal);
  font-weight: 500;
  font-family: var(--font-family-ui);
  cursor: pointer;
  transition: var(--transition-fast);
  position: relative;
  border-radius: var(--radius-md) var(--radius-md) 0 0;
  margin: 0 1px;
  min-height: 40px;
  user-select: none;
}

.nav-tab:hover {
  background-color: var(--hover-color);
  color: var(--text-color-primary);
  transform: translateY(-1px);
}

.nav-tab:active {
  transform: translateY(0);
}

.nav-tab.active {
  color: var(--text-color-primary);
  background-color: var(--bg-color-light);
  font-weight: 600;
  box-shadow: inset 0 -2px 0 var(--accent-color-blue);
}

.nav-tab.active:hover {
  transform: none;
}

.tab-icon {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.tab-icon :deep(svg) {
  width: 16px;
  height: 16px;
  stroke: currentColor;
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
  transition: var(--transition-fast);
}

.nav-tab:hover .tab-icon :deep(svg) {
  stroke-width: 2.5;
}

.nav-tab.active .tab-icon :deep(svg) {
  stroke: var(--accent-color-blue);
  stroke-width: 2.5;
}

.tab-label {
  font-size: var(--font-size-normal);
  white-space: nowrap;
  transition: var(--transition-fast);
}

.active-indicator {
  position: absolute;
  bottom: -1px;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--accent-color-blue), var(--number-color));
  border-radius: 1px;
}

.navbar-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.nav-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: none;
  border: none;
  color: var(--text-color-secondary);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: var(--transition-fast);
}

.nav-action-btn:hover {
  background-color: var(--hover-color);
  color: var(--text-color-primary);
}

.nav-action-btn svg {
  width: 14px;
  height: 14px;
  stroke: currentColor;
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

/* Focus states for accessibility */
.nav-tab:focus {
  outline: 2px solid var(--accent-color-blue);
  outline-offset: -2px;
  z-index: 1;
}

.nav-action-btn:focus {
  outline: 2px solid var(--accent-color-blue);
  outline-offset: 2px;
}

/* Responsive Design */
@media (max-width: 1024px) {
  .app-navbar {
    padding: 0 var(--spacing-sm);
  }

  .nav-tab {
    padding: var(--spacing-sm) var(--spacing-md);
    gap: var(--spacing-xs);
  }

  .tab-icon {
    width: 14px;
    height: 14px;
  }

  .tab-icon :deep(svg) {
    width: 14px;
    height: 14px;
  }
}

@media (max-width: 768px) {
  .app-navbar {
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
    padding: 0 var(--spacing-sm);
  }

  .app-navbar::-webkit-scrollbar {
    display: none;
  }

  .navbar-container {
    min-width: min-content;
  }

  .nav-tab {
    white-space: nowrap;
    min-width: fit-content;
    padding: var(--spacing-sm);
    gap: var(--spacing-xs);
  }

  .tab-label {
    font-size: var(--font-size-small);
  }

  .navbar-actions {
    flex-shrink: 0;
    margin-left: var(--spacing-sm);
  }
}

@media (max-width: 480px) {
  .nav-tab {
    padding: var(--spacing-xs) var(--spacing-sm);
  }

  .tab-icon {
    width: 12px;
    height: 12px;
  }

  .tab-icon :deep(svg) {
    width: 12px;
    height: 12px;
  }

  .tab-label {
    font-size: 11px;
  }

  .nav-action-btn {
    width: 24px;
    height: 24px;
  }

  .nav-action-btn svg {
    width: 12px;
    height: 12px;
  }
}

/* Animation for tab switching */
@keyframes tabSlide {
  from {
    opacity: 0;
    transform: translateY(2px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.nav-tab.active {
  animation: tabSlide var(--transition-fast);
}

/* High contrast mode */
@media (prefers-contrast: high) {
  .nav-tab.active {
    box-shadow: inset 0 -3px 0 var(--accent-color-blue);
  }

  .active-indicator {
    height: 3px;
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .nav-tab,
  .nav-action-btn,
  .tab-icon :deep(svg),
  .tab-label,
  .active-indicator {
    transition: none;
  }

  .nav-tab.active {
    animation: none;
  }

  .nav-tab:hover {
    transform: none;
  }
}
</style>