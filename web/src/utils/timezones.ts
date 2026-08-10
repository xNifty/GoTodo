/** Former allowlist — used when Intl.supportedValuesOf is unavailable. */
const FALLBACK_TIMEZONES = [
  'UTC',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'America/Anchorage',
  'America/Phoenix',
  'America/Toronto',
  'America/Vancouver',
  'America/Mexico_City',
  'America/Sao_Paulo',
  'America/Buenos_Aires',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Europe/Amsterdam',
  'Europe/Madrid',
  'Europe/Rome',
  'Europe/Stockholm',
  'Europe/Moscow',
  'Asia/Dubai',
  'Asia/Kolkata',
  'Asia/Bangkok',
  'Asia/Shanghai',
  'Asia/Hong_Kong',
  'Asia/Singapore',
  'Asia/Tokyo',
  'Asia/Seoul',
  'Australia/Sydney',
  'Australia/Melbourne',
  'Australia/Perth',
  'Pacific/Auckland',
  'Pacific/Honolulu',
  'Africa/Cairo',
  'Africa/Johannesburg',
]

/** Sorted IANA timezone IDs from the browser, or a curated fallback. */
export function listTimezones(): string[] {
  try {
    const supported = (Intl as unknown as { supportedValuesOf?: (key: string) => string[] }).supportedValuesOf
    if (typeof supported === 'function') {
      const zones = supported.call(Intl, 'timeZone')
      if (Array.isArray(zones) && zones.length > 0) {
        return [...zones].sort((a, b) => a.localeCompare(b))
      }
    }
  } catch {
    // fall through
  }
  return [...FALLBACK_TIMEZONES]
}

/** Ensure a stored value appears in the option list even if absent from Intl. */
export function ensureTimezoneOption(list: string[], current: string | null | undefined): string[] {
  const tz = (current || '').trim()
  if (!tz || list.includes(tz)) return list
  return [...list, tz].sort((a, b) => a.localeCompare(b))
}

export type TimezoneGroup = { label: string; zones: string[] }

/** Group IANA IDs by region prefix (America/… → America); bare names → General. */
export function groupTimezones(zones: string[]): TimezoneGroup[] {
  const byRegion = new Map<string, string[]>()
  for (const zone of zones) {
    const slash = zone.indexOf('/')
    const region = slash === -1 ? 'General' : zone.slice(0, slash)
    const bucket = byRegion.get(region)
    if (bucket) bucket.push(zone)
    else byRegion.set(region, [zone])
  }
  const labels = [...byRegion.keys()].sort((a, b) => {
    if (a === 'General') return -1
    if (b === 'General') return 1
    return a.localeCompare(b)
  })
  return labels.map((label) => ({ label, zones: byRegion.get(label)! }))
}

/** Current UTC offset label for a zone, e.g. "(UTC - 5)" or "(UTC + 5:30)". */
export function formatUtcOffset(zone: string, at: Date = new Date()): string {
  try {
    const parts = new Intl.DateTimeFormat('en-US', {
      timeZone: zone,
      timeZoneName: 'shortOffset',
    }).formatToParts(at)
    const raw = parts.find((p) => p.type === 'timeZoneName')?.value ?? ''
    if (/^(GMT|UTC)$/i.test(raw)) return '(UTC)'
    const match = raw.match(/(?:GMT|UTC)([+-])(\d{1,2})(?::(\d{2}))?/i)
    if (!match) return '(UTC)'
    const sign = match[1] === '+' ? '+' : '-'
    const hours = Number(match[2])
    const minutes = match[3] ? Number(match[3]) : 0
    if (hours === 0 && minutes === 0) return '(UTC)'
    if (minutes === 0) return `(UTC ${sign} ${hours})`
    return `(UTC ${sign} ${hours}:${String(minutes).padStart(2, '0')})`
  } catch {
    return ''
  }
}

/** Display label: "America/New_York (UTC - 5)". */
export function timezoneLabel(zone: string, at: Date = new Date()): string {
  const offset = formatUtcOffset(zone, at)
  return offset ? `${zone} ${offset}` : zone
}
