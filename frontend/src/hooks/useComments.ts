import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { commentApi } from '../api/comments';
import { Comment, CreateCommentRequest, UpdateCommentRequest } from '../types';

export function useComments(taskId: string) {
  return useQuery({
    queryKey: ['comments', taskId],
    queryFn: () => commentApi.getComments(taskId),
    enabled: !!taskId,
  });
}

export function useCreateComment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId, ...comment }: { taskId: string } & CreateCommentRequest) =>
      commentApi.createComment(taskId, comment),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['comments', variables.taskId] });
    },
  });
}

export function useUpdateComment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ commentId, ...comment }: { commentId: string } & UpdateCommentRequest) =>
      commentApi.updateComment(commentId, comment),
    onSuccess: (updatedComment) => {
      queryClient.invalidateQueries({ queryKey: ['comments', updatedComment.task_id] });
    },
  });
}

export function useDeleteComment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (commentId: string) => commentApi.deleteComment(commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments'] });
    },
  });
}
