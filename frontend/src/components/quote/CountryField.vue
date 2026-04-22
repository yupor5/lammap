<template>
  <div class="flex items-center justify-between p-2 rounded bg-gray-50 hover:bg-gray-100 transition-colors">
    <span class="text-sm text-gray-600 shrink-0 w-20">{{ label }}</span>
    <div class="flex-1 ml-2">
      <el-select
        v-if="isEditing"
        v-model="localValue"
        size="small"
        filterable
        clearable
        allow-create
        default-first-option
        :filter-method="filterCountry"
        @blur="handleConfirm"
        @change="handleConfirm"
        ref="selectRef"
        placeholder="输入国家/地区，支持首字母搜索"
        style="width: 100%"
      >
        <el-option
          v-for="opt in filteredOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>

      <div
        v-else
        class="flex items-center gap-1 cursor-pointer"
        @click="handleClickDisplay"
      >
        <span v-if="modelValue" class="text-sm text-gray-800">{{ modelValue }}</span>
        <el-tag v-else type="warning" size="small">未识别（点击选择/输入）</el-tag>
        <el-button text size="small" :icon="Edit" @click.stop="startEdit" class="ml-auto" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { Edit } from '@element-plus/icons-vue'

const props = defineProps<{
  label: string
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

type CountryOption = { value: string; label: string; keywords: string }

const baseOptions: CountryOption[] = [
  { value: 'China', label: 'China / 中国', keywords: 'cn china 中国 zhongguo' },
  { value: 'United States', label: 'United States / 美国', keywords: 'us usa united states america 美国 mei guo' },
  { value: 'Canada', label: 'Canada / 加拿大', keywords: 'ca canada 加拿大' },
  { value: 'Mexico', label: 'Mexico / 墨西哥', keywords: 'mx mexico 墨西哥' },
  { value: 'Brazil', label: 'Brazil / 巴西', keywords: 'br brazil 巴西' },
  { value: 'Argentina', label: 'Argentina / 阿根廷', keywords: 'ar argentina 阿根廷' },
  { value: 'Chile', label: 'Chile / 智利', keywords: 'cl chile 智利' },

  { value: 'United Kingdom', label: 'United Kingdom / 英国', keywords: 'uk united kingdom britain england 英国' },
  { value: 'Germany', label: 'Germany / 德国', keywords: 'de germany 德国' },
  { value: 'France', label: 'France / 法国', keywords: 'fr france 法国' },
  { value: 'Italy', label: 'Italy / 意大利', keywords: 'it italy 意大利' },
  { value: 'Spain', label: 'Spain / 西班牙', keywords: 'es spain 西班牙' },
  { value: 'Netherlands', label: 'Netherlands / 荷兰', keywords: 'nl netherlands holland 荷兰' },
  { value: 'Belgium', label: 'Belgium / 比利时', keywords: 'be belgium 比利时' },
  { value: 'Poland', label: 'Poland / 波兰', keywords: 'pl poland 波兰' },
  { value: 'Sweden', label: 'Sweden / 瑞典', keywords: 'se sweden 瑞典' },
  { value: 'Norway', label: 'Norway / 挪威', keywords: 'no norway 挪威' },
  { value: 'Denmark', label: 'Denmark / 丹麦', keywords: 'dk denmark 丹麦' },
  { value: 'Finland', label: 'Finland / 芬兰', keywords: 'fi finland 芬兰' },
  { value: 'Switzerland', label: 'Switzerland / 瑞士', keywords: 'ch switzerland 瑞士' },
  { value: 'Austria', label: 'Austria / 奥地利', keywords: 'at austria 奥地利' },
  { value: 'Ireland', label: 'Ireland / 爱尔兰', keywords: 'ie ireland 爱尔兰' },
  { value: 'Portugal', label: 'Portugal / 葡萄牙', keywords: 'pt portugal 葡萄牙' },
  { value: 'Czech Republic', label: 'Czech Republic / 捷克', keywords: 'cz czech 捷克' },
  { value: 'Greece', label: 'Greece / 希腊', keywords: 'gr greece 希腊' },
  { value: 'Turkey', label: 'Turkey / 土耳其', keywords: 'tr turkey 土耳其' },
  { value: 'Russia', label: 'Russia / 俄罗斯', keywords: 'ru russia 俄罗斯' },
  { value: 'Ukraine', label: 'Ukraine / 乌克兰', keywords: 'ua ukraine 乌克兰' },

  { value: 'Japan', label: 'Japan / 日本', keywords: 'jp japan 日本' },
  { value: 'South Korea', label: 'South Korea / 韩国', keywords: 'kr korea 韩国' },
  { value: 'India', label: 'India / 印度', keywords: 'in india 印度' },
  { value: 'Vietnam', label: 'Vietnam / 越南', keywords: 'vn vietnam 越南' },
  { value: 'Thailand', label: 'Thailand / 泰国', keywords: 'th thailand 泰国' },
  { value: 'Malaysia', label: 'Malaysia / 马来西亚', keywords: 'my malaysia 马来西亚' },
  { value: 'Singapore', label: 'Singapore / 新加坡', keywords: 'sg singapore 新加坡' },
  { value: 'Indonesia', label: 'Indonesia / 印度尼西亚', keywords: 'id indonesia 印尼 印度尼西亚' },
  { value: 'Philippines', label: 'Philippines / 菲律宾', keywords: 'ph philippines 菲律宾' },
  { value: 'Australia', label: 'Australia / 澳大利亚', keywords: 'au australia 澳大利亚' },
  { value: 'New Zealand', label: 'New Zealand / 新西兰', keywords: 'nz new zealand 新西兰' },

  { value: 'United Arab Emirates', label: 'UAE / 阿联酋', keywords: 'ae uae united arab emirates 阿联酋' },
  { value: 'Saudi Arabia', label: 'Saudi Arabia / 沙特', keywords: 'sa saudi 沙特' },
  { value: 'Israel', label: 'Israel / 以色列', keywords: 'il israel 以色列' },
  { value: 'South Africa', label: 'South Africa / 南非', keywords: 'za south africa 南非' },
  { value: 'Egypt', label: 'Egypt / 埃及', keywords: 'eg egypt 埃及' },
  { value: 'Nigeria', label: 'Nigeria / 尼日利亚', keywords: 'ng nigeria 尼日利亚' },
  { value: 'Kenya', label: 'Kenya / 肯尼亚', keywords: 'ke kenya 肯尼亚' },
]

const isEditing = ref(false)
const localValue = ref('')
const selectRef = ref()
const query = ref('')

const filteredOptions = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return baseOptions
  return baseOptions.filter((x) => {
    const hay = `${x.value} ${x.label} ${x.keywords}`.toLowerCase()
    // 支持首字母：输入 "u" 可匹配 United...
    return hay.includes(q)
  })
})

function startEdit() {
  localValue.value = String(props.modelValue || '')
  isEditing.value = true
  query.value = ''
  nextTick(() => selectRef.value?.focus?.())
}

defineExpose({ startEdit })

function handleClickDisplay() {
  if (!props.modelValue) startEdit()
}

function handleConfirm() {
  emit('update:modelValue', String(localValue.value || '').trim())
  isEditing.value = false
  query.value = ''
}

function filterCountry(q: string) {
  query.value = q
}
</script>

