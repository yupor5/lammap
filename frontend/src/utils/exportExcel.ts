import * as XLSX from 'xlsx'
import { saveAs } from 'file-saver'

export interface ExcelQuoteData {
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

export function exportQuoteExcel(data: ExcelQuoteData) {
  const wb = XLSX.utils.book_new()

  const headerRows: (string | number)[][] = [
    ['QUOTATION'],
    [],
    ['Quote #', data.quoteNumber || '-', '', 'Date', new Date().toLocaleDateString('en-US')],
    ['Customer', data.customerName, '', 'Country', data.country || '-'],
    ['Currency', data.currency, '', 'Delivery To', data.deliveryAddress || '-'],
  ]
  if (data.leadTime) {
    headerRows.push(['Lead Time', data.leadTime])
  }
  headerRows.push([])

  const tableHeader = ['#', 'Product', 'Model', 'Specs', 'Qty', 'Unit Price', 'Total']
  const tableRows = data.items.map((item, i) => [
    i + 1,
    item.productName,
    item.model,
    item.specs,
    item.quantity,
    item.unitPrice,
    item.totalPrice,
  ])
  const totalRow = ['', '', '', '', '', 'TOTAL', data.totalAmount]

  const allRows = [
    ...headerRows,
    tableHeader,
    ...tableRows,
    totalRow,
  ]

  if (data.terms) {
    allRows.push([], ['Terms & Conditions:', data.terms])
  }
  if (data.remarks) {
    allRows.push([], ['Remarks:', data.remarks])
  }

  const ws = XLSX.utils.aoa_to_sheet(allRows)

  ws['!cols'] = [
    { wch: 5 },
    { wch: 25 },
    { wch: 15 },
    { wch: 25 },
    { wch: 8 },
    { wch: 14 },
    { wch: 14 },
  ]

  XLSX.utils.book_append_sheet(wb, ws, 'Quotation')

  const buf = XLSX.write(wb, { bookType: 'xlsx', type: 'array' })
  const filename = data.quoteNumber
    ? `Quotation_${data.quoteNumber}.xlsx`
    : `Quotation_${data.customerName.replace(/\s+/g, '_')}_${new Date().toISOString().slice(0, 10)}.xlsx`

  saveAs(new Blob([buf], { type: 'application/octet-stream' }), filename)
}

export interface ProductImportRow {
  name: string
  sku: string
  category: string
  material: string
  size: string
  color: string
  process: string
  packaging: string
  price: number
  moq: number
  leadTime: string
  paymentTerms: string
  description: string
}

const COLUMN_MAP: Record<string, keyof ProductImportRow> = {
  '产品名称': 'name', 'name': 'name', 'product name': 'name', 'product': 'name',
  'sku': 'sku', '型号': 'sku', 'model': 'sku',
  '分类': 'category', 'category': 'category',
  '材质': 'material', 'material': 'material',
  '尺寸': 'size', 'size': 'size',
  '颜色': 'color', 'color': 'color',
  '工艺': 'process', 'process': 'process',
  '包装': 'packaging', 'packaging': 'packaging', 'packing': 'packaging',
  '价格': 'price', 'price': 'price', '参考价格': 'price', 'unit price': 'price',
  'moq': 'moq', '起订量': 'moq', 'min order': 'moq',
  '交期': 'leadTime', 'lead time': 'leadTime', 'leadtime': 'leadTime',
  '付款方式': 'paymentTerms', 'payment': 'paymentTerms', 'payment terms': 'paymentTerms',
  '描述': 'description', 'description': 'description', '简介': 'description',
}

export function parseProductExcel(file: File): Promise<ProductImportRow[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      try {
        const data = new Uint8Array(e.target!.result as ArrayBuffer)
        const wb = XLSX.read(data, { type: 'array' })
        const ws = wb.Sheets[wb.SheetNames[0]]
        const rawRows = XLSX.utils.sheet_to_json<Record<string, any>>(ws)

        if (rawRows.length === 0) {
          reject(new Error('Excel 文件为空'))
          return
        }

        const headerMap: Record<string, keyof ProductImportRow> = {}
        const firstRow = rawRows[0]
        for (const key of Object.keys(firstRow)) {
          const normalized = key.trim().toLowerCase()
          if (COLUMN_MAP[normalized]) {
            headerMap[key] = COLUMN_MAP[normalized]
          }
        }

        const products: ProductImportRow[] = rawRows.map((row) => {
          const product: ProductImportRow = {
            name: '', sku: '', category: '', material: '',
            size: '', color: '', process: '', packaging: '',
            price: 0, moq: 0, leadTime: '', paymentTerms: '', description: '',
          }
          for (const [excelCol, field] of Object.entries(headerMap)) {
            const val = row[excelCol]
            if (val !== undefined && val !== null) {
              if (field === 'price' || field === 'moq') {
                ;(product as any)[field] = Number(val) || 0
              } else {
                ;(product as any)[field] = String(val)
              }
            }
          }
          return product
        }).filter((p) => p.name.trim() !== '')

        resolve(products)
      } catch (err: any) {
        reject(new Error('Excel 解析失败: ' + err.message))
      }
    }
    reader.onerror = () => reject(new Error('文件读取失败'))
    reader.readAsArrayBuffer(file)
  })
}
