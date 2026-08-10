<script setup lang="ts">
import { computed } from 'vue'
import {
  ensureTimezoneOption,
  groupTimezones,
  listTimezones,
  timezoneLabel,
} from '@/utils/timezones'

const props = withDefaults(
  defineProps<{
    modelValue: string
    id?: string
    required?: boolean
  }>(),
  {
    id: undefined,
    required: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const groups = computed(() => {
  const zones = ensureTimezoneOption(listTimezones(), props.modelValue)
  return groupTimezones(zones)
})

function onChange(event: Event) {
  const target = event.target as HTMLSelectElement
  emit('update:modelValue', target.value)
}
</script>

<template>
  <select
    :id="id"
    class="form-select"
    :required="required"
    :value="modelValue"
    @change="onChange"
  >
    <optgroup v-for="group in groups" :key="group.label" :label="group.label">
      <option v-for="zone in group.zones" :key="zone" :value="zone">
        {{ timezoneLabel(zone) }}
      </option>
    </optgroup>
  </select>
</template>
