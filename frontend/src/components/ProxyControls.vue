<template>
  <div class="control-panel">
    <div class="panel-container">
      <div class="panel-info">
        <h3>Proxy Control</h3>
        <div class="status" :class="{ active: isRunning }">
          {{ statusText }}
        </div>
      </div>
      <div class="panel-actions">
        <button
            @click="toggleProxy"
            :class="['btn', isRunning ? 'btn-stop' : 'btn-start']"
        >
          {{ isRunning ? 'Stop Proxy' : 'Start Proxy' }}
        </button>
        <button @click="$emit('clear-sessions')" class="btn btn-trash">
          <svg class="icon-trash" viewBox="0 0 24 24">
            <polyline points="3,6 5,6 21,6" />
            <path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2v2" />
            <line x1="10" y1="11" x2="10" y2="17" />
            <line x1="14" y1="11" x2="14" y2="17" />
          </svg>
          Clear Sessions
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ProxyControls',
  props: {
    isRunning: {
      type: Boolean,
      required: true
    },
    statusText: {
      type: String,
      required: true
    }
  },
  emits: ['start-proxy', 'stop-proxy', 'clear-sessions'],
  methods: {
    toggleProxy() {
      if (this.isRunning) {
        this.$emit('stop-proxy');
      } else {
        this.$emit('start-proxy');
      }
    }
  }
};
</script>

<style scoped>
.control-panel {
  padding: var(--spacing-sm) var(--spacing-lg);
  background-color: var(--bg-color-medium);
  border-bottom: 1px solid var(--border-color);
  font-family: var(--font-family);
}

.panel-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-lg);
}

.panel-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex: 1;
}

.panel-info h3 {
  margin: 0;
  font-size: var(--font-size-normal);
  font-weight: 600;
  color: var(--text-color-primary);
}

.status {
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--font-size-small);
  letter-spacing: 0.3px;
  color: white;
  background-color: var(--text-color-secondary);
  transition: var(--transition-fast);
  border: 1px solid var(--text-color-secondary);
}

.status.active {
  background-color: var(--status-success);
  border-color: var(--status-success);
}

.panel-actions {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.btn {
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--font-size-small);
  cursor: pointer;
  transition: var(--transition-fast);
  border: 1px solid;
  color: white;
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-family: inherit;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn:hover:not(:disabled) {
  opacity: 0.8;
}

.btn-start {
  background-color: var(--status-success);
  border-color: var(--status-success);
}

.btn-stop {
  background-color: var(--status-error);
  border-color: var(--status-error);
}

.btn-trash {
  background-color: var(--status-error);
  border-color: var(--status-error);
}

.icon-trash {
  width: 12px;
  height: 12px;
  stroke: white;
  stroke-width: 2;
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
}

@media (max-width: 768px) {
  .control-panel {
    padding: 6px 12px;
  }

  .panel-container {
    flex-direction: column;
    gap: 8px;
    align-items: stretch;
  }

  .panel-info {
    justify-content: center;
    gap: 8px;
  }

  .panel-info h3 {
    font-size: 12px;
  }

  .panel-actions {
    justify-content: center;
    gap: 8px;
  }

  .btn {
    flex: 1;
    min-width: 80px;
  }
}

@media (max-width: 480px) {
  .panel-info {
    flex-direction: column;
    gap: 4px;
  }

  .panel-actions {
    flex-direction: column;
    gap: 4px;
  }

  .btn {
    width: 100%;
    min-width: auto;
  }
}
</style>