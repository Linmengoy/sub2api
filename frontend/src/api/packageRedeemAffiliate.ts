import { apiClient } from './client'
import type { BasePaginationResponse, RedeemCode } from '@/types'

export interface PackageRedeemAffiliateSummary {
  default_rebate_rate_percent: number
  total_codes: number
  unused_codes: number
  used_codes: number
  total_purchase_pay_amount: number
  total_rebate_amount: number
  available_quota: number
  frozen_quota: number
  history_quota: number
}

export interface ListPackageRedeemCodesParams {
  page?: number
  page_size?: number
  status?: string
}

export async function getSummary(): Promise<PackageRedeemAffiliateSummary> {
  const { data } = await apiClient.get<PackageRedeemAffiliateSummary>('/package-redeem-affiliate/summary')
  return data
}

export async function listCodes(params: ListPackageRedeemCodesParams = {}): Promise<BasePaginationResponse<RedeemCode>> {
  const { data } = await apiClient.get<BasePaginationResponse<RedeemCode>>('/package-redeem-affiliate/codes', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 20,
      status: params.status || undefined,
    },
  })
  return data
}

export const packageRedeemAffiliateAPI = {
  getSummary,
  listCodes,
}

export default packageRedeemAffiliateAPI
