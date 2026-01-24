import { apiClient } from './client';
import { Task, CreateTaskRequest, UpdateTaskRequest, TaskDependency } from '../types';

export const taskApi = {
  getTasks: async (projectId: string): Promise<Task[]> => {
    const response = await apiClient.get<Task[]>('/tasks', {
      params: { project_id: projectId },
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

  getDependencies: async (taskId: string): Promise<TaskDependency[]> => {
    const response = await apiClient.get<TaskDependency[]>(`/tasks/${taskId}/dependencies`);
    return response.data;
  },

  addDependency: async (taskId: string, dependsOnId: string): Promise<TaskDependency> => {
    const response = await apiClient.post<TaskDependency>(`/tasks/${taskId}/dependencies`, {
      depends_on_id: dependsOnId,
    });
    return response.data;
  },

  removeDependency: async (taskId: string, dependsOnId: string): Promise<void> => {
    await apiClient.delete(`/tasks/${taskId}/dependencies/${dependsOnId}`);
  },
};
