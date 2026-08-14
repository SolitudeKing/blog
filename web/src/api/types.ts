export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

// ListPage 是增长型列表的分页信息，与后端 pagination.ListPage 对应。
// count 是当前页条目数（= len(data)），不是 total。
export interface ListPage {
  page: number
  page_size: number
  count: number
  has_more: boolean
}

export interface ApiListResponse<T> {
  code: number
  message: string
  data: T[]
  page: number
  page_size: number
  count: number
  has_more: boolean
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

