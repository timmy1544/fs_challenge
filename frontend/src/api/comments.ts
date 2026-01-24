import { apiClient } from './client';
import { Comment, CreateCommentRequest, UpdateCommentRequest } from '../types';

export const commentApi = {
  getComments: async (taskId: string): Promise<Comment[]> => {
    const response = await apiClient.get<Comment[]>(`/tasks/${taskId}/comments`);
    return response.data;
  },

  createComment: async (taskId: string, comment: CreateCommentRequest): Promise<Comment> => {
    const response = await apiClient.post<Comment>(`/tasks/${taskId}/comments`, comment);
    return response.data;
  },

  updateComment: async (commentId: string, comment: UpdateCommentRequest): Promise<Comment> => {
    const response = await apiClient.put<Comment>(`/comments/${commentId}`, comment);
    return response.data;
  },

  deleteComment: async (commentId: string): Promise<void> => {
    await apiClient.delete(`/comments/${commentId}`);
  },
};
