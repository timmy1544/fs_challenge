import { apiClient } from './client';
import { Task, CreateTaskRequest, UpdateTaskRequest } from '../types';

export const taskApi = {
  getTasks: async (workspaceId: string): Promise<Task[]> => {
    const response = await apiClient.get<Task[]>('/tasks', {
      params: { workspace_id: workspaceId },
    });
    return response.data;
  },

  getTask: async (id: string): Promise<Task> => {
    const response = await apiClient.get<Task>(`/tasks/${id}`);
    return response.data;
  },

  createTask: async (task: CreateTaskRequest): Promise<Task> => {
    const response = await apiClient.post<Task>('/tasks', task);
    return response.data;
  },

  updateTask: async (id: string, task: UpdateTaskRequest): Promise<Task> => {
    const response = await apiClient.put<Task>(`/tasks/${id}`, task);
    return response.data;
  },

  deleteTask: async (id: string): Promise<void> => {
    await apiClient.delete(`/tasks/${id}`);
  },
};
