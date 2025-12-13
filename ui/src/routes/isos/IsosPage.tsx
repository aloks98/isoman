import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { IsoList } from '@/components/IsoList';
import { useWebSocket } from '@/hooks/useWebSocket';
import { createISO, deleteISO, listISOs, retryISO } from '@/lib/api';
import { useAppStore } from '@/stores';
import type { CreateISORequest, ISO, WSProgressMessage } from '@/types/iso';

export function IsosPage() {
  const queryClient = useQueryClient();

  // Get UI state from Zustand
  const viewMode = useAppStore((state) => state.viewMode);
  const setViewMode = useAppStore((state) => state.setViewMode);

  // Fetch ISOs with React Query
  const {
    data: isos = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['isos'],
    queryFn: listISOs,
  });

  // Handle WebSocket progress updates
  const handleWebSocketMessage = useCallback(
    (message: WSProgressMessage) => {
      if (message.type === 'progress') {
        queryClient.setQueryData(['isos'], (oldData: ISO[] | undefined) => {
          if (!oldData) return oldData;

          const updatedIsos = oldData.map((iso) =>
            iso.id === message.payload.id
              ? {
                  ...iso,
                  progress: message.payload.progress,
                  status: message.payload.status,
                }
              : iso,
          );

          // If status is complete or failed, refetch to get updated fields (error_message, checksum, etc.)
          if (
            message.payload.status === 'complete' ||
            message.payload.status === 'failed'
          ) {
            queryClient.invalidateQueries({ queryKey: ['isos'] });
          }

          return updatedIsos;
        });
      }
    },
    [queryClient],
  );

  // Set up WebSocket connection
  useWebSocket({ onMessage: handleWebSocketMessage });

  // Create mutation
  const createMutation = useMutation({
    mutationFn: createISO,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['isos'] });
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: deleteISO,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['isos'] });
    },
  });

  // Retry mutation
  const retryMutation = useMutation({
    mutationFn: retryISO,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['isos'] });
    },
  });

  const handleCreate = async (request: CreateISORequest) => {
    await createMutation.mutateAsync(request);
  };

  const handleDelete = (id: string) => {
    deleteMutation.mutate(id);
  };

  const handleRetry = (id: string) => {
    retryMutation.mutate(id);
  };

  return (
    <IsoList
      isos={isos}
      isLoading={isLoading}
      error={error as Error | null}
      viewMode={viewMode}
      onViewModeChange={setViewMode}
      onCreateISO={handleCreate}
      onDeleteISO={handleDelete}
      onRetryISO={handleRetry}
    />
  );
}
