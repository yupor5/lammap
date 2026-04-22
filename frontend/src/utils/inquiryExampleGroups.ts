/**
 * 询盘示例：按分组管理，支持本地持久化（localStorage）。
 * 分组含「名称 + 提示词」：提示词用于 AI 按主题生成本组示例；单条示例也可带备注提示词。
 */

export type ExampleLang = 'zh' | 'en'

export interface InquiryExampleItem {
  id: string
  title: string
  lang: ExampleLang
  content: string
  /** 可选：本条示例的补充说明，给 AI 或同事理解场景用 */
  promptHint?: string
}

export interface InquiryExampleGroup {
  id: string
  name: string
  /** 本组主题说明：AI 生成本组多条示例时作为系统上下文 */
  promptHint: string
  examples: InquiryExampleItem[]
}

const STORAGE_KEY = 'lammap-inquiry-example-groups-v1'

export function newId(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

function clone<T>(x: T): T {
  return JSON.parse(JSON.stringify(x))
}

/** 内置默认分组（首次无本地数据时使用） */
export const BUILTIN_EXAMPLE_GROUPS: InquiryExampleGroup[] = [
  {
    id: 'builtin-metal',
    name: '金属与加工',
    promptHint:
      '生成面向外贸 B2B 的询盘：不锈钢/铝件/CNC/螺丝等金属五金类。语气像真实采购邮件，含数量、规格、材质、交付地、包装、付款/交期等要素中的若干项；不要虚构具体价格数字。',
    examples: [
      {
        id: 'ex-1',
        title: '不锈钢桌腿（英文，参数较全）',
        lang: 'en',
        content:
          'Hi,\n\nWe are looking for stainless steel table legs, 70cm height, black powder coating. Quantity: 500 pcs.\nPacking by carton, please quote FOB Shenzhen and CIF Los Angeles.\nPlease advise MOQ, lead time, and payment terms.\n\nThanks.',
      },
      {
        id: 'ex-5',
        title: 'CNC 铝件（英文，提供图纸）',
        lang: 'en',
        content:
          'Hi,\n\nWe need CNC machined aluminum parts (6061-T6), anodized black.\nWe can provide drawings (STEP/PDF). Quantity: 200 pcs for first batch.\nPlease quote unit price, lead time, and shipping to Hamburg, Germany.\n\nThank you.',
      },
      {
        id: 'ex-9',
        title: '不锈钢螺丝（英文，多规格）',
        lang: 'en',
        content:
          'Hi,\n\nWe need stainless steel screws, material: 304.\nSizes: M4x12 (5,000 pcs), M4x20 (5,000 pcs), M5x20 (3,000 pcs).\nPacking: 100 pcs/bag.\nPlease quote and advise lead time and payment terms.\n\nThanks.',
      },
    ],
  },
  {
    id: 'builtin-alu-general',
    name: '型材与通用件',
    promptHint:
      '生成工业铝型材、支架、结构件类 B2B 询盘。可故意留部分参数空缺，体现「待确认」场景；中英文均可。',
    examples: [
      {
        id: 'ex-2',
        title: '铝型材支架（中文，缺少部分参数）',
        lang: 'zh',
        content:
          '你好，我们需要采购一批铝型材支架，用于工业设备安装。\n数量大概 300 套。\n请报价并告知起订量、交期和付款方式。\n如果有不同规格/承重方案也请推荐。',
      },
    ],
  },
  {
    id: 'builtin-plastic-electronics',
    name: '塑料与电子',
    promptHint:
      '生成注塑件、LED、电子外壳类询盘。可含打样、模具、认证等合理要求；不要编造具体认证编号。',
    examples: [
      {
        id: 'ex-3',
        title: '塑料注塑外壳（英文，带打样）',
        lang: 'en',
        content:
          'Hello,\n\nWe need injection molded ABS plastic housings for an electronics device.\nSize approx: 120x80x35mm, color: black.\nFirst order 2,000 pcs, and we may need a prototype/sample before mass production.\nPlease quote unit price, mold cost (if any), lead time, and packaging.\n\nRegards.',
      },
      {
        id: 'ex-7',
        title: 'LED 灯带（英文，规格细）',
        lang: 'en',
        content:
          'Hello,\n\nWe are interested in LED strip lights.\nSpecs: 24V, 10mm width, 120 LEDs/m, 4000K, CRI>90, IP20.\nQuantity: 1,000 meters.\nPlease quote and confirm packaging, lead time, and warranty.\n\nBest regards.',
      },
    ],
  },
  {
    id: 'builtin-packaging',
    name: '包装与日用品',
    promptHint:
      '生成纸箱、玻璃杯、陶瓷等包装或日用品类询盘。可含印刷、logo、交付城市；语气自然。',
    examples: [
      {
        id: 'ex-4',
        title: '玻璃杯定制（中文，带 logo）',
        lang: 'zh',
        content:
          '您好，我们想定制玻璃杯，容量 350ml 左右，透明杯身。\n需要丝印 1 色 logo（我们可提供 AI 文件）。\n数量 5000 个，单个独立彩盒包装。\n请报 EXW/FOB 价格、打样时间、量产交期和起订量。',
      },
      {
        id: 'ex-6',
        title: '纸箱包装（中文，常规外贸）',
        lang: 'zh',
        content:
          '你好，我们需要定制外箱纸箱。\n尺寸：45*35*30cm（可微调），五层瓦楞，外箱印 2 色。\n数量 10000 个，交付到宁波。\n请报价并告知起订量、交期。',
      },
      {
        id: 'ex-8',
        title: '陶瓷花盆（中文，中英混合）',
        lang: 'zh',
        content:
          'Hi，我们想采购 ceramic flower pots。\n尺寸：直径 15cm / 20cm 两个规格，颜色白色+哑光。\n数量：各 2000 个。\n请报 FOB 价格、包装方式（是否可加泡沫保护）、交期和 MOQ。',
      },
    ],
  },
  {
    id: 'builtin-textile',
    name: '纺织材料',
    promptHint:
      '生成面料、纺织品类询盘。关注克重、门幅、功能（防水等）、颜色、米数/码数；B2B 外贸口吻。',
    examples: [
      {
        id: 'ex-10',
        title: '纺织面料（中文，关注克重）',
        lang: 'zh',
        content:
          '您好，我们在找一款涤纶面料用于户外背包。\n要求：600D，克重约 240gsm，防泼水，颜色黑色。\n数量先做 2000 米。\n请报价（含运费到上海的方案也可）、交期、付款条款和样布安排。',
      },
    ],
  },
]

export function loadExampleGroups(): InquiryExampleGroup[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw) as InquiryExampleGroup[]
      if (Array.isArray(parsed) && parsed.length > 0) return parsed
    }
  } catch {
    // ignore
  }
  return clone(BUILTIN_EXAMPLE_GROUPS)
}

export function saveExampleGroups(groups: InquiryExampleGroup[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(groups))
  } catch {
    // ignore
  }
}

export function resetExampleGroupsToBuiltin() {
  localStorage.removeItem(STORAGE_KEY)
  return clone(BUILTIN_EXAMPLE_GROUPS)
}
