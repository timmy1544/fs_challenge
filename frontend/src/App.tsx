import { useState } from 'react';
import { useTasks, useCreateTask, useUpdateTask, useDeleteTask } from './hooks/useTasks';
import { useWebSocket } from './hooks/useWebSocket';
import { Task, WebSocketMessage } from './types';
import { useQueryClient } from '@tanstack/react-query';

// Temporary: In production, get from auth context
const WORKSPACE_ID = '00000000-0000-0000-0000-000000000001';
const USER_ID = '00000000-0000-0000-0000-000000000001';

function App() {
  const [newTaskTitle, setNewTaskTitle] = useState('');
  const queryClient = useQueryClient();

  const { data: tasks = [], isLoading } = useTasks(WORKSPACE_ID);
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();

  // Handle WebSocket messages
  const handleWebSocketMessage = (message: WebSocketMessage) => {
    console.log('WebSocket message received:', message);

    switch (message.type) {
      case 'TASK_CREATED':
      case 'TASK_UPDATED':
        if (message.data) {
          queryClient.setQueryData(['task', message.data.id], message.data);
          queryClient.invalidateQueries({ queryKey: ['tasks', WORKSPACE_ID] });
        }
        break;
      case 'TASK_DELETED':
        if (message.task_id) {
          queryClient.removeQueries({ queryKey: ['task', message.task_id] });
          queryClient.invalidateQueries({ queryKey: ['tasks', WORKSPACE_ID] });
        }
        break;
    }
  };

  const { isConnected } = useWebSocket({
    workspaceId: WORKSPACE_ID,
    userId: USER_ID,
    onMessage: handleWebSocketMessage,
  });

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTaskTitle.trim()) return;

    try {
      await createTask.mutateAsync({
        title: newTaskTitle,
        workspace_id: WORKSPACE_ID,
        status: 'todo',
      });
      setNewTaskTitle('');
    } catch (error) {
      console.error('Failed to create task:', error);
    }
  };

  const handleToggleStatus = async (task: Task) => {
    const nextStatus =
      task.status === 'todo'
        ? 'in_progress'
        : task.status === 'in_progress'
        ? 'done'
        : 'todo';

    try {
      await updateTask.mutateAsync({
        id: task.id,
        status: nextStatus,
        version: task.version,
      });
    } catch (error) {
      console.error('Failed to update task:', error);
      alert('Task was modified by another user. Please refresh and try again.');
    }
  };

  const handleDeleteTask = async (id: string) => {
    if (!confirm('Are you sure you want to delete this task?')) return;

    try {
      await deleteTask.mutateAsync(id);
    } catch (error) {
      console.error('Failed to delete task:', error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between mb-6">
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

          <form onSubmit={handleCreateTask} className="mb-6">
            <div className="flex gap-2">
              <input
                type="text"
                value={newTaskTitle}
                onChange={(e) => setNewTaskTitle(e.target.value)}
                placeholder="Add a new task..."
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                type="submit"
                disabled={createTask.isPending}
                className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {createTask.isPending ? 'Adding...' : 'Add Task'}
              </button>
            </div>
          </form>

          {isLoading ? (
            <div className="text-center py-8 text-gray-500">Loading tasks...</div>
          ) : tasks.length === 0 ? (
            <div className="text-center py-8 text-gray-500">No tasks yet. Create one above!</div>
          ) : (
            <div className="space-y-2">
              {tasks.map((task) => (
                <div
                  key={task.id}
                  className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                >
                  <div className="flex-1">
                    <h3 className="font-semibold text-gray-800">{task.title}</h3>
                    {task.description && (
                      <p className="text-sm text-gray-600 mt-1">{task.description}</p>
                    )}
                    <div className="flex items-center gap-4 mt-2">
                      <span
                        className={`px-2 py-1 text-xs rounded ${
                          task.status === 'done'
                            ? 'bg-green-100 text-green-800'
                            : task.status === 'in_progress'
                            ? 'bg-yellow-100 text-yellow-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {task.status.replace('_', ' ')}
                      </span>
                      <span className="text-xs text-gray-500">
                        v{task.version}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleToggleStatus(task)}
                      disabled={updateTask.isPending}
                      className="px-4 py-2 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                    >
                      Toggle
                    </button>
                    <button
                      onClick={() => handleDeleteTask(task.id)}
                      disabled={deleteTask.isPending}
                      className="px-4 py-2 text-sm bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
