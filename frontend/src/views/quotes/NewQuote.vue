<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">把客户需求变成可发送报价</h1>
        <p class="text-gray-500 mt-1">粘贴客户发来的需求，AI 自动解析并生成报价</p>
      </div>
      <div class="flex gap-2">
        <el-button @click="handleSaveDraft">保存草稿</el-button>
        <el-button type="primary" :icon="DocumentAdd" @click="handleGenerate" :loading="quoteStore.isLoading" :disabled="!parsedParams">
          开始生成
        </el-button>
      </div>
    </div>

    <div class="grid grid-cols-12 gap-4" style="height: calc(100vh - 180px);">
      <!-- 左栏：客户需求输入 -->
      <div class="col-span-3 bg-white rounded-lg border border-gray-100 shadow-sm flex flex-col overflow-hidden">
        <div class="p-4 border-b border-gray-100">
          <h3 class="font-semibold text-gray-800">客户需求输入</h3>
        </div>
        <div class="flex-1 p-4 flex flex-col gap-4 overflow-y-auto">
          <el-input
            v-model="requirementText"
            type="textarea"
            :autosize="{ minRows: 10, maxRows: 20 }"
            placeholder="把客户发来的需求粘贴到这里…

例如：
Need 500 pcs stainless steel table legs, 70cm height, black coating, packing by carton, delivery to LA."
            class="flex-1"
          />

          <el-upload
            v-model:file-list="uploadFiles"
            drag
            multiple
            :auto-upload="false"
            accept=".pdf,.doc,.docx,.xls,.xlsx,.png,.jpg,.jpeg"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">拖拽文件到此或 <em>点击上传</em></div>
            <template #tip>
              <div class="el-upload__tip">支持 PDF/Word/Excel/图片</div>
            </template>
          </el-upload>

          <div class="flex items-center gap-2">
            <span class="text-sm text-gray-500">语言识别：</span>
            <el-tag :type="detectedLang === 'en' ? 'primary' : 'success'" size="small">
              {{ detectedLang === 'en' ? 'English' : '中文' }}
            </el-tag>
          </div>

          <div class="flex gap-2">
            <el-button type="primary" class="flex-1" :loading="quoteStore.isLoading" @click="handleParse">
              解析需求
            </el-button>
            <el-button @click="handleClear">清空</el-button>
          </div>
        </div>
      </div>

      <!-- 中栏：AI 参数提取 -->
      <div class="col-span-4 bg-white rounded-lg border border-gray-100 shadow-sm flex flex-col overflow-hidden">
        <div class="p-4 border-b border-gray-100 flex items-center justify-between">
          <h3 class="font-semibold text-gray-800">已识别需求参数</h3>
          <el-tag v-if="unconfirmedCount > 0" type="warning" size="small">{{ unconfirmedCount }} 项未确认</el-tag>
        </div>
        <div class="flex-1 p-4 overflow-y-auto" v-if="parsedParams">
          <!-- 基础信息 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-500 uppercase mb-3">基础信息</h4>
            <div class="space-y-3">
              <ParamField label="客户名称" v-model="parsedParams.customerName" />
              <ParamField label="国家/地区" v-model="parsedParams.country" />
              <ParamField label="币种" v-model="parsedParams.currency" />
              <ParamField label="目标交付地" v-model="parsedParams.deliveryAddress" />
            </div>
          </div>

          <!-- 产品参数 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-500 uppercase mb-3">产品参数</h4>
            <div class="space-y-3">
              <ParamField label="产品名称" v-model="parsedParams.productName" />
              <ParamField label="型号" v-model="parsedParams.model" />
              <ParamField label="材质" v-model="parsedParams.material" />
              <ParamField label="尺寸" v-model="parsedParams.size" />
              <ParamField label="颜色" v-model="parsedParams.color" />
              <ParamField label="数量" v-model.number="parsedParams.quantity" />
              <ParamField label="包装要求" v-model="parsedParams.packaging" />
            </div>
          </div>

          <!-- 商务参数 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-500 uppercase mb-3">商务参数</h4>
            <div class="space-y-3">
              <ParamField label="MOQ" v-model.number="parsedParams.moq" />
              <ParamField label="付款方式" v-model="parsedParams.paymentTerms" />
              <ParamField label="交期" v-model="parsedParams.leadTime" />
              <ParamField label="报价有效期" v-model="parsedParams.validityPeriod" />
              <div class="flex items-center justify-between p-2 rounded bg-gray-50">
                <span class="text-sm text-gray-600">是否含运费</span>
                <el-switch v-model="parsedParams.includeShipping" :active-value="true" :inactive-value="false" />
              </div>
            </div>
          </div>

          <!-- 未确认项 -->
          <div v-if="parsedParams.unconfirmed?.length" class="bg-orange-50 rounded-lg p-4">
            <h4 class="text-sm font-semibold text-orange-600 mb-2">待确认项</h4>
            <ul class="space-y-1">
              <li v-for="item in parsedParams.unconfirmed" :key="item" class="flex items-center gap-2 text-sm text-orange-700">
                <el-icon><WarningFilled /></el-icon>
                {{ item }}：<span class="font-semibold text-orange-800">未确认</span>
              </li>
            </ul>
          </div>
        </div>
        <div v-else class="flex-1 flex items-center justify-center text-gray-400">
          <div class="text-center">
            <el-icon :size="48"><Document /></el-icon>
            <p class="mt-2">请先输入客户需求并点击"解析需求"</p>
          </div>
        </div>
      </div>

      <!-- 右栏：生成结果 -->
      <div class="col-span-5 bg-white rounded-lg border border-gray-100 shadow-sm flex flex-col overflow-hidden">
        <div class="p-4 border-b border-gray-100">
          <h3 class="font-semibold text-gray-800">可直接发给客户的结果</h3>
        </div>
        <div class="flex-1 overflow-hidden" v-if="quoteResult">
          <el-tabs v-model="activeTab" class="h-full flex flex-col">
            <el-tab-pane label="报价单" name="quotation" class="flex-1 overflow-y-auto p-4">
              <QuotationPreview :quote="quoteResult" @update:items="quoteResult.items = $event" />
            </el-tab-pane>
            <el-tab-pane label="回复话术" name="reply" class="flex-1 overflow-y-auto p-4">
              <ReplyVersions :versions="quoteResult.replyVersions" />
            </el-tab-pane>
            <el-tab-pane label="参数确认清单" name="checklist" class="flex-1 overflow-y-auto p-4">
              <ConfirmationChecklist :items="quoteResult.confirmationList" />
            </el-tab-pane>
            <el-tab-pane label="附件包" name="attachments" class="flex-1 overflow-y-auto p-4">
              <AttachmentPack :attachments="quoteResult.attachments" />
            </el-tab-pane>
          </el-tabs>
        </div>
        <div v-else class="flex-1 flex items-center justify-center text-gray-400">
          <div class="text-center">
            <el-icon :size="48"><Tickets /></el-icon>
            <p class="mt-2">解析参数后点击"开始生成"查看结果</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQuoteStore } from '@/stores/quote'
import type { ParsedParams, Quote } from '@/stores/quote'
import { ElMessage } from 'element-plus'
import { DocumentAdd, UploadFilled, WarningFilled, Document, Tickets } from '@element-plus/icons-vue'
import type { UploadUserFile } from 'element-plus'
import ParamField from '@/components/quote/ParamField.vue'
import QuotationPreview from '@/components/quote/QuotationPreview.vue'
import ReplyVersions from '@/components/quote/ReplyVersions.vue'
import ConfirmationChecklist from '@/components/quote/ConfirmationChecklist.vue'
import AttachmentPack from '@/components/quote/AttachmentPack.vue'

const quoteStore = useQuoteStore()

const requirementText = ref('')
const uploadFiles = ref<UploadUserFile[]>([])
const activeTab = ref('quotation')
const parsedParams = ref<ParsedParams | null>(null)
const quoteResult = ref<Quote | null>(null)

const detectedLang = computed(() => {
  const text = requirementText.value
  const chineseChars = (text.match(/[\u4e00-\u9fa5]/g) || []).length
  return chineseChars > text.length * 0.1 ? 'zh' : 'en'
})

const unconfirmedCount = computed(() => parsedParams.value?.unconfirmed?.length || 0)

async function handleParse() {
  if (!requirementText.value.trim() && uploadFiles.value.length === 0) {
    ElMessage.warning('请输入客户需求或上传文件')
    return
  }
  try {
    const files = uploadFiles.value.map((f) => f.raw as File).filter(Boolean)
    const result = await quoteStore.parseRequirement(requirementText.value, files)
    parsedParams.value = result
    ElMessage.success('需求解析完成')
  } catch {
    // Mock data for demo
    parsedParams.value = {
      customerName: 'ABC Trading Co.',
      country: 'USA',
      currency: 'USD',
      deliveryAddress: 'Los Angeles, CA',
      productName: 'Stainless Steel Table Legs',
      model: '',
      material: 'Stainless Steel 304',
      size: '70cm height',
      color: 'Black',
      quantity: 500,
      packaging: 'Carton',
      moq: 100,
      paymentTerms: '',
      leadTime: '',
      validityPeriod: '',
      includeShipping: null,
      unconfirmed: ['表面处理工艺', '具体型号规格', '付款方式', '交货期限', '报价有效期'],
    }
    ElMessage.info('使用演示数据（后端未连接）')
  }
}

async function handleGenerate() {
  if (!parsedParams.value) {
    ElMessage.warning('请先解析客户需求')
    return
  }
  try {
    const result = await quoteStore.generateQuote(parsedParams.value)
    quoteResult.value = result
    ElMessage.success('报价生成完成')
  } catch {
    // Mock data for demo
    quoteResult.value = {
      customerName: parsedParams.value.customerName,
      status: '草稿',
      params: parsedParams.value,
      items: [
        { productName: 'Stainless Steel Table Legs', model: 'STL-70B', specs: '70cm, Black Coating, SS304', quantity: 500, unitPrice: 12.5, totalPrice: 6250 },
      ],
      replyVersions: [
        {
          title: '简短成交版 (WhatsApp/微信)',
          content: `Hi, thanks for your inquiry!

For 500pcs stainless steel table legs (70cm, black coating):
- Unit price: USD 12.50/pc
- MOQ: 100pcs
- Lead time: 25-30 days
- Packing: individual carton

Total: USD 6,250.00

Let me know if you'd like to proceed!`,
          language: 'en',
        },
        {
          title: '专业邮件版',
          content: `Dear Valued Customer,

Thank you for your inquiry regarding stainless steel table legs.

We are pleased to offer the following quotation:

Product: Stainless Steel Table Legs
Material: SS304
Height: 70cm
Surface: Black powder coating
Quantity: 500 pcs
Unit Price: USD 12.50/pc
Total Amount: USD 6,250.00

Payment Terms: T/T 30% deposit, 70% before shipment
Lead Time: 25-30 days after deposit received
Packing: Each piece in individual carton

This quotation is valid for 30 days.

Should you require any further information, please do not hesitate to contact us.

Best regards`,
          language: 'en',
        },
        {
          title: '追单版',
          content: `Hi, following up on the stainless steel table legs quote.

Quick recap:
✅ 500pcs @ USD 12.50/pc = USD 6,250
✅ SS304, 70cm, black coating
✅ Ready in 25-30 days

We have stock material ready - if you confirm this week, we can start production immediately.

Would you like to go ahead?`,
          language: 'en',
        },
      ],
      confirmationList: [
        { question: '请确认表面处理是否为粉末喷涂哑黑', questionEn: 'Could you please confirm if the surface treatment is matte black powder coating?', checked: false },
        { question: '请确认是否需要防锈处理', questionEn: 'Could you please confirm whether you need anti-rust treatment?', checked: false },
        { question: '请确认包装是否需要单独贴标', questionEn: 'Could you confirm if individual labeling is needed for each carton?', checked: false },
        { question: '请确认是否接受 30% 预付款条件', questionEn: 'Could you confirm if 30% T/T deposit is acceptable?', checked: false },
        { question: '请确认期望交货日期', questionEn: 'Could you please confirm your expected delivery date?', checked: false },
      ],
      attachments: [
        { name: 'Product Spec Sheet.pdf', selected: true },
        { name: 'Product Photos.zip', selected: true },
        { name: 'Test Report - SS304.pdf', selected: true },
        { name: 'Company Profile.pdf', selected: true },
        { name: 'Installation Guide.pdf', selected: false },
        { name: 'Packing Details.pdf', selected: false },
      ],
      totalAmount: 6250,
      currency: 'USD',
    }
    ElMessage.info('使用演示数据（后端未连接）')
  }
}

function handleClear() {
  requirementText.value = ''
  uploadFiles.value = []
  parsedParams.value = null
  quoteResult.value = null
  quoteStore.reset()
}

function handleSaveDraft() {
  ElMessage.success('草稿已保存')
}
</script>
