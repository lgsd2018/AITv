export interface Prop {
    id: number
    drama_id: number
    name: string
    type?: string
    description?: string
    prompt?: string
    image_url?: string
    reference_images?: any
    attributes?: Record<string, any>
    created_by?: number
    characters?: Array<{ id: number; name: string }>
    scenes?: Array<{ id: number; location?: string; time?: string }>
    created_at: string
    updated_at: string
}

export interface CreatePropRequest {
    drama_id: number
    name: string
    type?: string
    description?: string
    prompt?: string
    image_url?: string
    attributes?: Record<string, any>
    character_ids?: number[]
    scene_ids?: number[]
    created_by?: number
}

export interface UpdatePropRequest {
    name?: string
    type?: string
    description?: string
    prompt?: string
    image_url?: string
    attributes?: Record<string, any>
    character_ids?: number[]
    scene_ids?: number[]
    created_by?: number
}
