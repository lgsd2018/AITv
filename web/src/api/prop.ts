import request from '../utils/request'
import type { Prop, CreatePropRequest, UpdatePropRequest } from '../types/prop'

export const propAPI = {
    list(dramaId: string | number) {
        return request.get<Prop[]>('/dramas/' + dramaId + '/props')
    },
    listByEpisode(episodeId: string | number) {
        return request.get<Prop[]>(`/episodes/${episodeId}/props`)
    },
    create(data: CreatePropRequest) {
        return request.post<Prop>('/props', data)
    },
    update(id: number, data: UpdatePropRequest) {
        return request.put<Prop>('/props/' + id, data)
    },
    delete(id: number) {
        return request.delete<void>('/props/' + id)
    },
    extractFromScript(episodeId: number) {
        return request.post<{ task_id: string }>(`/episodes/${episodeId}/props/extract`)
    },
    generateImage(id: number) {
        return request.post<{ task_id: string }>(`/props/${id}/generate`)
    },
    associateWithStoryboard(storyboardId: number, propIds: number[]) {
        return request.post<void>(`/storyboards/${storyboardId}/props`, { prop_ids: propIds })
    },
    listLibrary(userId: number) {
        return request.get('/props/library', { params: { user_id: userId } })
    },
    addToLibrary(propId: number, userId: number, permission?: string) {
        return request.post('/props/library', { prop_id: propId, user_id: userId, permission })
    },
    updateLibrary(id: number, permission: string) {
        return request.put(`/props/library/${id}`, { permission })
    },
    deleteLibrary(id: number) {
        return request.delete(`/props/library/${id}`)
    }
}
