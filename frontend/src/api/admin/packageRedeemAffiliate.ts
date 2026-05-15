import { apiClient } from '../client'
import type { BasePaginationResponse, RedeemCode } from '@/types'

export interface AdminPackageRedeemAffiliateSummary {
  total_codes: number
  unused_codes: number
  used_codes: number
  total_purchase_pay_amount: number
  total_rebate_amount: number
  applied_rebate_count: number
  pending_rebate_count: number
  failed_rebate_count: number
}

export interface ListAdminPackageRedeemCodesParams {
  page?: number
  page_size?: number
  status?: string
  search?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export async function getSummary(): Promise<AdminPackageRedeemAffiliateSummary> {
  const { data } = await apiClient.get<AdminPackageRedeemAffiliateSummary>('/admin/affiliates/package-redeem/summary')
  return data
}

export async function listCodes(params: ListAdminPackageRedeemCodesParams = {}): Promise<BasePaginationResponse<RedeemCode>> {
  const { data } = await apiClient.get<BasePaginationResponse<RedeemCode>>('/admin/affiliates/package-redeem/codes', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 20,
      status: params.status || undefined,
      search: params.search || undefined,
      sort_by: params.sort_by || undefined,
      sort_order: params.sort_order || undefined,
    },
  })
  return data
}

export async function getCode(id: number): Promise<RedeemCode> {
  const { data } = await apiClient.get<RedeemCode>(`/admin/affiliates/package-redeem/codes/${id}`)
  return data
}

export async function deleteCode(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/affiliates/package-redeem/codes/${id}`)
  return data
}

export const adminPackageRedeemAffiliateAPI = {
  getSummary,
  listCodes,
  getCode,
  delete: deleteCode,
}

export default adminPackageRedeemAffiliateAPI
