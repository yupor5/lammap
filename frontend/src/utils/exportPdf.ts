import jsPDF from 'jspdf'
import html2canvas from 'html2canvas'

export interface PdfQuoteData {
  quoteNumber?: string
  customerName: string
  country?: string
  currency: string
  deliveryAddress?: string
  leadTime?: string
  remarks?: string
  terms?: string
  items: {
    productName: string
    model: string
    specs: string
    quantity: number
    unitPrice: number
    totalPrice: number
  }[]
  totalAmount: number
}

export function exportQuotePdf(data: PdfQuoteData) {
  // 说明：jsPDF 默认字体不支持中文，直接 text/autoTable 会乱码。
  // 这里采用“HTML 渲染 → html2canvas → 写入 PDF 图片”的方式，保证中文不乱码。
  const container = document.createElement('div')
  container.style.position = 'fixed'
  container.style.left = '-10000px'
  container.style.top = '0'
  container.style.width = '794px' // 约等于 A4 在 96dpi 下的宽度
  container.style.padding = '24px'
  container.style.background = '#ffffff'
  container.style.color = '#111827'
  container.style.fontFamily =
    'ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, "Noto Sans", "PingFang SC", "Microsoft YaHei", sans-serif'

  const safe = (s: unknown) => String(s ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')

  const rows = data.items
    .map((it, i) => {
      return `
        <tr>
          <td style="padding:8px;border:1px solid #e5e7eb;text-align:center;">${i + 1}</td>
          <td style="padding:8px;border:1px solid #e5e7eb;">${safe(it.productName)}</td>
          <td style="padding:8px;border:1px solid #e5e7eb;">${safe(it.model)}</td>
          <td style="padding:8px;border:1px solid #e5e7eb;">${safe(it.specs)}</td>
          <td style="padding:8px;border:1px solid #e5e7eb;text-align:right;">${it.quantity}</td>
          <td style="padding:8px;border:1px solid #e5e7eb;text-align:right;">${safe(data.currency)} ${it.unitPrice.toFixed(2)}</td>
          <td style="padding:8px;border:1px solid #e5e7eb;text-align:right;">${safe(data.currency)} ${it.totalPrice.toFixed(2)}</td>
        </tr>
      `
    })
    .join('')

  const terms = data.terms
    ? `<div style="margin-top:18px;font-size:12px;color:#111827;">
        <div style="font-weight:700;margin-bottom:6px;">条款 / Terms</div>
        <div style="white-space:pre-wrap;line-height:1.6;color:#374151;">${safe(data.terms)}</div>
      </div>`
    : ''

  const remarks = data.remarks
    ? `<div style="margin-top:14px;font-size:12px;color:#111827;">
        <div style="font-weight:700;margin-bottom:6px;">备注 / Remarks</div>
        <div style="white-space:pre-wrap;line-height:1.6;color:#374151;">${safe(data.remarks)}</div>
      </div>`
    : ''

  container.innerHTML = `
    <div style="text-align:center;font-size:22px;font-weight:800;letter-spacing:0.08em;margin-bottom:10px;">QUOTATION</div>
    <div style="display:flex;justify-content:space-between;font-size:12px;color:#6b7280;margin-bottom:14px;">
      <div>${data.quoteNumber ? `Quote #: ${safe(data.quoteNumber)}` : ''}</div>
      <div>Date: ${safe(new Date().toLocaleDateString('en-US'))}</div>
    </div>

    <div style="background:#f9fafb;border:1px solid #f3f4f6;border-radius:10px;padding:12px;font-size:12px;color:#111827;">
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px 16px;">
        <div><span style="color:#6b7280;">Customer：</span>${safe(data.customerName)}</div>
        <div><span style="color:#6b7280;">Currency：</span>${safe(data.currency)}</div>
        <div><span style="color:#6b7280;">Country：</span>${safe(data.country || '-')}</div>
        <div><span style="color:#6b7280;">Delivery To：</span>${safe(data.deliveryAddress || '-')}</div>
        ${data.leadTime ? `<div><span style="color:#6b7280;">Lead Time：</span>${safe(data.leadTime)}</div>` : ''}
      </div>
    </div>

    <div style="margin-top:14px;">
      <table style="width:100%;border-collapse:collapse;font-size:12px;">
        <thead>
          <tr style="background:#3b82f6;color:#fff;">
            <th style="padding:8px;border:1px solid #3b82f6;width:36px;">#</th>
            <th style="padding:8px;border:1px solid #3b82f6;">Product</th>
            <th style="padding:8px;border:1px solid #3b82f6;width:90px;">Model</th>
            <th style="padding:8px;border:1px solid #3b82f6;">Specs</th>
            <th style="padding:8px;border:1px solid #3b82f6;width:60px;text-align:right;">Qty</th>
            <th style="padding:8px;border:1px solid #3b82f6;width:110px;text-align:right;">Unit Price</th>
            <th style="padding:8px;border:1px solid #3b82f6;width:110px;text-align:right;">Total</th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
        <tfoot>
          <tr style="background:#f3f4f6;font-weight:700;">
            <td style="padding:8px;border:1px solid #e5e7eb;"></td>
            <td style="padding:8px;border:1px solid #e5e7eb;" colspan="4"></td>
            <td style="padding:8px;border:1px solid #e5e7eb;text-align:right;">TOTAL</td>
            <td style="padding:8px;border:1px solid #e5e7eb;text-align:right;">${safe(data.currency)} ${data.totalAmount.toFixed(2)}</td>
          </tr>
        </tfoot>
      </table>
    </div>

    ${terms}
    ${remarks}
  `

  document.body.appendChild(container)

  const filename = data.quoteNumber
    ? `Quotation_${data.quoteNumber}.pdf`
    : `Quotation_${data.customerName.replace(/\s+/g, '_')}_${new Date().toISOString().slice(0, 10)}.pdf`

  html2canvas(container, { scale: 2, backgroundColor: '#ffffff', useCORS: true })
    .then((canvas) => {
      const imgData = canvas.toDataURL('image/png')
      const pdf = new jsPDF({ unit: 'pt', format: 'a4' })
      const pageWidth = pdf.internal.pageSize.getWidth()
      const pageHeight = pdf.internal.pageSize.getHeight()

      const imgWidth = pageWidth
      const imgHeight = (canvas.height * imgWidth) / canvas.width

      let y = 0
      let remaining = imgHeight
      let page = 0

      while (remaining > 0) {
        if (page > 0) pdf.addPage()
        pdf.addImage(imgData, 'PNG', 0, y, imgWidth, imgHeight)
        remaining -= pageHeight
        y -= pageHeight
        page++
      }

      pdf.save(filename)
    })
    .finally(() => {
      container.remove()
    })
}
