export interface TemplateContext {
  [k: string]: any
}

/**
 * 极简渲染：
 * - 支持 {{var}} 替换（嵌套用点号：{{quote.customer_name}}）
 * - 支持缺省：不存在则替换为空字符串
 *
 * 备注：后续要更强（循环/条件）可以换 Mustache/Handlebars；这里先满足你当前模板管理的需求。
 */
export function renderTemplate(template: string, ctx: TemplateContext) {
  const get = (path: string) => {
    const parts = path.split('.').map((x) => x.trim()).filter(Boolean)
    let cur: any = ctx
    for (const p of parts) {
      if (cur == null) return ''
      cur = cur[p]
    }
    if (cur == null) return ''
    return String(cur)
  }

  return String(template || '').replace(/\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}/g, (_m, key) => get(String(key)))
}

