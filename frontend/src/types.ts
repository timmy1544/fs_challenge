export interface Project {
  id: string;
  name: string;
  description: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  title: string;
  description: string;
  status: 'todo' | 'in_progress' | 'done' | 'blocked';
  assignee_id?: string;
  project_id: string;
  created_by: string;
  version: number;
  created_at: string;
  updated_at: string;
  dependencies?: TaskDependency[];
  comments?: Comment[];
}

export interface TaskDependency {
  id: string;
  task_id: string;
  depends_on_id: string;
  depends_on?: Task;
  created_at: string;
}

export interface Comment {
  id: string;
  task_id: string;
  user_id: string;
  content: string;
  parent_id?: string;
  created_at: string;
  updated_at: string;
  replies?: Comment[];
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
}

export interface CreateTaskRequest {
  title: string;
  description?: string;
  status?: string;
  project_id: string;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: string;
  assignee_id?: string;
  version: number;
}

export interface CreateCommentRequest {
  content: string;
  parent_id?: string;
}

export interface UpdateCommentRequest {
  content: string;
}

export interface WebSocketMessage {
  type: 
    | 'TASK_CREATED' 
    | 'TASK_UPDATED' 
    | 'TASK_DELETED'
    | 'TASK_DEPENDENCY_ADDED'
    | 'TASK_DEPENDENCY_REMOVED'
    | 'COMMENT_CREATED'
    | 'COMMENT_UPDATED'
    | 'COMMENT_DELETED'
    | 'PROJECT_CREATED'
    | 'PROJECT_UPDATED'
    | 'PROJECT_DELETED';
  project_id: string;
  task_id?: string;
  comment_id?: string;
  delta?: Record<string, any>;
  full_data?: any;
}
