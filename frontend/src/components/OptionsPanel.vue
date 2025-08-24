<template>
  <div class="options-panel">
    <div class="panel-header">
      <h3>Options</h3>
      <div class="status-indicator" :class="{ disabled: isRunning }">
        {{ isRunning ? 'Proxy Running - Options Locked' : 'Options Available' }}
      </div>
    </div>

    <div class="panel-content">
      <div class="option-section">
        <div class="section-header">
          <h4>Proxy Settings</h4>
          <p class="section-description">Configure proxy server parameters</p>
        </div>

        <div class="option-grid">
          <div class="option-item">
            <label for="port" class="option-label">
              <span class="label-text">Port</span>
              <span class="label-description">Server listening port (1024-65535)</span>
            </label>
            <input
                id="port"
                :value="proxyPort"
                @input="$emit('update-port', Number($event.target.value))"
                type="number"
                min="1024"
                max="65535"
                :disabled="isRunning"
                class="option-input"
                placeholder="8080"
            />
          </div>

          <div class="option-item">
            <label for="verbose" class="option-label">
              <span class="label-text">Verbose Logging</span>
              <span class="label-description">Enable detailed request/response logging</span>
            </label>
            <div class="toggle-container">
              <input
                  id="verbose"
                  :checked="verbose"
                  @change="$emit('update-verbose', $event.target.checked)"
                  type="checkbox"
                  :disabled="isRunning"
                  class="toggle-input"
              />
              <label for="verbose" class="toggle-slider"></label>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'OptionsPanel',
  props: {
    proxyPort: {
      type: Number,
      required: true
    },
    verbose: {
      type: Boolean,
      required: true
    },
    isRunning: {
      type: Boolean,
      required: true
    }
  },
  emits: ['update-port', 'update-verbose']
}
</script>

<style scoped>
.options-panel {
  padding: var(--spacing-md) var(--spacing-lg);
  background-color: var(--bg-color-dark);
  border-radius: 0;
  height: 100%;
  overflow-y: auto;
  font-family: var(--font-family);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--border-color);
}

.panel-header h3 {
  margin: 0;
  font-size: var(--font-size-normal);
  font-weight: 600;
  color: var(--text-color-primary);
}

.status-indicator {
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--font-size-small);
  letter-spacing: 0.3px;
  color: #ffffff;
  background-color: var(--status-success);
  transition: var(--transition-fast);
  border: 1px solid var(--status-success);
}

.status-indicator.disabled {
  background-color: var(--status-error);
  border-color: var(--status-error);
  color: white;
}

.panel-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.option-section {
  background-color: var(--bg-color-light);
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.section-header {
  margin-bottom: var(--spacing-md);
}

.section-header h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: var(--font-size-small);
  font-weight: 600;
  color: var(--text-color-primary);
}

.section-description {
  margin: 0;
  font-size: var(--font-size-small);
  color: var(--text-color-secondary);
}

.option-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-md);
}

.option-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-md);
}

.option-label {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.label-text {
  font-weight: 500;
  color: var(--text-color-primary);
  font-size: var(--font-size-small);
}

.label-description {
  font-size: var(--font-size-small);
  color: var(--text-color-secondary);
  line-height: 1.4;
}

.option-input {
  width: 80px;
  padding: var(--spacing-xs) var(--spacing-sm);
  border: 1px solid var(--input-border);
  border-radius: var(--radius-sm);
  background-color: var(--input-background);
  color: var(--text-color-primary);
  transition: var(--transition-fast);
  -moz-appearance: textfield;
  font-size: var(--font-size-small);
  font-family: inherit;
}

.option-input::-webkit-outer-spin-button,
.option-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.option-input:focus {
  border-color: var(--input-focus);
  outline: none;
}

.option-input:disabled {
  background-color: var(--bg-color-light);
  cursor: not-allowed;
  opacity: 0.7;
}

.toggle-container {
  position: relative;
  display: inline-block;
}

.toggle-input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: relative;
  display: inline-block;
  width: 32px;
  height: 16px;
  background-color: var(--text-color-secondary);
  border-radius: 16px;
  cursor: pointer;
  transition: var(--transition-fast);
  border: 1px solid var(--border-color);
}

.toggle-slider:before {
  content: "";
  position: absolute;
  height: 12px;
  width: 12px;
  left: 2px;
  top: 1px;
  background-color: white;
  border-radius: 50%;
  transition: var(--transition-fast);
}

.toggle-input:checked + .toggle-slider {
  background-color: var(--accent-color-blue);
  border-color: var(--accent-color-blue);
}

.toggle-input:checked + .toggle-slider:before {
  transform: translateX(16px);
}

.toggle-input:disabled + .toggle-slider {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .option-item {
    flex-direction: column;
    align-items: stretch;
  }

  .option-input {
    width: 100%;
  }

  .toggle-container {
    align-self: flex-start;
  }
}
</style>