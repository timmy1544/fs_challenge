export interface Task {
  id: string;
  title: string;
  description: string;
  status: 'todo' | 'in_progress' | 'done';
  assignee_id?: string;
  workspace_id: string;
  created_by: string;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface CreateTaskRequest {
  title: string;
  description?: string;
  status?: string;
  workspace_id: string;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: string;
  assignee_id?: string;
  version: number;
}

export interface WebSocketMessage {
  type: 'TASK_CREATED' | 'TASK_UPDATED' | 'TASK_DELETED';
  workspace_id: string;
  task_id?: string;
  data: Task | null;
}
