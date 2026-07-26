/** sessionStorage key for resuming app browser SSO after login/claim. */
export const DEVICE_AUTH_RETURN_KEY = 'gotodo.deviceAuthReturn'

/**
 * True for in-app paths to the device approve screen (optional query).
 * Rejects absolute/external URLs — mirrors server SafeDeviceReturnTo intent.
 */
export function isDeviceAuthPath(path: string | null | undefined): boolean {
  if (!path || typeof path !== 'string') return false
  const trimmed = path.trim()
  if (!trimmed.startsWith('/') || trimmed.startsWith('//')) return false
  if (/^[a-z][a-z0-9+.-]*:/i.test(trimmed)) return false
  const pathname = trimmed.split(/[?#]/)[0] || ''
  return pathname === '/auth/device' || pathname.startsWith('/auth/device/')
}

export function stashDeviceAuthReturn(path: string | null | undefined): void {
  if (!path || !isDeviceAuthPath(path)) return
  try {
    sessionStorage.setItem(DEVICE_AUTH_RETURN_KEY, path)
  } catch {
    // Ignore quota / private-mode failures.
  }
}

export function peekDeviceAuthReturn(): string | null {
  try {
    const stored = sessionStorage.getItem(DEVICE_AUTH_RETURN_KEY)
    if (!stored || !isDeviceAuthPath(stored)) return null
    return stored
  } catch {
    return null
  }
}

export function takeDeviceAuthReturn(): string | null {
  const stored = peekDeviceAuthReturn()
  if (!stored) return null
  try {
    sessionStorage.removeItem(DEVICE_AUTH_RETURN_KEY)
  } catch {
    // ignore
  }
  return stored
}

/** Prefer a safe device redirect from query, then stash, else fallback. */
export function resolvePostLoginRedirect(
  queryRedirect: string | null | undefined,
  fallback = '/',
): string {
  if (queryRedirect && isDeviceAuthPath(queryRedirect)) {
    return queryRedirect
  }
  const stashed = peekDeviceAuthReturn()
  if (stashed) return stashed
  if (typeof queryRedirect === 'string' && queryRedirect.startsWith('/') && !queryRedirect.startsWith('//')) {
    return queryRedirect
  }
  return fallback
}
