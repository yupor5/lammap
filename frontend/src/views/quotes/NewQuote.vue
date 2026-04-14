<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">把客户需求变成可发送报价</h1>
        <p class="text-gray-500 mt-1">粘贴客户发来的需求，AI 自动解析并生成报价</p>
      </div>
      <div class="flex gap-2">
        <el-button @click="handleSaveDraft">保存草稿</el-button>
        <el-button type="primary" :icon="DocumentAdd" @click="handleGenerate" :loading="quoteStore.isLoading" :disabled="quoteStore.isLoading">
          开始生成
        </el-button>
      </div>
    </div>

    <div class="grid grid-cols-12 gap-4 min-h-0" style="height: calc(100vh - 180px);">
      <!-- 左栏：客户需求输入 -->
      <div class="col-span-3 min-h-0 bg-white rounded-lg border border-gray-100 shadow-sm flex flex-col overflow-hidden">
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

          <div class="flex gap-2">
            <el-button class="flex-1" @click="openComposeDialog">对话生成</el-button>
            <el-button class="flex-1" @click="openExamplesDialog">示例</el-button>
            <el-button type="primary" class="flex-1" :loading="quoteStore.isLoading" @click="handleParse">解析需求</el-button>
            <el-button class="flex-1" @click="handleClear">清空</el-button>
          </div>

          <el-upload
            v-model:file-list="uploadFiles"
            drag
            multiple
            :http-request="handleUpload"
            :limit="10"
            accept=".pdf,.doc,.docx,.xls,.xlsx,.png,.jpg,.jpeg"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">拖拽文件到此或 <em>点击上传</em></div>
            <template #tip>
              <div class="el-upload__tip">支持 PDF/Word/Excel/图片，单文件 ≤10MB（本 Sprint 仅保存附件，不参与解析）</div>
            </template>
          </el-upload>

          <div class="flex items-center gap-2">
            <span class="text-sm text-gray-500">语言识别：</span>
            <el-tag :type="detectedLang === 'en' ? 'primary' : 'success'" size="small">
              {{ detectedLang === 'en' ? 'English' : '中文' }}
            </el-tag>
          </div>

          <!-- 操作按钮已上移到输入框下方 -->
        </div>
      </div>

      <el-dialog v-model="composeVisible" title="对话生成客户询盘" width="520px">
        <div class="space-y-3">
          <div class="text-sm text-gray-500">按提示回答几句话，我会帮你生成可粘贴的客户询盘文本。</div>

          <el-form label-width="110px">
            <el-form-item label="产品/品名">
              <el-input v-model="composeAnswers.product" placeholder="例如：stainless steel table legs" />
            </el-form-item>
            <el-form-item label="型号/规格">
              <el-input v-model="composeAnswers.spec" placeholder="例如：70cm height / size / model" />
            </el-form-item>
            <el-form-item label="材质/颜色">
              <el-input v-model="composeAnswers.materialColor" placeholder="例如：stainless steel, black coating" />
            </el-form-item>
            <el-form-item label="数量">
              <el-input v-model="composeAnswers.quantity" placeholder="例如：500 pcs" />
            </el-form-item>
            <el-form-item label="交付地">
              <el-input v-model="composeAnswers.delivery" placeholder="例如：Delivery to LA" />
            </el-form-item>
            <el-form-item label="包装/其他">
              <el-input v-model="composeAnswers.other" placeholder="例如：packing by carton / payment term / lead time" />
            </el-form-item>
          </el-form>
        </div>

        <template #footer>
          <div class="flex items-center justify-between w-full">
            <el-button @click="composeVisible = false">取消</el-button>
            <div class="flex gap-2">
              <el-button @click="resetCompose">重置</el-button>
              <el-button type="primary" :loading="composeLoading" @click="composeToTextarea">生成到输入框</el-button>
            </div>
          </div>
        </template>
      </el-dialog>

      <el-dialog v-model="examplesVisible" title="示例库（分组管理）" width="900px">
        <div class="space-y-3">
          <div class="text-sm text-gray-500">
            你可以按分组管理示例：增删改分组、为分组填写提示词，并用 AI 一键生成本组示例。点击示例会回填到左侧输入框。
          </div>

          <div class="flex gap-3">
            <!-- 左侧：分组列表 -->
            <div class="w-56 shrink-0 border border-gray-100 rounded-lg overflow-hidden bg-white">
              <div class="p-3 border-b border-gray-100 flex items-center justify-between">
                <div class="font-semibold text-gray-800 text-sm">分组</div>
                <el-button size="small" @click="openGroupEditor()">新增</el-button>
              </div>
              <div class="max-h-[520px] overflow-y-auto">
                <div
                  v-for="g in exampleGroups"
                  :key="g.id"
                  class="px-3 py-2 text-sm cursor-pointer hover:bg-gray-50 flex items-center justify-between"
                  :class="selectedGroupId === g.id ? 'bg-blue-50' : ''"
                  @click="selectedGroupId = g.id"
                >
                  <div class="truncate">{{ g.name }}</div>
                  <el-dropdown @command="(cmd) => onGroupCommand(cmd, g.id)">
                    <span class="text-gray-400 hover:text-gray-600">⋯</span>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="edit">编辑</el-dropdown-item>
                        <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </div>
            </div>

            <!-- 右侧：分组详情 -->
            <div class="flex-1 border border-gray-100 rounded-lg p-3 bg-white">
              <div v-if="selectedGroup" class="space-y-3">
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <div class="font-semibold text-gray-800">{{ selectedGroup.name }}</div>
                    <div class="text-xs text-gray-500 mt-1 whitespace-pre-wrap">{{ selectedGroup.promptHint || '（未设置提示词）' }}</div>
                  </div>
                  <div class="flex gap-2 shrink-0">
                    <el-button size="small" @click="openGroupEditor(selectedGroup.id)">编辑分组</el-button>
                    <el-button size="small" :loading="aiGenLoading" @click="openAiGenDialog">AI 生成示例</el-button>
                    <el-button size="small" @click="openExampleEditor()">新增示例</el-button>
                  </div>
                </div>

                <div class="grid grid-cols-1 gap-3 max-h-[480px] overflow-y-auto pr-1">
                  <div
                    v-for="ex in selectedGroup.examples"
                    :key="ex.id"
                    class="border border-gray-100 rounded-lg p-3 hover:border-blue-300 hover:bg-blue-50 cursor-pointer transition"
                    @click="applyExample(ex)"
                  >
                    <div class="flex items-center justify-between mb-2">
                      <div class="font-semibold text-gray-800">{{ ex.title }}</div>
                      <div class="flex items-center gap-2">
                        <el-tag size="small" :type="ex.lang === 'zh' ? 'success' : 'primary'">{{ ex.lang === 'zh' ? '中文' : 'English' }}</el-tag>
                        <el-button text size="small" @click.stop="openExampleEditor(ex.id)">编辑</el-button>
                        <el-button text size="small" type="danger" @click.stop="deleteExample(ex.id)">删除</el-button>
                      </div>
                    </div>
                    <div class="text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">{{ ex.content }}</div>
                  </div>
                </div>
              </div>

              <div v-else class="text-gray-400 text-sm flex items-center justify-center h-[540px]">
                请先选择一个分组
              </div>
            </div>
          </div>
        </div>

        <template #footer>
          <div class="flex items-center justify-between w-full">
            <el-button @click="resetExampleGroups">恢复内置分组</el-button>
            <el-button @click="examplesVisible = false">关闭</el-button>
          </div>
        </template>
      </el-dialog>

      <!-- 分组编辑 -->
      <el-dialog v-model="groupEditorVisible" :title="groupEditorId ? '编辑分组' : '新增分组'" width="640px">
        <el-form label-width="90px">
          <el-form-item label="分组名称">
            <el-input v-model="groupEditor.name" placeholder="例如：金属加工 / 包装日用品 / 塑料注塑" />
          </el-form-item>
          <el-form-item label="提示词">
            <el-input
              v-model="groupEditor.promptHint"
              type="textarea"
              :autosize="{ minRows: 6, maxRows: 10 }"
              placeholder="用于 AI 生成本组示例的提示词（越具体越稳定）。建议包含：品类范围、常见要素（数量/交付地/包装/付款/交期）、语气（真实采购邮件）、禁止项（不要编造价格/认证编号）。"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <div class="flex items-center justify-between w-full">
            <div class="text-xs text-gray-500">提示词会影响“AI 生成示例”的质量。</div>
            <div class="flex gap-2">
              <el-button @click="groupEditorVisible = false">取消</el-button>
              <el-button type="primary" @click="saveGroup">保存</el-button>
            </div>
          </div>
        </template>
      </el-dialog>

      <!-- 示例编辑 -->
      <el-dialog v-model="exampleEditorVisible" :title="exampleEditorId ? '编辑示例' : '新增示例'" width="720px">
        <el-form label-width="90px">
          <el-form-item label="标题">
            <el-input v-model="exampleEditor.title" placeholder="例如：不锈钢桌腿（英文，参数较全）" />
          </el-form-item>
          <el-form-item label="语言">
            <el-select v-model="exampleEditor.lang" style="width: 160px">
              <el-option label="中文" value="zh" />
              <el-option label="English" value="en" />
            </el-select>
          </el-form-item>
          <el-form-item label="内容">
            <el-input v-model="exampleEditor.content" type="textarea" :autosize="{ minRows: 8, maxRows: 14 }" placeholder="完整客户询盘文本（可多段）" />
          </el-form-item>
          <el-form-item label="提示词">
            <el-input
              v-model="exampleEditor.promptHint"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              placeholder="（可选）本条示例的补充提示：例如“强调打样/图纸/交期紧/中英混合”等。"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <div class="flex items-center justify-between w-full">
            <div />
            <div class="flex gap-2">
              <el-button @click="exampleEditorVisible = false">取消</el-button>
              <el-button type="primary" @click="saveExample">保存</el-button>
            </div>
          </div>
        </template>
      </el-dialog>

      <!-- AI 生成示例 -->
      <el-dialog v-model="aiGenVisible" title="AI 生成本组示例" width="680px">
        <div class="space-y-3">
          <div class="text-sm text-gray-500">会按本分组的“提示词”生成多条示例，并追加到本组中。</div>
          <el-form label-width="90px">
            <el-form-item label="生成语言">
              <el-select v-model="aiGenLang" style="width: 160px">
                <el-option label="中文" value="zh" />
                <el-option label="English" value="en" />
              </el-select>
            </el-form-item>
            <el-form-item label="数量">
              <el-input-number v-model="aiGenCount" :min="1" :max="10" />
            </el-form-item>
            <el-form-item label="补充说明">
              <el-input
                v-model="aiGenExtraHint"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 8 }"
                placeholder="（可选）临时追加要求：例如“更偏 WhatsApp 语气 / 更缺参 / 偏家具类 / 偏电子类”。"
              />
            </el-form-item>
          </el-form>
        </div>
        <template #footer>
          <div class="flex items-center justify-between w-full">
            <el-button @click="aiGenVisible = false">取消</el-button>
            <el-button type="primary" :loading="aiGenLoading" @click="runAiGenerateExamples">开始生成</el-button>
          </div>
        </template>
      </el-dialog>

      <!-- 中栏：AI 参数提取 -->
      <div class="col-span-4 min-h-0 bg-white rounded-lg border border-gray-100 shadow-sm flex flex-col overflow-hidden">
        <div class="p-4 border-b border-gray-100 flex items-center justify-between">
          <h3 class="font-semibold text-gray-800">已识别需求参数</h3>
          <div class="flex items-center gap-2">
            <el-tag v-if="autoUnconfirmedCount > 0" type="warning" size="small">{{ autoUnconfirmedCount }} 项待补齐</el-tag>
            <el-tag v-else-if="manualOnlyUnconfirmedCount > 0" type="info" size="small">
              {{ manualOnlyUnconfirmedCount }} 项细节待手动确认
            </el-tag>
            <el-button v-if="parsedParams" size="small" @click="runAutoValidate">确认校验</el-button>
            <el-button
              v-if="parsedParams && manualOnlyUnconfirmedCount > 0"
              size="small"
              @click="openManualConfirmDialog"
            >
              手动确认 ({{ manualOnlyUnconfirmedCount }})
            </el-button>
          </div>
        </div>
        <div class="flex-1 p-4 overflow-y-auto" v-if="parsedParams">
          <!-- 基础信息 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-500 uppercase mb-3">基础信息</h4>
            <div class="space-y-3">
              <ParamField ref="refCustomerName" label="客户名称" v-model="parsedParams.customerName" memory-key="customerName" />
              <CountryField ref="refCountry" label="国家/地区" v-model="parsedParams.country" />
              <ParamField ref="refCurrency" label="币种" v-model="parsedParams.currency" memory-key="currency" />
              <ParamField ref="refDelivery" label="目标交付地" v-model="parsedParams.deliveryAddress" memory-key="deliveryAddress" />
            </div>
          </div>

          <!-- 产品参数 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-500 uppercase mb-3">产品参数</h4>
            <div class="space-y-3">
              <ParamField ref="refProductName" label="产品名称" v-model="parsedParams.productName" memory-key="productName" />
              <ParamField ref="refModel" label="型号" v-model="parsedParams.model" memory-key="model" />
              <ParamField ref="refMaterial" label="材质" v-model="parsedParams.material" memory-key="material" />
              <ParamField ref="refSize" label="尺寸" v-model="parsedParams.size" memory-key="size" />
              <ParamField ref="refColor" label="颜色" v-model="parsedParams.color" memory-key="color" />
              <ParamField label="数量" v-model.number="parsedParams.quantity" numeric />
              <ParamField ref="refPackaging" label="包装要求" v-model="parsedParams.packaging" memory-key="packaging" />
            </div>
          </div>

          <!-- 商务参数 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-500 uppercase mb-3">商务参数</h4>
            <div class="space-y-3">
              <ParamField label="MOQ" v-model.number="parsedParams.moq" numeric />
              <ParamField ref="refPaymentTerms" label="付款方式" v-model="parsedParams.paymentTerms" memory-key="paymentTerms" />
              <ParamField ref="refLeadTime" label="交期" v-model="parsedParams.leadTime" memory-key="leadTime" />
              <ParamField ref="refValidity" label="报价有效期" v-model="parsedParams.validityPeriod" memory-key="validityPeriod" />
              <div class="flex items-center justify-between p-2 rounded bg-gray-50">
                <span class="text-sm text-gray-600">是否含运费</span>
                <el-switch v-model="parsedParams.includeShipping" :active-value="true" :inactive-value="false" />
              </div>
            </div>
          </div>

          <!-- 产品库匹配推荐 -->
          <div v-if="matchedProducts.length > 0" class="bg-blue-50 rounded-lg p-4 mb-4">
            <h4 class="text-sm font-semibold text-blue-600 mb-2">产品库推荐匹配</h4>
            <ul class="space-y-2">
              <li v-for="p in matchedProducts" :key="p.id" class="flex items-center justify-between text-sm bg-white rounded p-2 border border-blue-100">
                <div>
                  <span class="font-medium">{{ p.name }}</span>
                  <span class="text-gray-400 ml-2">{{ p.sku }}</span>
                  <span class="text-gray-400 ml-2">{{ p.material }}</span>
                </div>
                <div class="text-blue-600 font-medium">USD {{ (p.price || 0).toFixed(2) }}</div>
              </li>
            </ul>
          </div>

          <!-- 未确认项 -->
          <div v-if="derivedUnconfirmed.length" class="bg-orange-50 rounded-lg p-4">
            <h4 class="text-sm font-semibold text-orange-600 mb-2">待确认项</h4>
            <ul class="space-y-1">
              <li v-for="item in derivedUnconfirmed" :key="item" class="flex items-center gap-2 text-sm text-orange-700">
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

      <!-- 手动确认弹窗：仅用于无法映射到字段的细项 -->
      <el-dialog v-model="confirmDialogVisible" title="手动确认（无法自动校验的项）" width="720px">
        <div class="space-y-3">
          <div class="text-sm text-gray-500">
            这些项暂无对应字段，系统无法自动判断你是否确认完成。你可以手动勾选“已确认”，它将不再提示。
          </div>

          <el-checkbox-group v-if="manualOnlyUnconfirmed.length" v-model="manualConfirmedItems" class="space-y-2">
            <div
              v-for="item in manualOnlyUnconfirmed"
              :key="item"
              class="flex items-start justify-between gap-3 p-3 rounded border border-gray-100 bg-white"
            >
              <div class="min-w-0">
                <div class="font-semibold text-gray-800 text-sm">{{ item }}</div>
                <div class="text-xs text-gray-500 mt-1">该项暂无对应字段，可手动确认</div>
              </div>
              <el-checkbox :value="item">已确认</el-checkbox>
            </div>
          </el-checkbox-group>
          <div v-else class="text-sm text-gray-400">当前没有待确认项</div>
        </div>

        <template #footer>
          <div class="flex items-center justify-between w-full">
            <el-button @click="confirmDialogVisible = false">关闭</el-button>
            <el-button type="primary" @click="applyManualConfirm">应用确认</el-button>
          </div>
        </template>
      </el-dialog>

      <!-- 右栏：生成结果 -->
      <div class="col-span-5 min-h-0 bg-white rounded-lg border border-gray-100 shadow-sm flex flex-col overflow-hidden">
        <div class="p-4 border-b border-gray-100 shrink-0">
          <h3 class="font-semibold text-gray-800">可直接发给客户的结果</h3>
        </div>
        <div class="flex-1 min-h-0 overflow-hidden flex flex-col" v-if="quoteResult">
          <el-tabs v-model="activeTab" class="quote-result-tabs h-full min-h-0 flex flex-col">
            <el-tab-pane label="报价单" name="quotation" class="p-4">
              <QuotationPreview :quote="quoteResult" @update:items="onItemsUpdate" />
              <div class="mt-4 border-t border-gray-100 pt-4 space-y-2">
                <div class="flex items-center justify-between">
                  <div class="font-semibold text-gray-800">套用模板（报价单正文）</div>
                  <div class="flex items-center gap-2">
                    <el-select v-model="templateLang" size="small" class="w-28" @change="loadTemplatesForCurrentLang">
                      <el-option label="中文" value="zh" />
                      <el-option label="English" value="en" />
                    </el-select>
                    <el-select v-model="selectedTemplateIdByCat.quotation" size="small" class="w-64" placeholder="选择报价单模板">
                      <el-option v-for="t in templatesOfCategory('quotation')" :key="t.id" :label="t.name" :value="t.id" />
                    </el-select>
                    <el-button size="small" type="primary" @click="applyTemplate('quotation')">套用</el-button>
                  </div>
                </div>
                <el-input
                  v-model="renderedTextByCat.quotation"
                  type="textarea"
                  :autosize="{ minRows: 6, maxRows: 16 }"
                  placeholder="套用模板后会生成可编辑的报价单正文（可复制发客户）…"
                />
              </div>
            </el-tab-pane>
            <el-tab-pane label="回复话术" name="reply" class="p-4">
              <ReplyVersions :versions="quoteResult.replyVersions" :parsed-params="parsedParams" @update:versions="quoteResult.replyVersions = $event" />
              <div class="mt-4 border-t border-gray-100 pt-4 space-y-2">
                <div class="flex items-center justify-between">
                  <div class="font-semibold text-gray-800">套用模板（邮件/聊天正文）</div>
                  <div class="flex items-center gap-2">
                    <el-select v-model="templateLang" size="small" class="w-28" @change="loadTemplatesForCurrentLang">
                      <el-option label="中文" value="zh" />
                      <el-option label="English" value="en" />
                    </el-select>
                    <el-select v-model="selectedTemplateIdByCat.email" size="small" class="w-56" placeholder="选择邮件模板">
                      <el-option v-for="t in templatesOfCategory('email')" :key="t.id" :label="t.name" :value="t.id" />
                    </el-select>
                    <el-button size="small" @click="applyTemplate('email')">套用邮件</el-button>
                    <el-select v-model="selectedTemplateIdByCat.chat" size="small" class="w-56" placeholder="选择聊天模板">
                      <el-option v-for="t in templatesOfCategory('chat')" :key="t.id" :label="t.name" :value="t.id" />
                    </el-select>
                    <el-button size="small" @click="applyTemplate('chat')">套用聊天</el-button>
                  </div>
                </div>
                <el-input
                  v-model="renderedTextByCat.email"
                  type="textarea"
                  :autosize="{ minRows: 5, maxRows: 12 }"
                  placeholder="套用邮件模板后生成的正文（可编辑/复制）…"
                />
                <el-input
                  v-model="renderedTextByCat.chat"
                  type="textarea"
                  :autosize="{ minRows: 4, maxRows: 10 }"
                  placeholder="套用聊天模板后生成的正文（可编辑/复制）…"
                />
              </div>
            </el-tab-pane>
            <el-tab-pane label="参数确认清单" name="checklist" class="p-4">
              <ConfirmationChecklist :items="quoteResult.confirmationList" />
              <div class="mt-4 border-t border-gray-100 pt-4 space-y-2">
                <div class="flex items-center justify-between">
                  <div class="font-semibold text-gray-800">套用模板（确认清单正文）</div>
                  <div class="flex items-center gap-2">
                    <el-select v-model="templateLang" size="small" class="w-28" @change="loadTemplatesForCurrentLang">
                      <el-option label="中文" value="zh" />
                      <el-option label="English" value="en" />
                    </el-select>
                    <el-select v-model="selectedTemplateIdByCat.confirmation" size="small" class="w-64" placeholder="选择确认清单模板">
                      <el-option v-for="t in templatesOfCategory('confirmation')" :key="t.id" :label="t.name" :value="t.id" />
                    </el-select>
                    <el-button size="small" type="primary" @click="applyTemplate('confirmation')">套用</el-button>
                  </div>
                </div>
                <el-input
                  v-model="renderedTextByCat.confirmation"
                  type="textarea"
                  :autosize="{ minRows: 5, maxRows: 14 }"
                  placeholder="套用确认模板后生成的正文（可编辑/复制）…"
                />
              </div>
            </el-tab-pane>
            <el-tab-pane label="附件包" name="attachments" class="p-4">
              <AttachmentPack
                :attachments="quoteResult.attachments"
                :loading="attachmentPackLoading"
                :generating-indices="Array.from(aiGeneratingAttachmentIndices)"
                @generate-pack="handleGenerateEmailAttachmentPack"
                @download-all="handleDownloadAttachmentPack"
                @ai-generate="(att, idx) => handleAiGenerateSingleAttachment(att, idx)"
              />
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
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { useQuoteStore } from '@/stores/quote'
import { useProductStore } from '@/stores/product'
import type { ParsedParams, Quote } from '@/stores/quote'
import { mapGeneratePayload } from '@/stores/quote'
import { ElMessage, ElMessageBox } from 'element-plus'
import { DocumentAdd, UploadFilled, WarningFilled, Document, Tickets } from '@element-plus/icons-vue'
import type { UploadUserFile, UploadRequestOptions } from 'element-plus'
import ParamField from '@/components/quote/ParamField.vue'
import CountryField from '@/components/quote/CountryField.vue'
import QuotationPreview from '@/components/quote/QuotationPreview.vue'
import ReplyVersions from '@/components/quote/ReplyVersions.vue'
import ConfirmationChecklist from '@/components/quote/ConfirmationChecklist.vue'
import AttachmentPack from '@/components/quote/AttachmentPack.vue'
import request from '@/utils/request'
import { rememberValue } from '@/utils/paramMemory'
import { loadExampleGroups, saveExampleGroups, resetExampleGroupsToBuiltin, newId, type InquiryExampleGroup, type InquiryExampleItem } from '@/utils/inquiryExampleGroups'
import { useTemplateStore } from '@/stores/template'
import type { Template } from '@/stores/template'
import { renderTemplate } from '@/utils/renderTemplate'

const quoteStore = useQuoteStore()
const productStore = useProductStore()
const templateStore = useTemplateStore()
const route = useRoute()
const router = useRouter()

const requirementText = ref('')
const matchedProducts = ref<any[]>([])
const uploadFiles = ref<UploadUserFile[]>([])
const uploadedRefs = ref<{ filename: string; path: string }[]>([])
const uploadedAttachments = ref<{ name: string; url?: string; selected: boolean; source?: 'upload' }[]>([])
const activeTab = ref('quotation')
const parsedParams = ref<ParsedParams | null>(null)
const quoteResult = ref<Quote | null>(null)
const draftQuoteId = ref<number | null>(null)
const suppressDirty = ref(false)
const isDirty = ref(false)

const detectedLang = computed(() => {
  const text = requirementText.value
  const chineseChars = (text.match(/[\u4e00-\u9fa5]/g) || []).length
  return chineseChars > text.length * 0.1 ? 'zh' : 'en'
})

// 确认校验：用于自动定位仍未补齐的字段
type EditableFieldRef = { startEdit?: () => void; $el?: HTMLElement }
const refCustomerName = ref<EditableFieldRef | null>(null)
const refCountry = ref<EditableFieldRef | null>(null)
const refCurrency = ref<EditableFieldRef | null>(null)
const refDelivery = ref<EditableFieldRef | null>(null)
const refProductName = ref<EditableFieldRef | null>(null)
const refModel = ref<EditableFieldRef | null>(null)
const refMaterial = ref<EditableFieldRef | null>(null)
const refSize = ref<EditableFieldRef | null>(null)
const refColor = ref<EditableFieldRef | null>(null)
const refPackaging = ref<EditableFieldRef | null>(null)
const refPaymentTerms = ref<EditableFieldRef | null>(null)
const refLeadTime = ref<EditableFieldRef | null>(null)
const refValidity = ref<EditableFieldRef | null>(null)

const fieldRefByKey: Record<string, typeof refCustomerName> = {
  customerName: refCustomerName,
  country: refCountry,
  currency: refCurrency,
  deliveryAddress: refDelivery,
  productName: refProductName,
  model: refModel,
  material: refMaterial,
  size: refSize,
  color: refColor,
  packaging: refPackaging,
  paymentTerms: refPaymentTerms,
  leadTime: refLeadTime,
  validityPeriod: refValidity,
}

// 套用模板：选择 + 渲染后的可编辑文本（保存草稿时落库）
type TemplateCategory = 'quotation' | 'email' | 'chat' | 'confirmation'
const templateLang = ref<'zh' | 'en'>('zh')
const selectedTemplateIdByCat = ref<Record<TemplateCategory, number | null>>({
  quotation: null,
  email: null,
  chat: null,
  confirmation: null,
})
const renderedTextByCat = ref<Record<TemplateCategory, string>>({
  quotation: '',
  email: '',
  chat: '',
  confirmation: '',
})

watch(detectedLang, (v) => {
  templateLang.value = v
})

// 注意：不要在 derivedUnconfirmed 初始化前 watch 它，否则会导致运行时 TDZ 报错并白屏

async function loadTemplatesForCurrentLang() {
  await templateStore.fetchTemplates(undefined, templateLang.value)
}

function ctxForTemplate() {
  const p = parsedParams.value
  const q = quoteResult.value
  const items = q?.items || []
  const itemsTable = items
    .map((it, idx) => `${idx + 1}. ${it.productName} | ${it.model} | ${it.specs} | x${it.quantity} | ${q?.currency || ''} ${it.unitPrice} | ${q?.currency || ''} ${it.totalPrice}`)
    .join('\n')

  const unconfirmed = derivedUnconfirmed.value

  return {
    customer_name: p?.customerName || q?.customerName || '',
    country: p?.country || '',
    currency: p?.currency || q?.currency || '',
    delivery_address: p?.deliveryAddress || '',
    lead_time: p?.leadTime || '',
    payment_terms: p?.paymentTerms || '',
    validity: p?.validityPeriod || '',
    include_shipping: p?.includeShipping ?? '',
    product_name: p?.productName || '',
    model: p?.model || '',
    material: p?.material || '',
    size: p?.size || '',
    color: p?.color || '',
    quantity: p?.quantity || '',
    packaging: p?.packaging || '',
    moq: p?.moq || '',
    // 报价相关
    items_table: itemsTable,
    total_amount: q?.totalAmount ?? '',
    // 回复相关（如果模板要用）
    reply_short: q?.replyVersions?.[0]?.content || '',
    reply_professional: q?.replyVersions?.[1]?.content || '',
    reply_followup: q?.replyVersions?.[2]?.content || '',
    // 确认项
    unconfirmed_list: unconfirmed.join('\n'),
  }
}

function templatesOfCategory(cat: TemplateCategory) {
  // 这里用后端过滤更好；先前端过滤（已按 language 拉取）
  return (templateStore.templates as Template[]).filter((t) => t.category === cat && t.language === templateLang.value)
}

function applyTemplate(cat: TemplateCategory) {
  const id = selectedTemplateIdByCat.value[cat]
  if (!id) {
    ElMessage.warning('请先选择模板')
    return
  }
  const tmpl = (templateStore.templates as Template[]).find((t) => t.id === id)
  if (!tmpl) {
    ElMessage.error('模板不存在或未加载')
    return
  }
  const ctx = ctxForTemplate()
  renderedTextByCat.value[cat] = renderTemplate(tmpl.content, ctx)
  ElMessage.success('已套用模板')
}

const composeVisible = ref(false)
const composeLoading = ref(false)
const composeAnswers = ref({
  product: '',
  spec: '',
  materialColor: '',
  quantity: '',
  delivery: '',
  other: '',
})

const examplesVisible = ref(false)
const exampleGroups = ref<InquiryExampleGroup[]>(loadExampleGroups())
const selectedGroupId = ref<string>(exampleGroups.value[0]?.id || '')

// 待确认项：手动确认（用于无法映射到字段的更细颗粒度确认项）
const confirmDialogVisible = ref(false)
const manualConfirmedItems = ref<string[]>([])
const manualConfirmedSet = computed(() => new Set(manualConfirmedItems.value.map((x) => String(x).trim()).filter(Boolean)))

const selectedGroup = computed(() => exampleGroups.value.find((g) => g.id === selectedGroupId.value) || null)

watch(
  exampleGroups,
  (v) => {
    saveExampleGroups(v)
  },
  { deep: true }
)

// 分组编辑
const groupEditorVisible = ref(false)
const groupEditorId = ref<string | null>(null)
const groupEditor = ref({ name: '', promptHint: '' })

// 示例编辑
const exampleEditorVisible = ref(false)
const exampleEditorId = ref<string | null>(null)
const exampleEditor = ref<{ title: string; lang: 'zh' | 'en'; content: string; promptHint: string }>({
  title: '',
  lang: 'en',
  content: '',
  promptHint: '',
})

// AI 生成示例
const aiGenVisible = ref(false)
const aiGenLoading = ref(false)
const aiGenLang = ref<'zh' | 'en'>('en')
const aiGenCount = ref(5)
const aiGenExtraHint = ref('')

function isFilled(v: unknown): boolean {
  if (v == null) return false
  if (typeof v === 'string') return v.trim().length > 0
  if (typeof v === 'number') return Number.isFinite(v) && v !== 0
  if (typeof v === 'boolean') return true
  return Boolean(v)
}

const unconfirmedMap: Record<string, keyof ParsedParams> = {
  客户名称: 'customerName',
  客户名: 'customerName',
  客户姓名: 'customerName',
  客户: 'customerName',
  '国家/地区': 'country',
  国家: 'country',
  地区: 'country',
  币种: 'currency',
  货币: 'currency',
  目标交付地: 'deliveryAddress',
  交付地: 'deliveryAddress',
  交付地址: 'deliveryAddress',
  收货地址: 'deliveryAddress',
  产品名称: 'productName',
  型号: 'model',
  产品型号: 'model',
  材质: 'material',
  玻璃材质: 'material',
  尺寸: 'size',
  颜色: 'color',
  数量: 'quantity',
  包装要求: 'packaging',
  包装: 'packaging',
  '内/外箱自身包装方式': 'packaging',
  '内外箱包装方式': 'packaging',
  内箱包装方式: 'packaging',
  外箱包装方式: 'packaging',
  包装方式: 'packaging',
  MOQ: 'moq',
  起订量: 'moq',
  付款方式: 'paymentTerms',
  交期: 'leadTime',
  量产交期: 'leadTime',
  报价有效期: 'validityPeriod',
  是否含运费: 'includeShipping',
}

const derivedUnconfirmed = computed(() => {
  const p = parsedParams.value
  if (!p) return []

  const missing: string[] = []
  const addIfEmpty = (label: string, key: keyof ParsedParams) => {
    if (!isFilled((p as any)[key])) missing.push(label)
  }
  addIfEmpty('客户名称', 'customerName')
  addIfEmpty('国家/地区', 'country')
  addIfEmpty('币种', 'currency')
  addIfEmpty('目标交付地', 'deliveryAddress')
  addIfEmpty('产品名称', 'productName')
  addIfEmpty('型号', 'model')
  addIfEmpty('材质', 'material')
  addIfEmpty('尺寸', 'size')
  addIfEmpty('颜色', 'color')
  addIfEmpty('数量', 'quantity')
  addIfEmpty('包装要求', 'packaging')
  addIfEmpty('MOQ', 'moq')
  addIfEmpty('付款方式', 'paymentTerms')
  addIfEmpty('交期', 'leadTime')
  addIfEmpty('报价有效期', 'validityPeriod')
  addIfEmpty('是否含运费', 'includeShipping')

  const aiList = Array.isArray(p.unconfirmed) ? p.unconfirmed : []
  const merged = [...aiList, ...missing]
  const unique = Array.from(new Set(merged.map((x) => String(x).trim()).filter(Boolean)))

  // 对已填字段自动“确认”：如果 label 能映射到字段且字段已填，则不再显示未确认
  return unique.filter((label) => {
    if (manualConfirmedSet.value.has(label)) return false
    const key = unconfirmedMap[label]
    if (!key) return true
    return !isFilled((p as any)[key])
  })
})

const unconfirmedCount = computed(() => derivedUnconfirmed.value.length)

const manualOnlyUnconfirmed = computed(() => derivedUnconfirmed.value.filter((x) => !unconfirmedMap[x]))
const manualOnlyUnconfirmedCount = computed(() => manualOnlyUnconfirmed.value.length)

// 仅统计“能映射到字段且字段为空”的未确认项（也就是你真正需要补填的项）
const autoUnconfirmed = computed(() => {
  const p = parsedParams.value
  if (!p) return []
  return derivedUnconfirmed.value
    .filter((label) => !!unconfirmedMap[label])
    .map((label) => ({ label, key: unconfirmedMap[label] }))
    .filter((x) => x.key && !isFilled((p as any)[x.key]))
})
const autoUnconfirmedCount = computed(() => autoUnconfirmed.value.length)

watch(
  [requirementText, parsedParams, quoteResult, uploadFiles],
  () => {
    if (suppressDirty.value) return
    isDirty.value = true
  },
  { deep: true }
)

let autoSaveTimer: ReturnType<typeof setTimeout> | null = null

function buildDraftPayload(): Record<string, unknown> {
  return {
    id: draftQuoteId.value ?? undefined,
    customerName: parsedParams.value?.customerName || '',
    country: parsedParams.value?.country || '',
    currency: parsedParams.value?.currency || 'USD',
    deliveryAddress: parsedParams.value?.deliveryAddress || '',
    status: '草稿',
    totalAmount: quoteResult.value?.totalAmount || 0,
    rawRequirement: requirementText.value,
    parsedParams: parsedParams.value,
    replyVersions: quoteResult.value?.replyVersions || [],
    confirmationList: quoteResult.value?.confirmationList || [],
    attachmentList: quoteResult.value?.attachments || [],
    items: quoteResult.value?.items || [],
    templateMeta: {
      language: templateLang.value,
      selectedTemplateIdByCat: selectedTemplateIdByCat.value,
    },
    renderedContents: renderedTextByCat.value,
  }
}

async function runAutoSaveDraft() {
  if (suppressDirty.value) return
  if (!isDirty.value) return
  if (!parsedParams.value && !quoteResult.value) return
  try {
    const saved = (await quoteStore.saveQuote(buildDraftPayload())) as { id?: number }
    if (saved?.id) {
      draftQuoteId.value = saved.id
      isDirty.value = false
    }
  } catch {
    // 静默失败，避免打断输入
  }
}

watch(isDirty, (dirty) => {
  if (!dirty) return
  if (!parsedParams.value && !quoteResult.value) return
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(() => {
    autoSaveTimer = null
    void runAutoSaveDraft()
  }, 25000)
})

onUnmounted(() => {
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
})

function emptyParsedParams(): ParsedParams {
  return {
    customerName: '',
    country: '',
    currency: 'USD',
    deliveryAddress: '',
    productName: '',
    model: '',
    material: '',
    size: '',
    color: '',
    quantity: 0,
    packaging: '',
    moq: 0,
    paymentTerms: '',
    leadTime: '',
    validityPeriod: '',
    includeShipping: null,
    unconfirmed: [],
  }
}

function parseJsonField<T>(raw: unknown, fallback: T): T {
  if (raw == null || raw === '') return fallback
  if (typeof raw === 'string') {
    try {
      const v = JSON.parse(raw) as unknown
      // 兼容“双重编码”的 JSON：例如 raw 是 "\"[ {...} ]\"" 这种字符串
      if (typeof v === 'string') {
        try {
          return JSON.parse(v) as T
        } catch {
          return fallback
        }
      }
      return v as T
    } catch {
      return fallback
    }
  }
  return raw as T
}

function parseParsedParams(raw: unknown): ParsedParams {
  const base = emptyParsedParams()
  if (raw == null || raw === '') return base

  const objRaw = typeof raw === 'string' ? parseJsonField<unknown>(raw, {}) : raw
  if (!objRaw || typeof objRaw !== 'object' || Array.isArray(objRaw)) return base

  const obj = objRaw as Record<string, unknown>
  const u = obj?.unconfirmed
  const unconfirmed = Array.isArray(u) ? u.map((x) => String(x)) : []
  return {
    ...base,
    ...obj,
    unconfirmed,
  } as ParsedParams
}

function buildQuoteResultFromApi(q: Record<string, unknown>): Quote {
  const replyVersions = parseJsonField<Quote['replyVersions']>(q.replyVersions, [])
  const confirmationList = parseJsonField<Quote['confirmationList']>(q.confirmationList, [])
  const attachments = parseJsonField<Quote['attachments']>(q.attachmentList, [])
  const itemsRaw = (q.items as Record<string, unknown>[]) || []
  const items = itemsRaw.map((it) => ({
    productName: String(it.productName ?? ''),
    model: String(it.model ?? ''),
    specs: String(it.specs ?? ''),
    quantity: Number(it.quantity) || 0,
    unitPrice: Number(it.unitPrice) || 0,
    totalPrice: Number(it.totalPrice) || 0,
    remark: String(it.remark ?? ''),
  }))
  return {
    customerName: String(q.customerName ?? ''),
    status: String(q.status ?? '草稿'),
    params: parseParsedParams(q.parsedParams),
    items,
    replyVersions,
    confirmationList,
    attachments,
    totalAmount: Number(q.totalAmount) || 0,
    currency: String(q.currency ?? 'USD'),
  }
}

async function hydrateFromQuote(id: string) {
  suppressDirty.value = true
  try {
    const q = (await quoteStore.fetchQuote(Number(id))) as unknown as Record<string, unknown>
    draftQuoteId.value = Number(q.id)
    requirementText.value = String(q.rawRequirement ?? '')
    const pp = parseParsedParams(q.parsedParams)
    // 兜底：如果历史数据里 parsedParams 为空/缺失，则用 Quote 的顶层字段回填一份，避免“看见了又突然没了”
    if (
      !isFilled(pp.customerName) &&
      !isFilled(pp.country) &&
      !isFilled(pp.currency) &&
      !isFilled(pp.deliveryAddress) &&
      !isFilled(pp.productName) &&
      !isFilled(pp.model) &&
      !isFilled(pp.material) &&
      !isFilled(pp.size) &&
      !isFilled(pp.color) &&
      !isFilled(pp.packaging) &&
      !isFilled(pp.paymentTerms) &&
      !isFilled(pp.leadTime) &&
      !isFilled(pp.validityPeriod)
    ) {
      pp.customerName = String(q.customerName ?? '')
      pp.country = String(q.country ?? '')
      pp.currency = String(q.currency ?? 'USD')
      pp.deliveryAddress = String(q.deliveryAddress ?? '')
    }
    parsedParams.value = pp
    quoteResult.value = buildQuoteResultFromApi(q)
    isDirty.value = false
  } catch {
    ElMessage.error('加载报价失败')
  } finally {
    suppressDirty.value = false
  }
}

onMounted(async () => {
  const fromId = route.query.from
  if (fromId && typeof fromId === 'string') {
    await hydrateFromQuote(fromId)
  }
})

watch(
  () => route.query.from,
  async (v) => {
    if (v && typeof v === 'string') await hydrateFromQuote(v)
  }
)

onBeforeRouteLeave((_to, _from, next) => {
  stopGeneratePolling()
  if (!isDirty.value) {
    next()
    return
  }
  ElMessageBox.confirm('当前内容尚未保存，确定离开？', '提示', {
    type: 'warning',
    confirmButtonText: '离开',
    cancelButtonText: '取消',
  })
    .then(() => next())
    .catch(() => next(false))
})

function buildParseContent(): string {
  let t = requirementText.value.trim()
  const hints = uploadedRefs.value.map((x) => `[附件已上传] ${x.filename} (${x.path})`).join('\n')
  if (hints) {
    t = t ? `${t}\n\n${hints}` : hints
  }
  return t
}

async function handleUpload(opt: UploadRequestOptions) {
  const fd = new FormData()
  fd.append('file', opt.file as File)
  try {
    const res: any = await request.post('/upload', fd)
    uploadedRefs.value.push({ filename: res.data.filename, path: res.data.path })
    const url = String(res.data.url || res.data.path || '').trim()
    if (url) {
      const att = { name: String(res.data.filename || ''), url, selected: true, source: 'upload' as const }
      uploadedAttachments.value.push(att)
      if (quoteResult.value) {
        const exists = (quoteResult.value.attachments || []).some((x) => String(x.name).trim() === att.name)
        if (!exists) {
          quoteResult.value.attachments = [...(quoteResult.value.attachments || []), att]
        }
      }
    }
    ElMessage.success(`已上传 ${res.data.filename}`)
    opt.onSuccess?.(res)
  } catch (e) {
    opt.onError?.(e as Error)
  }
}

async function handleParse() {
  const content = buildParseContent()
  if (!content.trim()) {
    ElMessage.warning('请输入客户需求或上传文件')
    return
  }
  try {
    const result = await quoteStore.parseRequirement(content, [])
    parsedParams.value = parseParsedParams(result)
    ElMessage.success('需求解析完成')
    tryMatchProducts(parsedParams.value)
  } catch (e) {
    console.error(e)
    ElMessage.error('解析失败，请检查网络或稍后重试')
  }
}

function openComposeDialog() {
  composeVisible.value = true
}

function openExamplesDialog() {
  examplesVisible.value = true
}

function applyExample(ex: { title: string; lang: 'zh' | 'en'; content: string }) {
  requirementText.value = ex.content
  examplesVisible.value = false
  ElMessage.success('已回填示例询盘')
}

function runAutoValidate() {
  const p = parsedParams.value
  if (!p) return

  const needFill = derivedUnconfirmed.value
    .filter((label) => !!unconfirmedMap[label])
    .map((label) => ({ label, key: unconfirmedMap[label] }))
    .filter((x) => x.key && !isFilled((p as any)[x.key]))

  if (needFill.length > 0) {
    const first = needFill[0]
    const r = fieldRefByKey[String(first.key)]
    const inst = r?.value
    try {
      ;(inst as any)?.$el?.scrollIntoView?.({ behavior: 'smooth', block: 'center' })
    } catch {
      // ignore
    }
    ;(inst as any)?.startEdit?.()
    ElMessage.warning(`请先补齐：${needFill.map((x) => x.label).slice(0, 6).join('、')}${needFill.length > 6 ? '…' : ''}`)
    return
  }

  if (derivedUnconfirmed.value.length === 0) {
    ElMessage.success('已全部确认，无需处理')
    return
  }

  // 仅剩“无字段映射”的细项
  ElMessage.info(`还有 ${manualOnlyUnconfirmedCount.value} 项细节可手动确认`)
}

function openManualConfirmDialog() {
  confirmDialogVisible.value = true
}

function applyManualConfirm() {
  // 只允许“无字段映射”的项做手动确认；能映射字段的应靠补填字段自动消除
  const allow = manualOnlyUnconfirmed.value
  manualConfirmedItems.value = manualConfirmedItems.value.filter((x) => allow.includes(x))
  confirmDialogVisible.value = false
  if (derivedUnconfirmed.value.length === 0) {
    ElMessage.success('已完成确认')
  } else {
    ElMessage.success('已应用确认（仍有部分项需补填字段）')
  }
}

function resetExampleGroups() {
  exampleGroups.value = resetExampleGroupsToBuiltin()
  selectedGroupId.value = exampleGroups.value[0]?.id || ''
  ElMessage.success('已恢复内置分组')
}

function openGroupEditor(id?: string) {
  groupEditorId.value = id ?? null
  const g = id ? exampleGroups.value.find((x) => x.id === id) : null
  groupEditor.value = { name: g?.name || '', promptHint: g?.promptHint || '' }
  groupEditorVisible.value = true
}

function saveGroup() {
  const name = groupEditor.value.name.trim()
  const promptHint = groupEditor.value.promptHint.trim()
  if (!name) {
    ElMessage.warning('请填写分组名称')
    return
  }
  if (!promptHint) {
    ElMessage.warning('请填写分组提示词')
    return
  }
  if (groupEditorId.value) {
    const idx = exampleGroups.value.findIndex((x) => x.id === groupEditorId.value)
    if (idx >= 0) {
      exampleGroups.value[idx].name = name
      exampleGroups.value[idx].promptHint = promptHint
    }
  } else {
    const id = newId('grp')
    exampleGroups.value.unshift({ id, name, promptHint, examples: [] })
    selectedGroupId.value = id
  }
  groupEditorVisible.value = false
  ElMessage.success('分组已保存')
}

function onGroupCommand(cmd: string, id: string) {
  if (cmd === 'edit') {
    openGroupEditor(id)
    return
  }
  if (cmd === 'delete') {
    const idx = exampleGroups.value.findIndex((x) => x.id === id)
    if (idx >= 0) {
      exampleGroups.value.splice(idx, 1)
      selectedGroupId.value = exampleGroups.value[0]?.id || ''
      ElMessage.success('分组已删除')
    }
  }
}

function openExampleEditor(id?: string) {
  const g = selectedGroup.value
  if (!g) {
    ElMessage.warning('请先选择一个分组')
    return
  }
  exampleEditorId.value = id ?? null
  const ex = id ? g.examples.find((x) => x.id === id) : null
  exampleEditor.value = {
    title: ex?.title || '',
    lang: (ex?.lang as any) || 'en',
    content: ex?.content || '',
    promptHint: ex?.promptHint || '',
  }
  exampleEditorVisible.value = true
}

function saveExample() {
  const g = selectedGroup.value
  if (!g) return
  const title = exampleEditor.value.title.trim()
  const content = exampleEditor.value.content.trim()
  if (!title || !content) {
    ElMessage.warning('请填写标题与内容')
    return
  }
  const item: InquiryExampleItem = {
    id: exampleEditorId.value || newId('ex'),
    title,
    lang: exampleEditor.value.lang,
    content,
    promptHint: exampleEditor.value.promptHint.trim() || undefined,
  }
  if (exampleEditorId.value) {
    const idx = g.examples.findIndex((x) => x.id === exampleEditorId.value)
    if (idx >= 0) g.examples[idx] = item
  } else {
    g.examples.unshift(item)
  }
  exampleEditorVisible.value = false
  ElMessage.success('示例已保存')
}

function deleteExample(id: string) {
  const g = selectedGroup.value
  if (!g) return
  const idx = g.examples.findIndex((x) => x.id === id)
  if (idx >= 0) {
    g.examples.splice(idx, 1)
    ElMessage.success('示例已删除')
  }
}

function openAiGenDialog() {
  const g = selectedGroup.value
  if (!g) {
    ElMessage.warning('请先选择一个分组')
    return
  }
  if (!g.promptHint?.trim()) {
    ElMessage.warning('请先为该分组填写提示词')
    openGroupEditor(g.id)
    return
  }
  aiGenLang.value = detectedLang.value
  aiGenCount.value = 5
  aiGenExtraHint.value = ''
  aiGenVisible.value = true
}

async function runAiGenerateExamples() {
  const g = selectedGroup.value
  if (!g) return
  aiGenLoading.value = true
  try {
    const prompt = aiGenExtraHint.value.trim()
      ? `${g.promptHint}\n\n补充说明：${aiGenExtraHint.value.trim()}`
      : g.promptHint
    const created: any = await request.post('/ai/inquiry-example-jobs', {
      groupName: g.name,
      groupPrompt: prompt,
      language: aiGenLang.value,
      count: aiGenCount.value,
    })
    const jobId = Number(created.data?.jobId)
    if (!jobId) throw new Error('创建示例生成任务失败')
    const jobData: any = await pollJobGeneric(`/ai/inquiry-example-jobs/${jobId}`)
    const rows = (jobData?.examples || []) as Array<{ title: string; lang: 'zh' | 'en'; content: string }>
    if (!Array.isArray(rows) || rows.length === 0) {
      ElMessage.error('AI 未生成有效示例')
      return
    }
    const items: InquiryExampleItem[] = rows
      .map((x) => ({
        id: newId('ex'),
        title: String(x.title || '').trim() || '未命名示例',
        lang: (x.lang === 'zh' ? 'zh' : 'en') as any,
        content: String(x.content || '').trim(),
      }))
      .filter((x) => x.content)
    if (items.length === 0) {
      ElMessage.error('AI 未生成有效示例')
      return
    }
    g.examples.unshift(...items)
    aiGenVisible.value = false
    ElMessage.success(`已生成并添加 ${items.length} 条示例`)
  } catch (e) {
    console.error(e)
    ElMessage.error('生成失败，请检查网络或稍后重试')
  } finally {
    aiGenLoading.value = false
  }
}

function resetCompose() {
  composeAnswers.value = { product: '', spec: '', materialColor: '', quantity: '', delivery: '', other: '' }
}

async function composeToTextarea() {
  composeLoading.value = true
  try {
    const answers: Record<string, string> = {
      产品: composeAnswers.value.product,
      规格型号: composeAnswers.value.spec,
      材质颜色: composeAnswers.value.materialColor,
      数量: composeAnswers.value.quantity,
      交付地: composeAnswers.value.delivery,
      其他: composeAnswers.value.other,
    }
    const res: any = await request.post('/ai/compose-inquiry', {
      language: detectedLang.value,
      answers,
    })
    const content = String(res.data?.content ?? '')
    if (content.trim()) {
      requirementText.value = content.trim()
      composeVisible.value = false
      ElMessage.success('已生成并填入输入框')
    } else {
      ElMessage.error('生成结果为空，请补充信息后重试')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('生成失败，请检查网络或稍后重试')
  } finally {
    composeLoading.value = false
  }
}

function onItemsUpdate(items: Quote['items']) {
  if (!quoteResult.value) return
  quoteResult.value.items = items
  const sum = items.reduce((s, it) => s + (it.totalPrice || 0), 0)
  quoteResult.value.totalAmount = Math.round(sum * 100) / 100
}

let generatePollTimer: any = null
const attachmentPackLoading = ref(false)
/** 按附件在列表中的下标隔离「生成中」，避免同名文件或并行轮询互相干扰 */
const aiGeneratingAttachmentIndices = ref<Set<number>>(new Set())

function stopGeneratePolling() {
  if (generatePollTimer) {
    clearTimeout(generatePollTimer)
    generatePollTimer = null
  }
}

function sleep(ms: number) {
  return new Promise<void>((r) => setTimeout(r, ms))
}

/** 通用任务轮询：等待间隔使用独立 sleep，勿与全局 generatePollTimer 共用，以支持多任务并行 */
async function pollJobGeneric(url: string) {
  const start = Date.now()
  let interval = 1000
  while (true) {
    // eslint-disable-next-line no-await-in-loop
    const res: any = await request.get(url)
    const status = String(res.data?.status || '')
    if (status === 'succeeded') return res.data
    if (status === 'failed') throw new Error(String(res.data?.errorMsg || '任务失败'))

    const elapsed = Date.now() - start
    if (elapsed > 5 * 60 * 1000) throw new Error('任务超时（超过 5 分钟仍未完成）')
    if (elapsed > 10000 && interval < 2000) interval = 2000
    if (elapsed > 30000 && interval < 3000) interval = 3000
    if (elapsed > 60000 && interval < 5000) interval = 5000
    // eslint-disable-next-line no-await-in-loop
    await sleep(interval)
  }
}

async function pollGenerateJob(jobId: number) {
  const start = Date.now()
  let interval = 1000
  while (true) {
    // eslint-disable-next-line no-await-in-loop
    const res: any = await request.get(`/quotes/generate-jobs/${jobId}`)
    const status = String(res.data?.status || '')
    if (status === 'succeeded') {
      return res.data
    }
    if (status === 'failed') {
      throw new Error(String(res.data?.errorMsg || '生成失败'))
    }
    // queued/running
    const elapsed = Date.now() - start
    if (elapsed > 5 * 60 * 1000) {
      throw new Error('生成超时（超过 5 分钟仍未完成）')
    }
    // 退避：前 10s 每 1s，后面逐步降频
    if (elapsed > 10000 && interval < 2000) interval = 2000
    if (elapsed > 30000 && interval < 3000) interval = 3000
    if (elapsed > 60000 && interval < 5000) interval = 5000
    // eslint-disable-next-line no-await-in-loop
    await new Promise((r) => {
      stopGeneratePolling()
      generatePollTimer = setTimeout(r, interval)
    })
  }
}

async function handleAiGenerateSingleAttachment(
  att: { name: string; url?: string; selected: boolean; source?: string },
  rowIndex: number
) {
  if (!parsedParams.value || !quoteResult.value) {
    ElMessage.warning('请先生成报价后再生成附件')
    return
  }
  const name = String(att?.name || '').trim()
  if (!name) return
  if (!att.selected) {
    ElMessage.warning('请先勾选该附件再生成')
    return
  }
  if (String(att.url || '').trim()) return
  if (String(att.source || '') !== 'ai') return
  if (aiGeneratingAttachmentIndices.value.has(rowIndex)) return

  aiGeneratingAttachmentIndices.value = new Set(aiGeneratingAttachmentIndices.value).add(rowIndex)
  try {
    const created: any = await request.post('/quotes/attachment-generate-jobs', {
      params: parsedParams.value,
      quote: quoteResult.value,
      attachment: { name, selected: true, url: '', source: 'ai' },
    })
    const jobId = Number(created.data?.jobId)
    if (!jobId) throw new Error('创建附件生成任务失败')
    const jobData: any = await pollJobGeneric(`/quotes/attachment-generate-jobs/${jobId}`)
    const a = jobData?.attachment
    const url = String(a?.url || '').trim()
    if (url && quoteResult.value) {
      const list = quoteResult.value.attachments || []
      if (rowIndex >= 0 && rowIndex < list.length && String(list[rowIndex]?.name || '').trim() === name) {
        list[rowIndex] = { ...list[rowIndex], url, source: 'ai_generated' }
      } else {
        quoteResult.value.attachments = list.map((x: any) => {
          if (String(x?.name || '').trim() !== name) return x
          return { ...x, url, source: 'ai_generated' }
        })
      }
    }
    ElMessage.success('附件生成完成')
  } catch (e) {
    console.error(e)
    ElMessage.error('附件生成失败，请稍后重试')
  } finally {
    const next = new Set(aiGeneratingAttachmentIndices.value)
    next.delete(rowIndex)
    aiGeneratingAttachmentIndices.value = next
  }
}

async function handleGenerateEmailAttachmentPack() {
  if (!parsedParams.value || !quoteResult.value) {
    ElMessage.warning('请先生成报价后再生成附件包')
    return
  }
  if (attachmentPackLoading.value) return
  const selected = (quoteResult.value.attachments || []).filter((x: any) => x?.selected && String(x?.url || '').trim())
  if (selected.length === 0) {
    ElMessage.warning('请至少勾选 1 个“已存在文件”的附件（有 url 才能打包）')
    return
  }
  attachmentPackLoading.value = true
  try {
    const created: any = await request.post('/quotes/attachment-zip-jobs', { attachments: selected })
    const jobId = Number(created.data?.jobId)
    if (!jobId) throw new Error('创建附件任务失败')
    const jobData: any = await pollJobGeneric(`/quotes/attachment-zip-jobs/${jobId}`)
    const zipUrl = String(jobData?.zipUrl || '').trim()
    if (!zipUrl) throw new Error('未获取到 zipUrl')
    window.open(zipUrl, '_blank', 'noopener,noreferrer')
    ElMessage.success('附件包已生成，开始下载')
  } catch (e) {
    console.error(e)
    ElMessage.error('生成附件包失败，请稍后重试')
  } finally {
    attachmentPackLoading.value = false
  }
}

function handleDownloadAttachmentPack() {
  handleGenerateEmailAttachmentPack()
}

async function handleGenerate() {
  if (!parsedParams.value) {
    ElMessage.warning('请先解析客户需求')
    return
  }
  quoteStore.isLoading = true
  try {
    // 生成时记忆本次用户确认/编辑过的参数，便于下次快速选择与搜索（按最近更新时间排序）
    const p = parsedParams.value
    rememberValue('customerName', p.customerName)
    rememberValue('currency', p.currency)
    rememberValue('deliveryAddress', p.deliveryAddress)
    rememberValue('productName', p.productName)
    rememberValue('model', p.model)
    rememberValue('material', p.material)
    rememberValue('size', p.size)
    rememberValue('color', p.color)
    rememberValue('packaging', p.packaging)
    rememberValue('paymentTerms', p.paymentTerms)
    rememberValue('leadTime', p.leadTime)
    rememberValue('validityPeriod', p.validityPeriod)

    // 异步任务：先创建 job，再轮询状态，避免前端 30s 超时
    const created: any = await request.post('/quotes/generate-jobs', parsedParams.value)
    const jobId = Number(created.data?.jobId)
    if (!jobId) {
      throw new Error('创建生成任务失败')
    }
    const jobData = (await pollGenerateJob(jobId)) as Record<string, any>
    // 优先用后端已解析的 result；否则用 resultJson 兜底解析
    let rawResult = jobData?.result as Record<string, unknown> | null
    if (!rawResult && typeof jobData?.resultJson === 'string' && jobData.resultJson.trim()) {
      try {
        rawResult = JSON.parse(jobData.resultJson) as Record<string, unknown>
      } catch {
        rawResult = null
      }
    }
    if (!rawResult) {
      throw new Error('生成完成但未获取到结果数据')
    }
    quoteResult.value = mapGeneratePayload(rawResult, parsedParams.value)
    // AI 给出的 attachments 默认标记为 ai；并合并用户上传的真实附件（可预览）
    if (quoteResult.value) {
      quoteResult.value.attachments = (quoteResult.value.attachments || []).map((a: any) => ({
        ...a,
        source: a?.source || 'ai',
      }))
      for (const up of uploadedAttachments.value) {
        const exists = (quoteResult.value.attachments || []).some((x: any) => String(x.name).trim() === String(up.name).trim())
        if (!exists) {
          quoteResult.value.attachments.push(up as any)
        }
      }
    }
    ElMessage.success('报价生成完成')
  } catch (e) {
    console.error(e)
    ElMessage.error('生成失败，请检查网络或稍后重试')
  } finally {
    quoteStore.isLoading = false
  }
}

async function tryMatchProducts(params: ParsedParams) {
  try {
    const res = await productStore.matchProducts({
      productName: params.productName,
      material: params.material,
      size: params.size,
      color: params.color,
      model: params.model,
    })
    matchedProducts.value = res.products || []
    if (matchedProducts.value.length > 0) {
      ElMessage.success(`从产品库匹配到 ${matchedProducts.value.length} 个推荐产品`)
    }
  } catch {
    matchedProducts.value = []
  }
}

function handleClear() {
  requirementText.value = ''
  uploadFiles.value = []
  uploadedRefs.value = []
  parsedParams.value = null
  quoteResult.value = null
  matchedProducts.value = []
  draftQuoteId.value = null
  isDirty.value = false
  quoteStore.reset()
}

async function handleSaveDraft() {
  if (!quoteResult.value && !parsedParams.value) {
    ElMessage.warning('请先解析需求或生成报价')
    return
  }
  try {
    const saved = (await quoteStore.saveQuote(buildDraftPayload())) as { id?: number }
    if (saved?.id) {
      draftQuoteId.value = saved.id
      isDirty.value = false
      ElMessage.success('草稿已保存')
      await router.push(`/quotes/${saved.id}`)
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('保存失败，请稍后重试')
  }
}
</script>

<style scoped>
/* 右侧结果区：flex 子项默认 min-height:auto 会撑破布局，需 min-h-0 + 可滚动内容区，否则长话术/报价被裁切 */
.quote-result-tabs :deep(.el-tabs__header) {
  flex-shrink: 0;
}
.quote-result-tabs :deep(.el-tabs__content) {
  flex: 1 1 0%;
  min-height: 0;
  overflow: auto;
}
</style>
