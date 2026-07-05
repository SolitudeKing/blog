export interface AssetItem {
  id: number
  display_name: string
  alt_text: string
  storage_key: string
  url: string
  thumb_url: string
  mime_type: string
  ext: string
  size: number
  width: number
  height: number
  sha256: string
  status: 'ready' | 'processing' | 'failed'
  ref_count: number
  created_at: string
  updated_at: string
}

export interface AssetUpdatePayload {
  display_name: string
  alt_text: string
}

export interface AssetReferenceItem {
  id: number
  type: string
  title: string
  url: string
}
