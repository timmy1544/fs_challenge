import { useState } from 'react';
import { useProjects, useCreateProject } from './hooks/useProjects';
import { useTasks, useTask, useCreateTask, useUpdateTask, useDeleteTask } from './hooks/useTasks';
import { useComments, useCreateComment } from './hooks/useComments';
import { useWebSocket } from './hooks/useWebSocket';
import { WebSocketMessage, Task, Project } from './types';
import { useQueryClient } from '@tanstack/react-query';

// Temporary: In production, get from auth context
const USER_ID = '00000000-0000-0000-0000-000000000001';

function App() {
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [newTaskTitle, setNewTaskTitle] = useState('');
  const [newProjectName, setNewProjectName] = useState('');
  const [newComment, setNewComment] = useState('');
  const queryClient = useQueryClient();

  const { data: projects = [], isLoading: projectsLoading } = useProjects();
  const { data: tasks = [], isLoading: tasksLoading } = useTasks(selectedProjectId || '');
  const { data: selectedTask } = useTask(selectedTaskId || '');
  const { data: comments = [] } = useComments(selectedTaskId || '');

  const createProject = useCreateProject();
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();
  const createComment = useCreateComment();

  // Handle WebSocket messages with delta updates
  const handleWebSocketMessage = (message: WebSocketMessage) => {
    console.log('WebSocket message received:', message);

    switch (message.type) {
      case 'TASK_CREATED':
        if (message.full_data) {
          queryClient.setQueryData(['task', message.full_data.id], message.full_data);
          queryClient.invalidateQueries({ queryKey: ['tasks', message.project_id] });
        }
        break;
      case 'TASK_UPDATED':
        if (message.task_id && message.delta) {
          // Apply delta update
          queryClient.setQueryData(['task', message.task_id], (old: Task | undefined) => {
            if (!old) return old;
            return { ...old, ...message.delta };
          });
          queryClient.invalidateQueries({ queryKey: ['tasks', message.project_id] });
        }
        break;
      case 'TASK_DELETED':
        if (message.task_id) {
          queryClient.removeQueries({ queryKey: ['task', message.task_id] });
          queryClient.invalidateQueries({ queryKey: ['tasks', message.project_id] });
        }
        break;
      case 'COMMENT_CREATED':
        if (message.full_data && message.task_id) {
          queryClient.invalidateQueries({ queryKey: ['comments', message.task_id] });
        }
        break;
      case 'COMMENT_UPDATED':
        if (message.comment_id && message.task_id && message.delta) {
          queryClient.invalidateQueries({ queryKey: ['comments', message.task_id] });
        }
        break;
      case 'COMMENT_DELETED':
        if (message.task_id) {
          queryClient.invalidateQueries({ queryKey: ['comments', message.task_id] });
        }
        break;
      case 'PROJECT_CREATED':
        queryClient.invalidateQueries({ queryKey: ['projects'] });
        break;
      case 'PROJECT_UPDATED':
        if (message.delta) {
          queryClient.setQueryData(['project', message.project_id], (old: Project | undefined) => {
            if (!old) return old;
            return { ...old, ...message.delta };
          });
          queryClient.invalidateQueries({ queryKey: ['projects'] });
        }
        break;
      case 'TASK_DEPENDENCY_ADDED':
      case 'TASK_DEPENDENCY_REMOVED':
        if (message.task_id) {
          queryClient.invalidateQueries({ queryKey: ['task', message.task_id] });
        }
        break;
    }
  };

  const { isConnected } = useWebSocket({
    projectId: selectedProjectId || '',
    userId: USER_ID,
    onMessage: handleWebSocketMessage,
  });

  const handleCreateProject = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newProjectName.trim()) return;

    try {
      const project = await createProject.mutateAsync({ name: newProjectName });
      setNewProjectName('');
      setSelectedProjectId(project.id);
    } catch (error) {
      console.error('Failed to create project:', error);
    }
  };

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTaskTitle.trim() || !selectedProjectId) return;

    try {
      await createTask.mutateAsync({
        title: newTaskTitle,
        project_id: selectedProjectId,
        status: 'todo',
      });
      setNewTaskTitle('');
    } catch (error) {
      console.error('Failed to create task:', error);
    }
  };

  const handleToggleStatus = async (task: Task) => {
    const statusMap: Record<string, string> = {
      todo: 'in_progress',
      in_progress: 'done',
      done: 'todo',
      blocked: 'todo',
    };

    const nextStatus = statusMap[task.status] || 'todo';

    try {
      await updateTask.mutateAsync({
        id: task.id,
        status: nextStatus,
        version: task.version,
      });
    } catch (error: any) {
      console.error('Failed to update task:', error);
      if (error.response?.status === 400) {
        alert(error.response.data.error || 'Invalid status transition or dependency not met');
      } else {
        alert('Task was modified by another user. Please refresh and try again.');
      }
    }
  };

  const handleDeleteTask = async (id: string) => {
    if (!confirm('Are you sure you want to delete this task?')) return;

    try {
      await deleteTask.mutateAsync(id);
      if (selectedTaskId === id) {
        setSelectedTaskId(null);
      }
    } catch (error) {
      console.error('Failed to delete task:', error);
    }
  };

  const handleCreateComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newComment.trim() || !selectedTaskId) return;

    try {
      await createComment.mutateAsync({
        taskId: selectedTaskId,
        content: newComment,
      });
      setNewComment('');
    } catch (error) {
      console.error('Failed to create comment:', error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <div className="flex items-center justify-between">
            <h1 className="text-3xl font-bold text-gray-800">Task Manager</h1>
            <div className="flex items-center gap-2">
              <div
                className={`w-3 h-3 rounded-full ${
                  isConnected ? 'bg-green-500' : 'bg-red-500'
                }`}
              />
              <span className="text-sm text-gray-600">
                {isConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Sidebar - Projects */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4">Projects</h2>
            <form onSubmit={handleCreateProject} className="mb-4">
              <div className="flex gap-2">
                <input
                  type="text"
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  placeholder="New project name..."
                  className="flex-1 px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  type="submit"
                  disabled={createProject.isPending}
                  className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                >
                  Add
                </button>
              </div>
            </form>
            {projectsLoading ? (
              <div className="text-sm text-gray-500">Loading projects...</div>
            ) : (
              <div className="space-y-2">
                {projects.map((project) => (
                  <button
                    key={project.id}
                    onClick={() => {
                      setSelectedProjectId(project.id);
                      setSelectedTaskId(null);
                    }}
                    className={`w-full text-left px-3 py-2 rounded ${
                      selectedProjectId === project.id
                        ? 'bg-blue-100 text-blue-800'
                        : 'hover:bg-gray-100'
                    }`}
                  >
                    {project.name}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Middle - Tasks */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4">
              Tasks {selectedProjectId && `(${tasks.length})`}
            </h2>
            {selectedProjectId ? (
              <>
                <form onSubmit={handleCreateTask} className="mb-4">
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={newTaskTitle}
                      onChange={(e) => setNewTaskTitle(e.target.value)}
                      placeholder="Add a new task..."
                      className="flex-1 px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <button
                      type="submit"
                      disabled={createTask.isPending}
                      className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                    >
                      Add
                    </button>
                  </div>
                </form>
                {tasksLoading ? (
                  <div className="text-sm text-gray-500">Loading tasks...</div>
                ) : tasks.length === 0 ? (
                  <div className="text-sm text-gray-500">No tasks yet. Create one above!</div>
                ) : (
                  <div className="space-y-2 max-h-96 overflow-y-auto">
                    {tasks.map((task) => (
                      <div
                        key={task.id}
                        className={`p-3 rounded border ${
                          selectedTaskId === task.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200'
                        } cursor-pointer hover:bg-gray-50`}
                        onClick={() => setSelectedTaskId(task.id)}
                      >
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <h3 className="font-semibold text-gray-800">{task.title}</h3>
                            {task.description && (
                              <p className="text-sm text-gray-600 mt-1">{task.description}</p>
                            )}
                            <div className="flex items-center gap-2 mt-2">
                              <span
                                className={`px-2 py-1 text-xs rounded ${
                                  task.status === 'done'
                                    ? 'bg-green-100 text-green-800'
                                    : task.status === 'in_progress'
                                    ? 'bg-yellow-100 text-yellow-800'
                                    : task.status === 'blocked'
                                    ? 'bg-red-100 text-red-800'
                                    : 'bg-gray-100 text-gray-800'
                                }`}
                              >
                                {task.status.replace('_', ' ')}
                              </span>
                              {task.dependencies && task.dependencies.length > 0 && (
                                <span className="text-xs text-gray-500">
                                  {task.dependencies.length} dependency
                                  {task.dependencies.length > 1 ? 'ies' : ''}
                                </span>
                              )}
                            </div>
                          </div>
                          <div className="flex gap-1 ml-2">
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleToggleStatus(task);
                              }}
                              className="px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600"
                            >
                              Toggle
                            </button>
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteTask(task.id);
                              }}
                              className="px-2 py-1 text-xs bg-red-500 text-white rounded hover:bg-red-600"
                            >
                              Delete
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </>
            ) : (
              <div className="text-sm text-gray-500">Select a project to view tasks</div>
            )}
          </div>

          {/* Right Sidebar - Task Details & Comments */}
          <div className="bg-white rounded-lg shadow-md p-6">
            {selectedTaskId && selectedTask ? (
              <>
                <h2 className="text-xl font-semibold mb-4">Task Details</h2>
                <div className="mb-4">
                  <h3 className="font-bold text-lg">{selectedTask.title}</h3>
                  {selectedTask.description && (
                    <p className="text-sm text-gray-600 mt-2">{selectedTask.description}</p>
                  )}
                  <div className="mt-2">
                    <span
                      className={`px-2 py-1 text-xs rounded ${
                        selectedTask.status === 'done'
                          ? 'bg-green-100 text-green-800'
                          : selectedTask.status === 'in_progress'
                          ? 'bg-yellow-100 text-yellow-800'
                          : selectedTask.status === 'blocked'
                          ? 'bg-red-100 text-red-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}
                    >
                      {selectedTask.status.replace('_', ' ')}
                    </span>
                  </div>
                </div>

                {/* Dependencies */}
                {selectedTask.dependencies && selectedTask.dependencies.length > 0 && (
                  <div className="mb-4">
                    <h4 className="font-semibold text-sm mb-2">Dependencies</h4>
                    <ul className="text-sm space-y-1">
                      {selectedTask.dependencies.map((dep) => (
                        <li key={dep.id} className="text-gray-600">
                          • {dep.depends_on?.title || dep.depends_on_id}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {/* Comments */}
                <div className="border-t pt-4">
                  <h4 className="font-semibold text-sm mb-2">Comments</h4>
                  <form onSubmit={handleCreateComment} className="mb-4">
                    <textarea
                      value={newComment}
                      onChange={(e) => setNewComment(e.target.value)}
                      placeholder="Add a comment..."
                      className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                      rows={2}
                    />
                    <button
                      type="submit"
                      disabled={createComment.isPending}
                      className="mt-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                    >
                      Add Comment
                    </button>
                  </form>
                  <div className="space-y-2 max-h-64 overflow-y-auto">
                    {comments.map((comment) => (
                      <div key={comment.id} className="p-2 bg-gray-50 rounded text-sm">
                        <p className="text-gray-800">{comment.content}</p>
                        <p className="text-xs text-gray-500 mt-1">
                          {new Date(comment.created_at).toLocaleString()}
                        </p>
                        {comment.replies && comment.replies.length > 0 && (
                          <div className="ml-4 mt-2 space-y-1">
                            {comment.replies.map((reply) => (
                              <div key={reply.id} className="p-2 bg-white rounded">
                                <p className="text-gray-800">{reply.content}</p>
                                <p className="text-xs text-gray-500 mt-1">
                                  {new Date(reply.created_at).toLocaleString()}
                                </p>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              </>
            ) : (
              <div className="text-sm text-gray-500">Select a task to view details</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
