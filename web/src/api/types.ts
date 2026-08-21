export type User = {
  id: number
  email: string
  user_name: string
  timezone: string
  items_per_page: number
  permissions: string[]
  digest_enabled: boolean
  digest_hour: number
  allow_project_invites: boolean
        username_change_available: boolean
}

export type UserSearchHit = {
  user_name: string
}

export type Tag = {
  id: number
  name: string
  color: string
  project_id?: number | null
}

export type WorkflowMode = 'classic' | 'kanban'

export type Project = {
  id: number
  name: string
  description?: string
  workflow_mode?: WorkflowMode
  role?: 'owner' | 'editor' | 'viewer'
  owner_email?: string
  owner_user_name?: string
  owner_user_id?: number
}

export type ProjectStatus = {
  id: number
  project_id: number
  name: string
  description?: string
  position: number
  is_done: boolean
  is_default: boolean
  created_at: string
}

export type TaskTimeEntry = {
  id: number
  task_id: number
  user_id: number
  minutes: number
  note: string
  created_at: string
  user_email?: string
  user_name?: string
}

export type TaskCommentLink = {
  id: number
  title: string
}

export type TaskComment = {
  id: number
  task_id: number
  user_id: number
  user_name?: string
  body: string
  created_at: string
  deleted: boolean
  deleted_at?: string | null
  deleted_by_user_id?: number
  deleted_by_kind?: 'user' | 'owner' | string
  links?: TaskCommentLink[]
}

export type ProjectMember = {
  user_id: number
  email: string
  user_name: string
  role: 'owner' | 'editor' | 'viewer'
  created_at: string
}

export type ProjectInvite = {
  id: number
  project_id: number
  email: string
  user_name?: string
  role: 'editor' | 'viewer'
  expires_at: string
  created_at: string
  project_name?: string
  inviter_email?: string
  inviter_user_name?: string
}

export type ProjectEvent = {
  id: number
  project_id: number
  actor_user_id: number
  actor_email?: string
  actor_user_name?: string
  event_type: string
  source: 'project' | 'task'
  task_id?: number
  label: string
  metadata?: Record<string, unknown>
  created_at: string
}

export type ShareLink = {
  id: number
  token: string
  url: string
  scope_type: 'project'
  scope_id: number
  expires_at?: string | null
  created_at: string
}

export type ShareLinkTask = {
  id: number
  title: string
  description?: string
  completed: boolean
  due_date: string
  priority: number
  project?: string
  tags?: Tag[]
  status_name?: string
}

export type ShareLinkView = {
  scope_type: string
  scope_id: number
  tasks: ShareLinkTask[]
}

export type TaskGitHubIssue = {
  issue_number: number
  issue_id: number
  issue_url: string
  issue_state: string
  issue_title?: string
  last_sync_error?: string
}

export type Task = {
  id: number
  title: string
  description: string
  completed: boolean
  due_date: string
  project_id?: number | null
  project?: string
  priority: number
  favorite: boolean
  position: number
  parent_id?: number | null
  child_count?: number
  children_completed?: number
  children?: Task[]
  tags: Tag[]
  created_at: string
  modified_at: string
  status_id?: number | null
  status_name?: string
  estimate_points?: number | null
  time_spent_minutes?: number
  project_workflow?: WorkflowMode | string
  claimed_by?: number | null
  claimed_by_name?: string
  github?: TaskGitHubIssue | null
}

export type GitHubConnection = {
  connected: boolean
  github_login?: string
  auth_method?: 'oauth' | 'pat' | string
  connected_at?: string
}

export type ProjectGitHubRepo = {
  linked: boolean
  owner?: string
  repo?: string
  full_name?: string
  html_url?: string
  repo_id?: number
  linked_by_user_id?: number
  webhook_secret?: string
  linked_at?: string
}

export type Notification = {
  id: number
  type: string
  title: string
  body: string
  project_id?: number | null
  task_id?: number | null
  project_name?: string
  actor_name?: string
  read_at?: string | null
  created_at: string
}

export type NotificationList = {
  notifications: Notification[]
  total: number
  page: number
  per_page: number
  unread_count: number
}

export type TaskEvent = {
  id: number
  task_id: number
  event_type: string
  label: string
  metadata?: Record<string, unknown>
  created_at: string
  actor_user_id?: number
  actor_user_name?: string
  actor_email?: string
}

export type TaskList = {
  tasks: Task[]
  total: number
  page: number
  per_page: number
  total_pages: number
  completed_count: number
  incomplete_count: number
}

export type SiteInfo = {
  site_name: string
  show_changelog: boolean
  enable_registration?: boolean
  invite_only?: boolean
  enable_join_requests?: boolean
  meta_description?: string
  enable_global_announcement: boolean
  global_announcement_text: string
  announcement_dismissed: boolean
  github_oauth_configured?: boolean
}

export type ChangelogEntry = {
  version: string
  date: string
  title: string
  notes: string[]
  html?: string
  prerelease?: boolean
}

export type SavedViewFilter = {
  project?: string
  status?: string
  due?: string
  completed?: string
  priority?: string
  tag?: string
  sort?: string
  search?: string
}

export type SavedView = {
  id: number
  name: string
  filter: SavedViewFilter
  sort_order: number
  created_at: string
  updated_at: string
}

export type DashboardStats = {
  overdue_count: number
  due_today_count: number
  due_this_week_count: number
  completed_this_week: number
  completed_this_month: number
  streak_days: number
  by_project: { name: string; count: number }[]
  by_priority: { priority: number; label: string; count: number }[]
  completions_last_7_days: { date: string; count: number }[]
}

export type CalendarMonthTask = {
  id: number
  title: string
  due: string
  priority: number
  project_name: string
  completed: boolean
}

export type CalendarMonthCell = {
  date: string
  day: number
  in_month: boolean
  is_today: boolean
  tasks: CalendarMonthTask[]
}

export type CalendarMonth = {
  year_month: string
  month_label: string
  prev_month: string
  next_month: string
  today_month: string
  year: number
  weeks: CalendarMonthCell[][]
}

export type CalendarInfo = {
  token: string
  feed_url: string
}

export type Invite = {
  id: number
  email: string
  token: string
  used: boolean
}

export type JoinRequest = {
  id: number
  email: string
  message: string
  status: 'pending' | 'approved' | 'denied'
  created_at: string
  invite_id?: number
  reviewed_at?: string
  reviewed_by?: number
  invite_token?: string
}

export type AdminSettings = {
  site_name: string
  default_timezone: string
  show_changelog: boolean
  site_version: string
  enable_registration: boolean
  invite_only: boolean
  enable_join_requests: boolean
  meta_description: string
  enable_global_announcement: boolean
  global_announcement_text: string
  enable_api: boolean
  email_provider: string
  email_from_address: string
  email_from_name: string
  email_mailgun_domain: string
  email_mailgun_api_key_set: boolean
  email_smtp_host: string
  email_smtp_port: number
  email_smtp_username: string
  email_smtp_password_set: boolean
  email_smtp_tls: boolean
  github_oauth_client_id: string
  github_oauth_client_secret_set: boolean
  github_oauth_configured: boolean
}

/** Write-only secret fields accepted by PATCH /admin/settings. */
export type AdminSettingsPatch = Partial<AdminSettings> & {
  email_mailgun_api_key?: string
  email_smtp_password?: string
  github_oauth_client_secret?: string
}

export type AdminUser = {
  id: number
  email: string
  user_name: string
  is_banned: boolean
}

export type APIKey = {
  id: number
  name: string
  key_prefix: string
  created_at: string
  last_used_at?: string | null
}

export type DeviceStatus = {
  user_code: string
  client_name: string
  status: string
  redirect_uri?: string
}

export type DeviceDecisionResult = {
  ok: boolean
  status: string
  redirect_uri?: string
}

export type APIErrorBody = {
  error: string
  message: string
}

export class APIError extends Error {
  code: string
  status: number

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
  }
}
