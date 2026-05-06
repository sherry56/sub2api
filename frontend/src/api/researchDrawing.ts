import { apiClient } from './client'

export interface ResearchDrawingGenerateRequest {
  method_content: string
  caption?: string
  optimize_method_content?: boolean
  generation_mode?: string
  exp_mode?: string
  retrieval_setting?: string
  num_candidates?: number
  aspect_ratio?: string
  max_critic_rounds?: number
  max_refine_resolution?: string
  main_model_name?: string
  image_gen_model_name?: string
}

export interface ResearchDrawingGenerateResponse {
  job_id: string
  status: string
  mode?: string
  paperbanana_url?: string
  paperbanana_user?: string
  charge: number
  quota_need: number
}

export interface ResearchDrawingJobStatus {
  ok?: boolean
  job_id?: string
  username?: string
  status: 'running' | 'done' | 'error' | 'unknown'
  mode?: string
  elapsed_sec?: number
  candidate_count?: number
  candidate_ids?: number[]
  images?: Array<{
    candidate_id: number
    url?: string
  }>
  error?: string
}

export const researchDrawingAPI = {
  async generate(payload: ResearchDrawingGenerateRequest): Promise<ResearchDrawingGenerateResponse> {
    const response = await apiClient.post('/research-drawing/generate', payload)
    return response.data
  },

  async getJobStatus(jobId: string, paperBananaUser?: string): Promise<ResearchDrawingJobStatus> {
    const response = await apiClient.get(`/research-drawing/jobs/${encodeURIComponent(jobId)}`, {
      params: paperBananaUser ? { paperbanana_user: paperBananaUser } : undefined,
    })
    return response.data
  },

  async getJobImage(jobId: string, candidateId: number, paperBananaUser?: string): Promise<Blob> {
    const response = await apiClient.get(
      `/research-drawing/jobs/${encodeURIComponent(jobId)}/images/${encodeURIComponent(String(candidateId))}`,
      {
        params: paperBananaUser ? { paperbanana_user: paperBananaUser } : undefined,
        responseType: 'blob',
      },
    )
    return response.data
  },
}
