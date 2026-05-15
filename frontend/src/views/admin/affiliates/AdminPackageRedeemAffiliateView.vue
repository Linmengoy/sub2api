<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="兑换码总数" :value="String(summary?.total_codes || 0)" />
        <StatCard label="未使用" :value="String(summary?.unused_codes || 0)" />
        <StatCard label="总销售额" :value="formatCurrency(summary?.total_purchase_pay_amount || 0)" />
        <StatCard label="已发返利" :value="formatCurrency(summary?.total_rebate_amount || 0)" highlight />
      </div>

      <div class="grid gap-4 sm:grid-cols-3">
        <StatCard label="已返利记录" :value="String(summary?.applied_rebate_count || 0)" />
        <StatCard label="待处理记录" :value="String(summary?.pending_rebate_count || 0)" />
        <StatCard label="失败记录" :value="String(summary?.failed_rebate_count || 0)" />
      </div>

      <div class="card p-6">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h1 class="text-lg font-semibold text-gray-900 dark:text-white">分销返利</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">查看用户购买的订阅兑换码，统计分销返利，并可删除未使用或异常兑换码。</p>
          </div>
          <div class="flex flex-col gap-2 sm:flex-row">
            <input v-model="filters.search" class="input sm:w-64" placeholder="搜索兑换码/用户邮箱" @input="debounceLoad" />
            <select v-model="filters.status" class="input sm:w-36" @change="loadCodes(1)">
              <option value="">全部状态</option>
              <option value="unused">未使用</option>
              <option value="used">已使用</option>
            </select>
            <button class="btn btn-secondary" :disabled="loading" @click="refreshAll">
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
              <span>刷新</span>
            </button>
          </div>
        </div>

        <div class="mt-5 overflow-x-auto">
          <table class="w-full min-w-[980px] text-left text-sm">
            <thead><tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400"><th class="px-3 py-2">ID</th><th class="px-3 py-2">兑换码</th><th class="px-3 py-2">购买者</th><th class="px-3 py-2">状态</th><th class="px-3 py-2">订阅天数</th><th class="px-3 py-2">实际支付</th><th class="px-3 py-2">使用者</th><th class="px-3 py-2">创建时间</th><th class="px-3 py-2 text-right">操作</th></tr></thead>
            <tbody>
              <tr v-for="code in codes" :key="code.id" class="border-b border-gray-100 last:border-b-0 dark:border-dark-800">
                <td class="px-3 py-3 text-gray-500">#{{ code.id }}</td>
                <td class="px-3 py-3 font-mono text-gray-900 dark:text-white">{{ code.code }}</td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">#{{ code.purchased_by || '-' }}</td>
                <td class="px-3 py-3"><span :class="['rounded-full px-2 py-1 text-xs font-medium', statusClass(code.status)]">{{ statusText(code.status) }}</span></td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ code.validity_days || '-' }} 天</td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatCurrency(code.purchase_pay_amount || 0) }}</td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ code.used_by ? `#${code.used_by}` : '-' }}</td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(code.created_at) }}</td>
                <td class="px-3 py-3 text-right"><button class="btn btn-danger btn-sm" :disabled="code.status === 'used' || deletingId === code.id" @click="deleteCode(code)">删除</button></td>
              </tr>
            </tbody>
          </table>
          <div v-if="!loading && codes.length === 0" class="py-8 text-center text-sm text-gray-500 dark:text-dark-400">暂无分销兑换码</div>
        </div>
        <div class="mt-4 flex items-center justify-between text-sm text-gray-500"><span>共 {{ pagination.total }} 条</span><div class="flex gap-2"><button class="btn btn-secondary btn-sm" :disabled="pagination.page <= 1" @click="loadCodes(pagination.page - 1)">上一页</button><button class="btn btn-secondary btn-sm" :disabled="pagination.page * pagination.page_size >= pagination.total" @click="loadCodes(pagination.page + 1)">下一页</button></div></div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { defineComponent, h, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import adminPackageRedeemAffiliateAPI, { type AdminPackageRedeemAffiliateSummary } from '@/api/admin/packageRedeemAffiliate'
import type { RedeemCode } from '@/types'
import { useAppStore } from '@/stores/app'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const appStore = useAppStore()
const summary = ref<AdminPackageRedeemAffiliateSummary | null>(null)
const codes = ref<RedeemCode[]>([])
const loading = ref(false)
const deletingId = ref<number | null>(null)
const filters = reactive({ search: '', status: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
let timer: ReturnType<typeof setTimeout> | null = null

function statusText(status: string) { return status === 'unused' ? '未使用' : status === 'used' ? '已使用' : status }
function statusClass(status: string) { return status === 'unused' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300' }
async function loadSummary() { summary.value = await adminPackageRedeemAffiliateAPI.getSummary() }
async function loadCodes(page = pagination.page) {
  loading.value = true
  try {
    const data = await adminPackageRedeemAffiliateAPI.listCodes({ page, page_size: pagination.page_size, status: filters.status, search: filters.search })
    codes.value = data.items || []
    pagination.page = data.page
    pagination.page_size = data.page_size
    pagination.total = data.total
  } catch (error) { appStore.showError(extractApiErrorMessage(error, '加载分销兑换码失败')) } finally { loading.value = false }
}
function debounceLoad() { if (timer) clearTimeout(timer); timer = setTimeout(() => void loadCodes(1), 350) }
async function refreshAll() { await Promise.all([loadSummary(), loadCodes(pagination.page)]) }
async function deleteCode(code: RedeemCode) {
  if (!window.confirm(`确认删除兑换码 ${code.code}？`)) return
  deletingId.value = code.id
  try { await adminPackageRedeemAffiliateAPI.delete(code.id); appStore.showSuccess('兑换码已删除'); await refreshAll() } catch (error) { appStore.showError(extractApiErrorMessage(error, '删除失败')) } finally { deletingId.value = null }
}

const StatCard = defineComponent({ props: { label: { type: String, required: true }, value: { type: String, required: true }, highlight: Boolean }, setup(props) { return () => h('div', { class: 'card p-5' }, [h('p', { class: 'text-sm text-gray-500 dark:text-dark-400' }, props.label), h('p', { class: ['mt-2 text-2xl font-semibold', props.highlight ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-900 dark:text-white'] }, props.value)]) } })

onMounted(() => { void refreshAll() })
</script>
