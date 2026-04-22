// 在浏览器环境里把 pinyin-pro 挂到 globalThis，避免 paramMemory.ts 引入 Node 的 require 类型。
import { pinyin } from 'pinyin-pro'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
(globalThis as any).__lammapPinyinPro = { pinyin }

