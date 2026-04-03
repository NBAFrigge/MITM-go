package sortedMap

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SortedMap struct {
	Entries map[string]interface{} `json:"entries"`
	Order   []string               `json:"order"`
}

func New() *SortedMap {
	return &SortedMap{
		Entries: make(map[string]interface{}),
		Order:   []string{},
	}
}

func (sm *SortedMap) Get(key string) (interface{}, bool) {
	value, exists := sm.Entries[key]
	if !exists {
		return nil, false
	}
	return value, true
}

func (sm *SortedMap) Put(key string, value interface{}) {
	if _, exists := sm.Entries[key]; !exists {
		sm.Order = append(sm.Order, key)
	}
	sm.Entries[key] = value
}

func (sm *SortedMap) Delete(key string) {
	delete(sm.Entries, key)

	for i, k := range sm.Order {
		if k == key {
			sm.Order[i] = sm.Order[len(sm.Order)-1]
			sm.Order = sm.Order[:len(sm.Order)-1]
			break
		}
	}
}

func (sm *SortedMap) Equal(other *SortedMap) bool {
	if len(sm.Entries) != len(other.Entries) || len(sm.Order) != len(other.Order) {
		return false
	}

	for key, value := range sm.Entries {
		if otherValue, exists := other.Entries[key]; !exists || value != otherValue {
			return false
		}
	}

	for i, key := range sm.Order {
		if key != other.Order[i] {
			return false
		}
	}

	return true
}

func (sm *SortedMap) Keys() []string {
	return sm.Order
}

func (sm *SortedMap) Len() int {
	return len(sm.Entries)
}

func (sm *SortedMap) OrderedValues() []interface{} {
	values := make([]interface{}, 0, len(sm.Entries))
	for _, key := range sm.Order {
		if value, exists := sm.Entries[key]; exists {
			values = append(values, value)
		}
	}
	return values
}

func (sm *SortedMap) String() string {
	var sb strings.Builder
	for _, key := range sm.Order {
		if value, exists := sm.Entries[key]; exists {
			sb.WriteString(key)
			sb.WriteString(" ")
			sb.WriteString(fmt.Sprintf("%v \n", value))
		}
	}
	return sb.String()
}

func (sm *SortedMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"order":   sm.Order,
		"entries": sm.Entries,
	})
}

func (sm *SortedMap) UnmarshalJSON(data []byte) error {
	var temp struct {
		Order   []string               `json:"order"`
		Entries map[string]interface{} `json:"entries"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	sm.Order = temp.Order
	sm.Entries = temp.Entries
	return nil
}
