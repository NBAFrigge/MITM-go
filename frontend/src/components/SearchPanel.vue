<template>
  <div class="search-panel">
    <div class="search-controls">
      <!-- Main search row - always visible -->
      <div class="main-search-row">
        <div class="url-search-group">
          <input
              v-model="searchCriteria.URL"
              @keyup.enter="performSearch"
              placeholder="Search URL..."
              class="search-input url-input"
          />
        </div>
        <div class="main-actions">
          <button @click="toggleAdvanced" class="btn btn-icon" :class="{ active: showAdvanced }">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M3 4H21V6H20V18C20 19.1 19.1 20 18 20H6C4.9 20 4 19.1 4 18V6H3V4ZM6 6V18H18V6H6ZM8 8H16V10H8V8ZM8 12H14V14H8V12Z" fill="currentColor"/>
              <circle cx="18" cy="8" r="2" fill="currentColor"/>
              <circle cx="18" cy="16" r="2" fill="currentColor"/>
            </svg>
          </button>
          <button @click="performSearch" class="btn btn-primary">
            Search
          </button>
        </div>
      </div>

      <!-- Advanced search options - collapsible -->
      <transition name="slide" mode="out-in">
        <div v-if="showAdvanced" class="advanced-search">
          <div class="search-row">
            <div class="search-group">
              <input
                  v-model="searchCriteria.HeadersKey"
                  @keyup.enter="performSearch"
                  placeholder="Header key..."
                  class="search-input search-input-half"
              />
              <input
                  v-model="searchCriteria.HeadersVal"
                  @keyup.enter="performSearch"
                  placeholder="Header value..."
                  class="search-input search-input-half"
              />
            </div>
            <div class="search-group">
              <input
                  v-model="searchCriteria.Body"
                  @keyup.enter="performSearch"
                  placeholder="Search in body..."
                  class="search-input"
              />
            </div>
          </div>
          <div class="search-row">
            <div class="search-group">
              <input
                  v-model="searchCriteria.CookiesKey"
                  @keyup.enter="performSearch"
                  placeholder="Cookie key..."
                  class="search-input search-input-half"
              />
              <input
                  v-model="searchCriteria.CookiesVal"
                  @keyup.enter="performSearch"
                  placeholder="Cookie value..."
                  class="search-input search-input-half"
              />
            </div>
            <div class="advanced-actions">
              <button @click="clearSearch" class="btn btn-secondary">
                Clear all
              </button>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script>
import { SearchSessions } from '../../wailsjs/go/main/App'

export default {
  name: 'SearchPanel',
  data() {
    return {
      showAdvanced: false,
      searchCriteria: {
        URL: '',
        HeadersKey: '',
        HeadersVal: '',
        Body: '',
        CookiesKey: '',
        CookiesVal: ''
      }
    }
  },
  methods: {
    toggleAdvanced() {
      this.showAdvanced = !this.showAdvanced
    },
    async performSearch() {
      try {
        const hasSearchTerms = Object.values(this.searchCriteria).some(value => value.trim() !== '')
        if (!hasSearchTerms) {
          this.clearSearch()
          return
        }
        const results = await SearchSessions(this.searchCriteria)
        this.$emit('search-results', results)
      } catch (error) {
        console.error('Error performing search:', error)
      }
    },
    clearSearch() {
      this.searchCriteria = {
        URL: '',
        HeadersKey: '',
        HeadersVal: '',
        Body: '',
        CookiesKey: '',
        CookiesVal: ''
      }
      this.$emit('search-cleared')
    }
  }
}
</script>

<style scoped>
.search-panel {
  background-color: var(--bg-color-light);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin-bottom: var(--spacing-md);
  font-family: var(--font-family);
}

.search-controls {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.main-search-row {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.url-search-group {
  flex: 1;
  padding-right: var(--spacing-md);
}

.url-input {
  width: 100%;
}

.main-actions {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.advanced-search {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  padding-top: var(--spacing-sm);
  border-top: 1px solid var(--border-color);
}

.slide-enter-active,
.slide-leave-active {
  transition: var(--transition-fast);
  overflow: hidden;
}

.slide-enter-from {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  border-top-width: 0;
}

.slide-enter-to {
  opacity: 1;
  max-height: 120px;
  padding-top: var(--spacing-sm);
  border-top-width: 1px;
}

.slide-leave-from {
  opacity: 1;
  max-height: 120px;
  padding-top: var(--spacing-sm);
  border-top-width: 1px;
}

.slide-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  border-top-width: 0;
}

.search-row {
  display: flex;
  gap: var(--spacing-sm);
  align-items: flex-end;
}

.search-group {
  flex: 1;
  display: flex;
  gap: var(--spacing-sm);
}

.advanced-actions {
  display: flex;
  align-items: flex-end;
}

.search-input {
  flex: 1;
  padding: var(--spacing-sm) var(--spacing-sm);
  border: 1px solid var(--input-border);
  border-radius: var(--radius-sm);
  background-color: var(--input-background);
  color: var(--text-color-primary);
  font-size: var(--font-size-small);
  transition: var(--transition-fast);
  min-width: 0;
  font-family: inherit;
}

.search-input:focus {
  outline: none;
  border-color: var(--input-focus);
}

.search-input::placeholder {
  color: var(--text-color-secondary);
}

.search-input-half {
  flex: 0.5;
}

.btn {
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--input-border);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-small);
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-xs);
  white-space: nowrap;
  font-family: inherit;
}
.btn-primary {
  background-color: var(--button-primary);
  color: var(--text-color-primary);
  border-color: var(--button-primary);
}

.btn-primary:hover {
  background-color: var(--hover-color);
  border-color: var(--accent-color-blue);
}

.btn-secondary {
  background-color: var(--button-secondary);
  color: var(--text-color-primary);
}

.btn-secondary:hover {
  background-color: var(--hover-color);
}

.btn-icon {
  background-color: var(--button-secondary);
  color: var(--text-color-secondary);
  padding: var(--spacing-sm);
  min-width: 24px;
  transition: var(--transition-fast);
}

.btn-icon:hover {
  background-color: var(--hover-color);
  color: var(--text-color-primary);
}

.btn-icon.active {
  background-color: var(--accent-color-blue);
  color: var(--text-color-primary);
  border-color: var(--accent-color-blue);
}

.btn-icon svg {
  transition: var(--transition-fast);
}

.btn-icon.active svg {
  transform: rotate(180deg);
}

@media (max-width: 1024px) {
  .search-row {
    flex-direction: column;
    gap: 6px;
    align-items: stretch;
  }

  .search-group {
    flex-direction: column;
  }

  .search-input-half {
    flex: 1;
  }

  .advanced-actions {
    align-items: stretch;
  }
}

@media (max-width: 768px) {
  .search-panel {
    padding: 8px;
    margin-bottom: 8px;
  }

  .main-search-row {
    flex-direction: column;
    gap: 6px;
  }

  .main-actions {
    width: 100%;
    justify-content: space-between;
  }

  .btn {
    flex: 1;
    justify-content: center;
  }

  .btn-icon {
    flex: 0 0 auto;
    min-width: 32px;
  }
}

@media (max-width: 480px) {
  .search-input {
    font-size: 16px;
  }

  .main-actions {
    flex-direction: row;
    gap: 6px;
  }
}
</style>