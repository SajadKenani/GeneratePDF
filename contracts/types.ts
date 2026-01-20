import type salesIraq from "../contracts/data/iraq/sales.json"
import type careIraq from "../contracts/data/iraq/care.json"
import type rentalIraq from "../contracts/data/iraq/rental.json"
import type authIraq from "../contracts/data/iraq/auth.json"

import type salesRomain from "../contracts/data/romain/hen/sales.json"
import type careRomain from "../contracts/data/romain/hen/care.json"
import type rentalRomain from "../contracts/data/romain/hen/rental.json"
import type authRomain from "../contracts/data/romain/hen/auth.json"
import type cattleRomain from "../contracts/data/romain/cattle/cattle.json"

import type withdrawShared from "../contracts/data/shared/withdraw.json"

export interface AlTaerEmailProps {
  createdAt: Date
  depositNumId: number
  realm: string
  currency: string
  currencyAmount: number
  paymentMethod: string
  itemAmount: number
  item: string
  userEmail: string
  userName?: string | null
  userPhone?: string | null
  insurance?: boolean | null
  contentType?: "sales" | "care" | "rental" | "auth" | "withdraw" | "cattle"
  content?: typeof salesIraq | typeof careIraq | typeof rentalIraq | typeof authIraq | typeof salesRomain | typeof careRomain | typeof rentalRomain | typeof authRomain | typeof withdrawShared | typeof cattleRomain
}

export interface AttachmentProps {
  filename?: string
  content?: Buffer | string
  contentType?: string
  encoding?: string
}

export interface ContractData {
  numId: number
  createdAt: Date
  realm: string
  type: string
  currency: string
  currencyAmount: number
  item: string
  itemAmount: number
  paymentMethod: string
  valueInUsd: number
  insurance: boolean
  userEmail: string
  userName: string | null
  userPhone: string | null
  content?: typeof salesIraq | typeof careIraq | typeof rentalIraq | typeof authIraq | typeof salesRomain | typeof careRomain | typeof rentalRomain | typeof authRomain | typeof withdrawShared | typeof cattleRomain
}

export const PDF_OPTIONS = {
  format: "A4" as const,
  printBackground: true,
  displayHeaderFooter: true,
  headerTemplate: "<div></div>",
  footerTemplate: "<div></div>",
  margin: {
    top: "0mm",
    right: "0mm",
    bottom: "0mm",
    left: "0mm",
  },
  timeout: 25000,
  preferCSSPageSize: true,
}