export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface CursorPage {
  cursor: string
  next_cursor: string
  limit: number
  has_more: boolean
}

export interface ApiListResponse<T> {
  code: number
  message: string
  data: T[]
  page: CursorPage
}

export class ApiError extends Error {
  constructor(
    public readonly code: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

