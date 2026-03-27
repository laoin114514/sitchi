import type { LoginData, LoginPayload } from '@/api/types'
import { postJson } from '@/utils/request'

export function login(payload: LoginPayload) {
  return postJson<LoginData, LoginPayload>('/login', payload)
}
